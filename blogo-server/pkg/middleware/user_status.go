// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/pkg/util"
)

// UserStatusCheckConfig 用户状态实时校验中间件配置
type UserStatusCheckConfig struct {
	// AllowedPathPrefixes: 仅这些路径前缀需要状态检查
	AllowedPathPrefixes []string

	// SkippedPathPrefixes: 完全跳过状态检查的路径（如登录、注册、验证码）
	SkippedPathPrefixes []string

	// CheckStatus: 根据 userID 查询数据库，返回 error 表示已禁用
	CheckStatus func(userID string) error
}

// UserStatusCheckWithConfig 创建全局用户状态校验中间件。
// 拦截所有携带有效 token 但 status=disabled 的用户，确保封禁即时生效。
func UserStatusCheckWithConfig(cfg UserStatusCheckConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 路径过滤
		if !AllowedPathPrefixes(c, cfg.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, cfg.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		if cfg.CheckStatus == nil {
			c.Next()
			return
		}

		// 优先从上下文获取（OptionalAuth 已设置），否则用轻量 token 解析
		userID := util.FromUserID(c.Request.Context())
		if userID == "" {
			token := util.GetToken(c)
			if token == "" {
				c.Next()
				return
			}
			userID = util.ParseTokenSub(token)
		}
		if userID == "" {
			c.Next()
			return
		}

		// 实时查库校验状态
		if err := cfg.CheckStatus(userID); err != nil {
			util.ResError(c, err)
			return
		}

		c.Next()
	}
}
