// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SkippedPathPrefixes 需要跳过的前缀
func SkippedPathPrefixes(c *gin.Context, prefixes ...string) bool {
	// 如果前缀列表为空，表示没有路径需要跳过
	if len(prefixes) == 0 {
		return false
	}

	//获取当前请求路径
	path := c.Request.URL.Path
	// 匹配前缀
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	// 匹配失败，不跳过
	return false
}

// AllowedPathPrefixes 判断当前请求路径是否匹配任意一个允许的前缀。
// 如果匹配，则表示该路径是“被允许的”；否则应被拒绝。
func AllowedPathPrefixes(c *gin.Context, prefixes ...string) bool {
	// 如果允许列表为空，表示所有路径都允许
	if len(prefixes) == 0 {
		return true
	}

	path := c.Request.URL.Path
	// 匹配前缀
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// Empty 返回一个空的 gin 中间件
func Empty() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
