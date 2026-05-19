// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package schema

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/zhian9/blogo-server/pkg/errors"
)

// Captcha 表示验证码响应结构。
// 通常由 /api/v1/captcha 接口返回。
// @name Captcha
type Captcha struct {
	// CaptchaID: 验证码唯一标识（用于后续验证）
	// 前端需在登录时提交此 ID + 用户输入的验证码
	CaptchaID string `json:"captcha_id" example:"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`
}

// LoginForm 表示用户登录请求结构。
// 由前端提交到 /api/v1/login 接口。
// @name LoginForm
type LoginForm struct {
	// Username: 登录用户名（必填）
	// Example: admin
	Username string `json:"username" binding:"required" example:"admin"`

	// Password: 登录密码（必填）
	// Example: goadmin123456
	Password string `json:"password" binding:"required" example:"goadmin123456"`

	// CaptchaID: 验证码 ID（必填）
	// 来自 Captcha 接口的 captcha_id
	// Example: 6ba7b810-9dad-11d1-80b4-00c04fd430c8
	CaptchaID string `json:"captcha_id" binding:"required" example:"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`

	// CaptchaCode: 用户输入的验证码（必填）
	// 前端从验证码图片中识别的字符
	// Example: aB3x9
	CaptchaCode string `json:"captcha_code" binding:"required" example:"aB3x9"`
}

// Trim 对 LoginForm 的字符串字段进行去空格处理。
// 防止用户输入前后空格导致验证失败。
func (lf *LoginForm) Trim() *LoginForm {
	lf.Username = strings.TrimSpace(lf.Username)
	lf.CaptchaCode = strings.TrimSpace(lf.CaptchaCode)
	return lf
}

// UpdateLoginPassword 表示修改登录密码请求结构。
// 由已登录用户提交到 /api/v1/current/password 接口。
// @name UpdateLoginPassword
type UpdateLoginPassword struct {
	// OldPassword: 旧密码
	// 用于验证用户身份
	// Example: oldPass123
	OldPassword string `json:"old_password" binding:"required" example:"oldPass123"`

	// NewPassword: 新密码
	// Example: newSecurePass!2025
	NewPassword string `json:"new_password" binding:"required" example:"newSecurePass!2025"`
}

// LoginToken 表示登录成功后的令牌响应结构。
// 由 /api/v1/login 接口返回。
// @name LoginToken
type LoginToken struct {
	// AccessToken: 访问令牌（JWT 字符串）
	// 前端需在后续请求的 Authorization 头中携带
	// Example: eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.xxxxx
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.xxxxx"`

	// TokenType: 令牌类型（通常为 "Bearer"）
	// 使用方式: Authorization: ${token_type} ${access_token}
	// Example: Bearer
	TokenType string `json:"token_type" example:"Bearer"`

	// ExpiresAt: 令牌过期时间（Unix 时间戳，秒）
	// 前端可用于自动刷新令牌
	// Example: 1730466240
	ExpiresAt int64 `json:"expires_at" example:"1730466240"`
}

// UpdateCurrentUser 表示更新当前用户信息请求结构。
// 由已登录用户提交到 /api/v1/current/user 接口。
// @name UpdateCurrentUser
type UpdateCurrentUser struct {
	// Name: 用户姓名（必填，最大 64 字符）
	// Example: 李星云
	Name string `json:"name" binding:"required,max=64" example:"李星云"`

	// Phone: 手机号码（可选，最大 32 字符）
	// Example: +8613800138000
	Phone string `json:"phone" binding:"max=32" example:"+8613800138000"`

	// Email: 邮箱地址（可选，最大 128 字符）
	// Example: lxy911@gmail.com
	Email string `json:"email" binding:"max=128" example:"lxy911@gmail.com"`

	// Avatar: 头像 URL（可选，最大 512 字符）
	// Example: https://example.com/avatar.jpg
	Avatar string `json:"avatar" binding:"max=512" example:"https://example.com/avatar.jpg"`

	// Bio: 个人简介/个性签名（可选，最大 512 字符）
	// Example: 全栈开发者，热爱开源
	Bio string `json:"bio" binding:"max=512" example:"全栈开发者，热爱开源"`

	// Remark: 备注信息（可选，最大 1024 字符）
	// Example: 开发者
	Remark string `json:"remark" binding:"max=1024" example:"开发者"`
}

// RegisterForm 表示公开注册请求结构。
// 由未登录用户提交到 /api/v1/register 接口。
// @name RegisterForm
type RegisterForm struct {
	// Username: 登录用户名（必填）
	// Example: newuser2025
	Username string `json:"username" binding:"required,max=64" example:"newuser2025"`

	// Password: 登录密码（必填）
	// Example: MyPass123!
	Password string `json:"password" binding:"required,max=64" example:"MyPass123!"`

	// ConfirmPassword: 二次确认密码（必填）
	// Example: MyPass123!
	ConfirmPassword string `json:"confirm_password" binding:"required,max=64" example:"MyPass123!"`

	// Phone: 手机号（必填）
	// Example: +8613900001111
	Phone string `json:"phone" binding:"required,max=32" example:"+8613900001111"`

	// Email: 邮箱（必填，合法格式）
	// Example: user2025@example.com
	Email string `json:"email" binding:"required,max=128" example:"user2025@example.com"`

	// CaptchaID: 验证码 ID（必填）
	// Example: 6ba7b810-9dad-11d1-80b4-00c04fd430c8
	CaptchaID string `json:"captcha_id" binding:"required" example:"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`

	// CaptchaCode: 验证码输入（必填）
	// Example: xK9m2
	CaptchaCode string `json:"captcha_code" binding:"required" example:"xK9m2"`
}

// Trim 对 RegisterForm 的字符串字段进行去空格处理。
func (rf *RegisterForm) Trim() *RegisterForm {
	rf.Username = strings.TrimSpace(rf.Username)
	rf.CaptchaCode = strings.TrimSpace(rf.CaptchaCode)
	rf.Phone = strings.TrimSpace(rf.Phone)
	rf.Email = strings.TrimSpace(rf.Email)
	return rf
}

// Validate 校验注册表单
func (rf *RegisterForm) Validate() error {
	if rf.Password != rf.ConfirmPassword {
		return errors.BadRequest("", "Confirm password does not match")
	}
	if rf.Email != "" && validator.New().Var(rf.Email, "email") != nil {
		return errors.BadRequest("", "Invalid email address")
	}
	return nil
}
