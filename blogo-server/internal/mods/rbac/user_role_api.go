// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package rbac

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const superAdminRoleCode = "super_admin"

// ChangeUserRole 修改用户角色（仅 super_admin 可操作，禁止自杀/降权超管）
func (r *RBAC) ChangeUserRole(c *gin.Context) {
	ctx := c.Request.Context()
	targetUserID := c.Param("id")

	// 1. 鉴权：仅 super_admin 可执行
	if !r.isCurrentUserSuperAdmin(ctx) {
		util.ResError(c, errors.Forbidden("", "仅超级管理员可变更用户角色"))
		return
	}

	// 2. 解析请求体
	var req struct {
		RoleCode string `json:"role_code" binding:"required"`
	}
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	// 3. 查找目标用户
	var targetUser schema.User
	if err := r.DB.Where("id = ?", targetUserID).First(&targetUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			util.ResError(c, errors.NotFound("", "用户不存在"))
			return
		}
		util.ResError(c, err)
		return
	}

	// 4. 死锁：禁止变更超级管理员角色
	if isSuperAdminUser(r.DB, targetUserID) {
		util.ResError(c, errors.Forbidden("", "系统核心超级管理员角色不可被变更或降权"))
		return
	}

	// 5. 查找目标角色
	var targetRole schema.Role
	if err := r.DB.Where("code = ?", req.RoleCode).First(&targetRole).Error; err != nil {
		util.ResError(c, errors.BadRequest("", "无效的角色码: "+req.RoleCode))
		return
	}

	// 6. 执行角色变更（事务）
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// 清除目标用户的所有现有角色
		if err := tx.Where("user_id = ?", targetUserID).Delete(&schema.UserRole{}).Error; err != nil {
			return err
		}
		// 绑定新角色
		ur := schema.UserRole{
			ID:        util.NewXID(),
			UserID:    targetUserID,
			RoleID:    targetRole.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return tx.Create(&ur).Error
	})
	if err != nil {
		util.ResError(c, err)
		return
	}

	// 7. 清除目标用户缓存（强制下次请求重新加载权限）
	if err := r.UserAPI.UserBIZ.Cache.Delete(ctx, config.CacheNSForUser, targetUserID); err != nil {
		logging.Context(ctx).Error("Failed to delete user cache after role change",
			zap.String("target_user", targetUserID), zap.Error(err))
	} else {
		logging.Context(ctx).Info("User cache cleared after role change",
			zap.String("target_user", targetUserID),
		)
	}

	// 8. 触发 Casbin 策略重载（如果启用）
	if r.Casbinx != nil {
		r.Casbinx.TriggerReload(ctx)
	}

	logging.Context(ctx).Info("user role changed",
		zap.String("target_user", targetUserID),
		zap.String("new_role", req.RoleCode),
	)

	util.ResOK(c)
}

// isCurrentUserSuperAdmin 检查当前登录用户是否为 super_admin
func (r *RBAC) isCurrentUserSuperAdmin(ctx context.Context) bool {
	if util.FromIsRootUser(ctx) {
		return true
	}
	userID := util.FromUserID(ctx)
	if userID == "" {
		return false
	}
	return isSuperAdminUser(r.DB, userID)
}

// isSuperAdminUser 检查用户是否拥有 super_admin 角色
func isSuperAdminUser(db *gorm.DB, userID string) bool {
	var count int64
	db.Table("user_role AS ur").
		Joins("JOIN role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.code IN ?", userID, []string{superAdminRoleCode, "admin"}).
		Count(&count)
	return count > 0
}

// CheckUserStatus 实时校验用户状态（供 Auth 中间件使用）。
func (r *RBAC) CheckUserStatus(c *gin.Context, userID string) error {
	if userID == config.C.General.Root.ID {
		return nil
	}
	var user schema.User
	if err := r.DB.Select("status").Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.Unauthorized("", "用户不存在")
	}
	if user.Status == schema.UserStatusInactive {
		return errors.Unauthorized("", "账号尚未激活，请查收注册邮箱中的激活邮件并完成验证")
	} else if user.Status != schema.UserStatusActivated {
		return errors.Unauthorized("", "您的账号已被系统管理员禁用，已被强制下线，请联系运维人员")
	}
	return nil
}

// CheckStatus 全局状态校验（供全局中间件 UserStatusCheck 使用，仅需 userID）。
func (r *RBAC) CheckStatus(userID string) error {
	if userID == config.C.General.Root.ID {
		return nil
	}
	var user schema.User
	if err := r.DB.Select("status").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil // 用户不存在不报错，可能是 token 中的旧用户
	}
	if user.Status == schema.UserStatusInactive {
		return errors.Unauthorized("", "账号尚未激活，请查收注册邮箱中的激活邮件并完成验证")
	} else if user.Status != schema.UserStatusActivated {
		return errors.Unauthorized("", "您的账号已被系统管理员禁用，已被强制下线，请联系运维人员")
	}
	return nil
}

// ChangeUserStatus 切换用户启用/禁用状态（仅 super_admin 可操作）
func (r *RBAC) ChangeUserStatus(c *gin.Context) {
	ctx := c.Request.Context()
	targetUserID := c.Param("id")

	// 鉴权：仅 super_admin 可执行
	if !r.isCurrentUserSuperAdmin(ctx) {
		util.ResError(c, errors.Forbidden("", "仅超级管理员可变更用户状态"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=activated freezed"`
	}
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	// 禁止冻结超级管理员自己
	if isSuperAdminUser(r.DB, targetUserID) && req.Status == "freezed" {
		util.ResError(c, errors.Forbidden("", "系统核心超级管理员不可被禁用"))
		return
	}

	if err := r.DB.Model(&schema.User{}).Where("id = ?", targetUserID).Update("status", req.Status).Error; err != nil {
		util.ResError(c, err)
		return
	}

	// 清除用户缓存（状态变更后强制重新加载）
	if err := r.UserAPI.UserBIZ.Cache.Delete(ctx, config.CacheNSForUser, targetUserID); err != nil {
		logging.Context(ctx).Error("Failed to delete user cache after status change",
			zap.String("target_user", targetUserID), zap.Error(err))
	}

	logging.Context(ctx).Info("user status changed",
		zap.String("target_user", targetUserID),
		zap.String("status", req.Status),
	)
	util.ResOK(c)
}
