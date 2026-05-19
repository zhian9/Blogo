// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gomail "github.com/wneessen/go-mail"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/rbac/dal"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/cachex"
	"github.com/zhian9/blogo-server/pkg/captcha"
	"github.com/zhian9/blogo-server/pkg/crypto/hash"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/jwtx"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
)

// Login 是登录业务的核心对象，聚合了认证、缓存、数据访问等依赖。
type Login struct {
	Cache           cachex.Cacher     // 缓存客户端（存储用户角色等）
	Auth            jwtx.Auther       // JWT 认证器
	UserDAL         *dal.User         // 用户数据访问层
	UserRoleDAL     *dal.UserRole     // 用户角色数据访问层
	MenuDAL         *dal.Menu         // 菜单数据访问层
	UserBIZ         *User             // 用户业务对象（用于获取角色）
	CaptchaSvc      *captcha.Service  // 验证码
	OperationLogDAL *dal.OperationLog // 审计日志数据访问层
}

// ParseUserID 从请求中解析用户 ID（供 Auth 中间件使用）。
// 流程：
//  1. 检查认证是否禁用（返回 root ID）
//  2. 从 Token 解析用户 ID
//  3. 检查超级管理员
//  4. 查询缓存（用户角色）
//  5. 缓存未命中 → 查询数据库 + 写入缓存
func (l *Login) ParseUserID(c *gin.Context) (string, error) {
	rootID := config.C.General.Root.ID

	// 1. 认证禁用时返回 root ID（开发模式）
	if config.C.Middleware.Auth.Disable {
		return rootID, nil
	}

	// 2. 定义无效 Token 错误
	invalidToken := errors.Unauthorized(config.ErrInvalidTokenID, "Invalid access token")

	// 3. 获取 Token
	token := util.GetToken(c)
	if token == "" {
		return "", invalidToken
	}

	// 4. 注入 Token 到上下文
	ctx := c.Request.Context()
	ctx = util.NewUserToken(ctx, token)

	// 5. 解析 Token 获取用户 ID
	userID, err := l.Auth.ParseSubject(ctx, token)
	if err != nil {
		if err == jwtx.ErrInvalidToken {
			return "", invalidToken
		}
		return "", err
	}

	// 6. 超级管理员特殊处理
	if userID == rootID {
		c.Request = c.Request.WithContext(util.NewIsRootUser(ctx))
		return userID, nil
	}

	// 7. 查询缓存（用户角色）
	userCacheVal, ok, err := l.Cache.Get(ctx, config.CacheNSForUser, userID)
	if err != nil {
		return "", err
	} else if ok {
		// 缓存命中 → 注入用户缓存到上下文
		userCache := util.ParseUserCache(userCacheVal)
		c.Request = c.Request.WithContext(util.NewUserCache(ctx, userCache))
		return userID, nil
	}

	// 8. 缓存未命中 → 查询数据库
	// 8.1 检查用户状态
	user, err := l.UserDAL.Get(ctx, userID, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{SelectFields: []string{"status"}},
	})
	if err != nil {
		return "", err
	} else if user == nil {
		return "", invalidToken
	} else if user.Status != schema.UserStatusActivated {
		return "", errors.Unauthorized("", "您的账号已被系统管理员禁用，请联系运维人员") // 用户不存在或未激活
	}

	// 8.2 获取用户角色
	roleIDs, err := l.UserBIZ.GetRoleIDs(ctx, userID)
	if err != nil {
		return "", err
	}

	// 8.3 写入缓存（有效期由配置决定）
	userCache := util.UserCache{RoleIDs: roleIDs}
	err = l.Cache.Set(ctx, config.CacheNSForUser, userID, userCache.String(),
		time.Duration(config.C.Dictionary.UserCacheExp)*time.Hour)
	if err != nil {
		return "", err
	}

	// 8.4 注入用户缓存到上下文
	c.Request = c.Request.WithContext(util.NewUserCache(ctx, userCache))
	return userID, nil
}

// GetCaptcha 只返回 captcha_id
func (l *Login) GetCaptcha(ctx context.Context) (*schema.Captcha, error) {
	id, err := l.CaptchaSvc.Generate(ctx)
	if err != nil {
		return nil, err
	}
	return &schema.Captcha{CaptchaID: id}, nil
}

// ResponseCaptcha 根据 ID 返回图片
func (l *Login) ResponseCaptcha(ctx context.Context, w http.ResponseWriter, id string, reload bool) error {
	if id == "" {
		return errors.New(errors.DefaultBadRequestID, "captcha id is required", 23)
	}

	b64s, err := l.CaptchaSvc.GetImage(ctx, id)
	if err != nil {
		return err
	}

	// 解码 base64
	imgData := strings.Replace(b64s, "data:image/png;base64,", "", 1)
	pngBytes, err := base64.StdEncoding.DecodeString(imgData)
	if err != nil {
		return err
	}

	// 设置响应头
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "image/png")

	_, _ = w.Write(pngBytes)
	return nil
}

// genUserToken 生成用户登录 Token（JWT），包含角色编码。
func (l *Login) genUserToken(ctx context.Context, userID string) (*schema.LoginToken, error) {
	// 1. 获取用户角色编码
	var roleCode string = "guest" // 默认游客
	if userID == config.C.General.Root.ID {
		roleCode = "admin" // 超级管理员视为 admin
	} else {
		roleIDs, err := l.UserBIZ.GetRoleIDs(ctx, userID)
		if err != nil {
			return nil, err
		}
		if len(roleIDs) > 0 {
			roleCode = roleIDs[0] // 假设 roleID 就是 roleCode
		}
	}

	// 2. 生成包含角色编码的 Token
	token, err := l.Auth.(interface {
		GenerateTokenWithRole(context.Context, string, string) (jwtx.TokenInfo, error)
	}).GenerateTokenWithRole(ctx, userID, roleCode)
	if err != nil {
		return nil, err
	}

	// 3. 记录日志
	logging.Context(ctx).Info("Generate user token", zap.Int64("expires_at", token.GetExpiresAt()))

	return &schema.LoginToken{
		AccessToken: token.GetAccessToken(),
		TokenType:   token.GetTokenType(),
		ExpiresAt:   token.GetExpiresAt(),
	}, nil
}

// Login 处理用户登录请求。
// 流程：
//  1. 验证码校验
//  2. 超级管理员登录
//  3. 普通用户登录（密码校验 + 状态检查）
//  4. 写入用户缓存
//  5. 生成 Token
//  6. 记录审计日志
func (l *Login) Login(ctx context.Context, formItem *schema.LoginForm, clientIP, userAgent, realIP string) (*schema.LoginToken, error) {
	operator := formItem.Username
	if realIP == "" {
		realIP = clientIP
	}

	// 1. 验证码校验
	if !l.CaptchaSvc.Verify(ctx, formItem.CaptchaID, formItem.CaptchaCode) {
		l.writeAuditLog(ctx, operator, realIP, userAgent, "登录", "用户 ["+operator+"] 登录失败：验证码错误", "验证码错误", false, http.StatusBadRequest)
		return nil, errors.BadRequest(config.ErrInvalidCaptchaID, "Incorrect captcha")
	}

	ctx = logging.NewTag(ctx, logging.TagKeyLogin)

	// 2. 超级管理员登录（root 密码使用 bcrypt 哈希验证）
	if formItem.Username == config.C.General.Root.Username {
		if err := hash.CompareHashAndPassword(config.C.General.Root.Password, formItem.Password); err == nil {
			userID := config.C.General.Root.ID
			ctx = logging.NewUserID(ctx, userID)
			logging.Context(ctx).Info("Login by root")
			l.writeAuditLog(ctx, operator, realIP, userAgent, "登录", "用户 ["+operator+"] 成功登录系统", "", true, http.StatusOK)
			return l.genUserToken(ctx, userID)
		}
		l.writeAuditLog(ctx, operator, realIP, userAgent, "登录", "用户 ["+operator+"] 登录失败：密码错误", "密码错误", false, http.StatusBadRequest)
	}

	// 3. 普通用户登录
	// 3.1 查询用户（仅需密码和状态）
	user, err := l.UserDAL.GetByUsername(ctx, formItem.Username, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"id", "password", "status"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		l.writeAuditLog(ctx, operator, realIP, userAgent, "登录", "用户 ["+operator+"] 登录失败：用户名或密码错误", "用户名或密码错误", false, http.StatusBadRequest)
		return nil, errors.BadRequest(config.ErrInvalidUsernameOrPassword, "Incorrect username or password")
	} else if user.Status == schema.UserStatusInactive {
		l.writeAuditLog(ctx, operator, realIP, userAgent, "登录", "用户 ["+operator+"] 登录失败：账号未激活", "账号未激活", false, http.StatusBadRequest)
		return nil, errors.BadRequest("", "账号尚未激活，请查收注册邮箱中的激活邮件并完成验证")
	} else if user.Status != schema.UserStatusActivated {
		l.writeAuditLog(ctx, operator, realIP, userAgent, "登录", "用户 ["+operator+"] 登录失败：账号已禁用", "账号已禁用", false, http.StatusBadRequest)
		return nil, errors.BadRequest("", "您的账号已被系统管理员禁用，请联系运维人员")
	}

	// 3.2 密码校验
	if err := hash.CompareHashAndPassword(user.Password, formItem.Password); err != nil {
		l.writeAuditLog(ctx, operator, realIP, userAgent, "登录", "用户 ["+operator+"] 登录失败：密码错误", "密码错误", false, http.StatusBadRequest)
		return nil, errors.BadRequest(config.ErrInvalidUsernameOrPassword, "Incorrect username or password")
	}

	// 3.3 原子更新最后登录时间和 IP（非关键路径，失败仅记录日志）
	userID := user.ID
	now := time.Now()
	if err := l.UserDAL.UpdateLastLogin(ctx, userID, now, clientIP); err != nil {
		logging.Context(ctx).Error("Failed to update last login", zap.Error(err),
			zap.String("user_id", userID), zap.String("client_ip", clientIP))
	}

	// 3.4 写入用户缓存
	ctx = logging.NewUserID(ctx, userID)
	roleIDs, err := l.UserBIZ.GetRoleIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	userCache := util.UserCache{RoleIDs: roleIDs}
	err = l.Cache.Set(ctx, config.CacheNSForUser, userID, userCache.String(),
		time.Duration(config.C.Dictionary.UserCacheExp)*time.Hour)
	if err != nil {
		logging.Context(ctx).Error("Failed to set cache", zap.Error(err))
	}
	logging.Context(ctx).Info("Login success", zap.String("username", formItem.Username))

	// 3.5 审计日志
	l.writeAuditLog(ctx, operator, realIP, userAgent, "登录", "用户 ["+operator+"] 成功登录系统", "", true, http.StatusOK)

	// 3.6 生成 Token
	return l.genUserToken(ctx, userID)
}

// Register 处理公开注册请求。
// 流程：
//  1. 检查是否允许注册
//  2. 验证码校验
//  3. 用户名/邮箱唯一性校验
//  4. 创建用户（inactive，含激活 token）并绑定 user 角色
//  5. 异步发送激活邮件
func (l *Login) Register(ctx context.Context, formItem *schema.RegisterForm, clientIP, userAgent, realIP string) error {
	if !config.C.General.AllowRegister {
		return errors.BadRequest("", "Register is disabled")
	}

	// 基础校验
	if err := formItem.Validate(); err != nil {
		return err
	}

	// 验证码
	if !l.CaptchaSvc.Verify(ctx, formItem.CaptchaID, formItem.CaptchaCode) {
		return errors.BadRequest(config.ErrInvalidCaptchaID, "Incorrect captcha")
	}

	// 用户名唯一性
	exists, err := l.UserDAL.ExistsUsername(ctx, formItem.Username)
	if err != nil {
		return err
	} else if exists {
		return errors.BadRequest("", "用户名已存在")
	}

	// 邮箱唯一性
	var emailCount int64
	if err := l.UserDAL.DB.Model(&schema.User{}).Where("email = ?", formItem.Email).Count(&emailCount).Error; err != nil {
		return err
	} else if emailCount > 0 {
		return errors.BadRequest("", "该邮箱已被注册，请直接登录或更换其他邮箱")
	}

	// 生成激活 token
	activationToken := uuid.New().String()

	// 构造用户（inactive）
	user := &schema.User{
		ID:              util.NewXID(),
		Username:        formItem.Username,
		Name:            formItem.Username,
		Phone:           formItem.Phone,
		Email:           formItem.Email,
		Status:          schema.UserStatusInactive,
		ActivationToken: activationToken,
		CreatedAt:       time.Now(),
	}
	// 生成密码哈希
	hashPass, err := hash.GeneratePassword(formItem.Password)
	if err != nil {
		return errors.BadRequest("", "Failed to generate hash password: %s", err.Error())
	}
	user.Password = hashPass

	// 创建用户
	if err := l.UserDAL.Create(ctx, user); err != nil {
		return err
	}

	// 注册用户默认绑定 "user" 角色，杜绝通过注册接口提权
	var userRole schema.Role
	if err := l.UserRoleDAL.DB.Where("code = ?", "user").First(&userRole).Error; err == nil {
		_ = l.UserRoleDAL.Create(ctx, &schema.UserRole{
			ID:        util.NewXID(),
			UserID:    user.ID,
			RoleID:    userRole.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	go l.sendActivationEmail(user.Email, user.Username, activationToken)

	if realIP == "" {
		realIP = clientIP
	}
	l.writeAuditLog(ctx, formItem.Username, realIP, userAgent, "注册", "新用户 ["+formItem.Username+"] 注册成功", "", true, http.StatusOK)

	logging.Context(ctx).Info("User registered, activation email sent",
		zap.String("username", formItem.Username),
		zap.String("email", formItem.Email),
	)

	return nil
}

// ActivateAccount 通过激活 token 激活账户。
func (l *Login) ActivateAccount(ctx context.Context, token string) error {
	if token == "" {
		return errors.BadRequest("", "Missing activation token")
	}

	user, err := l.UserDAL.GetByActivationToken(ctx, token)
	if err != nil {
		return err
	} else if user == nil {
		return errors.BadRequest("", "Invalid or expired activation token")
	}

	if user.Status != schema.UserStatusInactive {
		return errors.BadRequest("", "Account is already activated")
	}

	now := time.Now()
	if err := l.UserDAL.ActivateUser(ctx, user.ID, now); err != nil {
		return err
	}

	logging.Context(ctx).Info("Account activated",
		zap.String("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return nil
}

// RefreshToken 刷新用户 Token。
// 用于 Token 即将过期时延长会话。
func (l *Login) RefreshToken(ctx context.Context) (*schema.LoginToken, error) {
	userID := util.FromUserID(ctx)

	// 检查用户状态
	user, err := l.UserDAL.Get(ctx, userID, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"status"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.BadRequest("", "Incorrect user")
	} else if user.Status == schema.UserStatusInactive {
		return nil, errors.BadRequest("", "账号尚未激活，请查收注册邮箱中的激活邮件并完成验证")
	} else if user.Status != schema.UserStatusActivated {
		return nil, errors.BadRequest("", "您的账号已被系统管理员禁用，请联系运维人员")
	}

	return l.genUserToken(ctx, userID)
}

// Logout 处理用户登出请求。
// 流程：
//  1. 使 Token 失效（加入黑名单）
//  2. 删除用户缓存
func (l *Login) Logout(ctx context.Context) error {
	userToken := util.FromUserToken(ctx)
	if userToken == "" {
		return nil
	}

	ctx = logging.NewTag(ctx, logging.TagKeyLogout)

	// 1. 使 Token 失效
	if err := l.Auth.DestroyToken(ctx, userToken); err != nil {
		return err
	}

	// 2. 删除用户缓存
	userID := util.FromUserID(ctx)
	err := l.Cache.Delete(ctx, config.CacheNSForUser, userID)
	if err != nil {
		logging.Context(ctx).Error("Failed to delete user cache", zap.Error(err))
	}
	logging.Context(ctx).Info("Logout success")

	return nil
}

// GetUserInfo 获取当前用户信息（含角色）。
func (l *Login) GetUserInfo(ctx context.Context) (*schema.User, error) {
	// 超级管理员特殊处理
	if util.FromIsRootUser(ctx) {
		return &schema.User{
			ID:       config.C.General.Root.ID,
			Username: config.C.General.Root.Username,
			Name:     config.C.General.Root.Name,
			Status:   schema.UserStatusActivated,
		}, nil
	}

	// 查询普通用户信息（排除密码）
	userID := util.FromUserID(ctx)
	user, err := l.UserDAL.Get(ctx, userID, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			OmitFields: []string{"password"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.NotFound("", "User not found")
	}

	// 查询用户角色（关联角色名称）
	userRoleResult, err := l.UserRoleDAL.Query(ctx, schema.UserRoleQueryParam{
		UserID: userID,
	}, schema.UserRoleQueryOptions{
		JoinRole: true,
	})
	if err != nil {
		return nil, err
	}
	user.Roles = userRoleResult.Data

	return user, nil
}

// UpdatePassword 修改当前用户登录密码。
func (l *Login) UpdatePassword(ctx context.Context, updateItem *schema.UpdateLoginPassword) error {
	// 超级管理员禁止修改密码
	if util.FromIsRootUser(ctx) {
		return errors.BadRequest("", "Root user cannot change password")
	}

	userID := util.FromUserID(ctx)
	// 查询用户密码
	user, err := l.UserDAL.Get(ctx, userID, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"password"},
		},
	})
	if err != nil {
		return err
	} else if user == nil {
		return errors.NotFound("", "User not found")
	}

	// 校验旧密码
	if err := hash.CompareHashAndPassword(user.Password, updateItem.OldPassword); err != nil {
		return errors.BadRequest("", "Incorrect old password")
	}

	// 生成新密码哈希
	newPassword, err := hash.GeneratePassword(updateItem.NewPassword)
	if err != nil {
		return err
	}
	return l.UserDAL.UpdatePasswordByID(ctx, userID, newPassword)
}

// QueryMenus 查询当前用户有权限的菜单树。
// 流程：
//  1. 超级管理员返回所有启用菜单
//  2. 普通用户返回关联菜单 + 祖先菜单
//  3. 构建树形结构
func (l *Login) QueryMenus(ctx context.Context) (schema.Menus, error) {
	menuQueryParams := schema.MenuQueryParam{
		Status: schema.MenuStatusEnabled,
	}

	isRoot := util.FromIsRootUser(ctx)
	if !isRoot {
		menuQueryParams.UserID = util.FromUserID(ctx)
	}

	// 1. 查询用户菜单
	menuResult, err := l.MenuDAL.Query(ctx, menuQueryParams, schema.MenuQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: schema.MenusOrderParams,
		},
	})
	if err != nil {
		return nil, err
	} else if isRoot {
		return menuResult.Data.ToTree(), nil
	}

	// 2. 补全祖先菜单（确保菜单树完整）
	if parentIDs := menuResult.Data.SplitParentIDs(); len(parentIDs) > 0 {
		var missMenusIDs []string
		menuIDMapper := menuResult.Data.ToMap()
		for _, parentID := range parentIDs {
			if _, ok := menuIDMapper[parentID]; !ok {
				missMenusIDs = append(missMenusIDs, parentID)
			}
		}
		if len(missMenusIDs) > 0 {
			// 查询缺失的祖先菜单
			parentResult, err := l.MenuDAL.Query(ctx, schema.MenuQueryParam{
				InIDs: missMenusIDs,
			})
			if err != nil {
				return nil, err
			}
			// 合并菜单列表并排序
			menuResult.Data = append(menuResult.Data, parentResult.Data...)
			sort.Sort(menuResult.Data)
		}
	}

	// 3. 构建树形结构
	return menuResult.Data.ToTree(), nil
}

// UpdateUser 更新当前用户信息。
func (l *Login) UpdateUser(ctx context.Context, updateItem *schema.UpdateCurrentUser) error {
	// 超级管理员禁止更新
	if util.FromIsRootUser(ctx) {
		return errors.BadRequest("", "Root user cannot update")
	}

	userID := util.FromUserID(ctx)
	user, err := l.UserDAL.Get(ctx, userID)
	if err != nil {
		return err
	} else if user == nil {
		return errors.NotFound("", "User not found")
	}

	// 更新字段
	user.Name = updateItem.Name
	user.Phone = updateItem.Phone
	user.Email = updateItem.Email
	user.Avatar = updateItem.Avatar
	user.Bio = updateItem.Bio
	user.Remark = updateItem.Remark
	return l.UserDAL.Update(ctx, user, "name", "phone", "email", "avatar", "bio", "remark")
}

// activationEmailTemplate 激活邮件 HTML 模板
const activationEmailTemplate = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background-color:#0B1311;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif">
<table width="100%" cellpadding="0" cellspacing="0" style="background-color:#0B1311"><tr><td align="center" style="padding:40px 16px">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:560px;background-color:#141D1A;border-radius:16px;border:1px solid rgba(255,255,255,0.06);overflow:hidden">
<tr><td style="padding:32px 32px 0;text-align:center">
<div style="font-family:Georgia,'Times New Roman',serif;font-style:italic;font-size:32px;font-weight:400;color:#FFFFFF;letter-spacing:-0.02em;margin-bottom:4px">Blogo</div>
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:11px;font-weight:500;color:rgba(255,255,255,0.3);letter-spacing:0.16em;text-transform:uppercase">Account Activation</div>
</td></tr>
<tr><td style="padding:0 32px"><div style="height:1px;background:rgba(255,255,255,0.06);margin:24px 0"></div></td></tr>
<tr><td style="padding:28px 32px 8px;text-align:center">
<div style="font-family:Georgia,serif;font-size:20px;font-weight:400;color:#FFFFFF;letter-spacing:-0.01em">激活你的账号 ✦</div>
</td></tr>
<tr><td style="padding:8px 32px 28px;text-align:center">
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:14px;color:rgba(255,255,255,0.55);line-height:1.8;max-width:440px;margin:0 auto">
感谢注册 <span style="font-family:Georgia,serif;font-style:italic;color:#FFFFFF">Blogo</span>。<br>
点击下方按钮完成邮箱验证，激活你的账号。
</div>
</td></tr>
<tr><td style="padding:8px 32px 32px;text-align:center">
<a href="{{.ActivationURL}}" style="display:inline-block;padding:14px 36px;background:linear-gradient(135deg,rgba(79,110,247,0.9),rgba(99,102,241,0.85));color:#FFFFFF;font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:15px;font-weight:600;text-decoration:none;border-radius:12px;box-shadow:0 4px 20px rgba(79,110,247,0.3);letter-spacing:0.03em" target="_blank">激活账号</a>
</td></tr>
<tr><td style="padding:0 32px 28px;text-align:center">
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:12px;color:rgba(255,255,255,0.3);line-height:1.6">
如果按钮无法点击，请复制以下链接在浏览器中打开：<br>
<span style="color:rgba(255,255,255,0.45);font-size:11px">{{.ActivationURL}}</span>
</div>
</td></tr>
<tr><td style="padding:32px;text-align:center">
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:11px;color:rgba(255,255,255,0.22);line-height:1.6">
此邮件发送至 {{.Recipient}}<br>
由 <span style="font-family:Georgia,serif;font-style:italic">Blogo</span> 自动发送 · 如非本人操作请忽略
</div>
</td></tr>
</table>
</td></tr></table>
</body>
</html>`

var parsedActivationTemplate = template.Must(template.New("activation").Parse(activationEmailTemplate))

type activationEmailData struct {
	ActivationURL string
	Recipient     string
}

// sendActivationEmail 异步发送邮箱激活邮件。
func (l *Login) sendActivationEmail(to, username, token string) {
	cfg := config.C.Email
	if cfg.Host == "" || cfg.Password == "" {
		logging.Context(context.Background()).Warn("Activation email not sent: SMTP not configured")
		return
	}

	activationURL := fmt.Sprintf("%s/api/v1/verify-email?token=%s",
		strings.TrimRight(config.C.General.SiteURL, "/"),
		url.QueryEscape(token))

	var buf bytes.Buffer
	if err := parsedActivationTemplate.Execute(&buf, activationEmailData{
		ActivationURL: activationURL,
		Recipient:     to,
	}); err != nil {
		logging.Context(context.Background()).Error("Failed to render activation email template", zap.Error(err))
		return
	}

	m := gomail.NewMsg()
	m.From(cfg.FromEmail)
	m.To(to)
	m.Subject(fmt.Sprintf("欢迎注册 Blogo · %s，请激活你的账号", username))
	m.SetBodyString("text/html", buf.String())

	client, err := gomail.NewClient(cfg.Host,
		gomail.WithPort(cfg.Port), gomail.WithSSL(),
		gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(cfg.FromEmail), gomail.WithPassword(cfg.Password),
		gomail.WithTLSConfig(&tls.Config{ServerName: cfg.Host}),
	)
	if err != nil {
		logging.Context(context.Background()).Error("Failed to create mail client", zap.Error(err))
		return
	}

	if err := client.DialAndSend(m); err != nil {
		logging.Context(context.Background()).Error("Failed to send activation email",
			zap.String("to", to), zap.Error(err))
		return
	}

	logging.Context(context.Background()).Info("Activation email sent",
		zap.String("to", to), zap.String("username", username))
}

// ─── 审计日志（认证模块） ───

const (
	auditModuleAuth = "认证模块"
)

// writeAuditLog 写入认证模块审计日志（支持未登录场景）。
func (l *Login) writeAuditLog(ctx context.Context, operator, operatorIP, userAgent, actionType, description, errorMsg string, success bool, statusCode int) {
	if l.OperationLogDAL == nil {
		return
	}
	entry := &schema.OperationLog{
		ID:            util.NewXID(),
		Operator:      operator,
		OperatorIP:    operatorIP,
		UserAgent:     userAgent,
		Module:        auditModuleAuth,
		ActionType:    actionType,
		Description:   description,
		RequestPath:   "/api/v1/login",
		RequestMethod: "POST",
		Status:        success,
		StatusCode:    statusCode,
		ErrorMsg:      errorMsg,
		CreatedAt:     time.Now(),
	}
	go func() {
		if err := l.OperationLogDAL.Create(context.Background(), entry); err != nil {
			logging.Context(context.Background()).Error("Failed to write audit log",
				zap.Error(err),
				zap.String("operator", operator),
				zap.String("action_type", actionType),
			)
		}
	}()
}
