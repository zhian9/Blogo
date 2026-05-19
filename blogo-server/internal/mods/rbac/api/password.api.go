// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/cachex"
	"github.com/zhian9/blogo-server/pkg/crypto/hash"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/mail"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

const (
	pwdResetCacheNS     = "pwd_reset"
	pwdRateLimitCacheNS = "pwd_ratelimit"
)

type Password struct {
	DB    *gorm.DB
	Cache cachex.Cacher
}

func generateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

func (p *Password) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	// 60秒发送频率限制
	ctx := c.Request.Context()
	if _, ok, _ := p.Cache.Get(ctx, pwdRateLimitCacheNS, req.Email); ok {
		util.ResError(c, errors.BadRequest("", "请60秒后再试"))
		return
	}

	var user schema.User
	if err := p.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			util.ResError(c, errors.NotFound("", "该邮箱未注册"))
			return
		}
		util.ResError(c, err)
		return
	}

	code := generateCode()
	if err := p.Cache.Set(ctx, pwdResetCacheNS, req.Email, code, 5*time.Minute); err != nil {
		util.ResError(c, err)
		return
	}
	// 设置60秒频率限制
	_ = p.Cache.Set(ctx, pwdRateLimitCacheNS, req.Email, "1", 60*time.Second)

	// 异步发送邮件
	go sendResetEmail(req.Email, code)

	util.ResOK(c)
}

func (p *Password) ResetPassword(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Code        string `json:"code" binding:"required,len=6"`
		NewPassword string `json:"new_password" binding:"required,min=6,max=64"`
	}
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	stored, ok, err := p.Cache.Get(ctx, pwdResetCacheNS, req.Email)
	if err != nil || !ok || stored != req.Code {
		util.ResError(c, errors.BadRequest("", "验证码错误或已过期"))
		return
	}

	var user schema.User
	if err := p.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		util.ResError(c, errors.NotFound("", "用户不存在"))
		return
	}

	hashed, err := hash.GeneratePassword(req.NewPassword)
	if err != nil {
		util.ResError(c, err)
		return
	}
	if err := p.DB.Model(&user).Update("password", hashed).Error; err != nil {
		util.ResError(c, err)
		return
	}

	_ = p.Cache.Delete(ctx, pwdResetCacheNS, req.Email)

	util.ResOK(c)
}

// sendResetEmail 异步发送密码重置验证码邮件
func sendResetEmail(to, code string) {
	body := fmt.Sprintf(`
	<div style="max-width:480px;margin:0 auto;padding:32px 24px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#f8f9fa;border-radius:12px">
	  <div style="text-align:center;margin-bottom:24px">
		<h2 style="color:#1a1a2e;margin:0">Blogo</h2>
		<p style="color:#666;font-size:14px;margin:4px 0 0">密码重置验证</p>
	  </div>
	  <div style="background:#fff;padding:24px;border-radius:10px;box-shadow:0 2px 8px rgba(0,0,0,0.04)">
		<p style="color:#333;font-size:14px;line-height:1.6;margin:0 0 16px">
		  您正在申请重置 Blogo 账号的登录密码。
		</p>
		<div style="text-align:center;padding:20px 0;background:#f0f3ff;border-radius:8px;margin-bottom:16px">
		  <span style="font-size:28px;font-weight:700;color:#4f6ef7;letter-spacing:6px;font-family:'Courier New',monospace">%s</span>
		  <p style="color:#888;font-size:12px;margin:8px 0 0">6位安全验证码 · 5分钟内有效</p>
		</div>
		<p style="color:#999;font-size:12px;line-height:1.5;margin:0">
		  如非本人操作，请忽略此邮件。验证码请勿泄露给他人。
		</p>
	  </div>
	</div>`, code)

	_ = mail.SendToOne(context.Background(), to, "【Blogo】密码重置验证码", body)
}

// CheckEmail 检测邮箱是否已被注册
func (p *Password) CheckEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		util.ResError(c, errors.BadRequest("", "请提供邮箱地址"))
		return
	}
	var count int64
	if err := p.DB.Model(&schema.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, gin.H{"available": count == 0})
}
