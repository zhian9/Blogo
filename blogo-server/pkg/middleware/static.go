// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// StaticConfig 静态文件配置
type StaticConfig struct {
	SkippedPathPrefixes []string
	Root                string
}

// StaticWithConfig 根据配置创建静态文件服务中间件。
func StaticWithConfig(config StaticConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 路径过滤
		if SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		// 构建文件系统路径
		p := c.Request.URL.Path
		fpath := filepath.Join(config.Root, filepath.FromSlash(p))
		_, err := os.Stat(fpath)
		// 检查文件是否存在
		if err != nil && os.IsNotExist(err) {
			fpath = filepath.Join(config.Root, "index.html")
		}
		// 返回文件
		c.File(fpath)

		// 终止后续中间件
		c.Abort()
	}
}
