// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSConfig cors 中间件的配置结构体
type CORSConfig struct {
	Enable                 bool     `mapstructure:"enable"`
	AllowAllOrigins        bool     `mapstructure:"allow_all_origins"`
	AllowOrigins           []string `mapstructure:"allow_origins"`
	AllowMethods           []string `mapstructure:"allow_methods"`
	AllowHeaders           []string `mapstructure:"allow_headers"`
	AllowCredentials       bool     `mapstructure:"allow_credentials"`
	ExposeHeaders          []string `mapstructure:"expose_headers"`
	MaxAge                 int      `mapstructure:"max_age"`
	AllowWildcard          bool     `mapstructure:"allow_wildcard"`
	AllowBrowserExtensions bool     `mapstructure:"allow_browser_extensions"`
	AllowWebSockets        bool     `mapstructure:"allow_web_sockets"`
	AllowFiles             bool     `mapstructure:"allow_files"`
}

// DefaultCORSConfig CORS默认配置
var DefaultCORSConfig = CORSConfig{
	AllowAllOrigins: true, // 设置为 true
	AllowOrigins:    []string{"*"},
	AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "HEAD", "PATCH", "OPTIONS"},
}

// CORSWithConfig 根据配置创建 CORS 中间件
func CORSWithConfig(cfg CORSConfig) gin.HandlerFunc {
	// 未启用 CORS 则返回空结构体
	if !cfg.Enable {
		return Empty()
	}

	return cors.New(cors.Config{
		AllowAllOrigins:        cfg.AllowAllOrigins,
		AllowOrigins:           cfg.AllowOrigins,
		AllowMethods:           cfg.AllowMethods,
		AllowHeaders:           cfg.AllowHeaders,
		AllowCredentials:       cfg.AllowCredentials,
		ExposeHeaders:          cfg.ExposeHeaders,
		MaxAge:                 time.Second * time.Duration(cfg.MaxAge), // 转换为 time.Duration
		AllowWildcard:          cfg.AllowWildcard,
		AllowBrowserExtensions: cfg.AllowBrowserExtensions,
		AllowWebSockets:        cfg.AllowWebSockets,
		AllowFiles:             cfg.AllowFiles,
	})
}
