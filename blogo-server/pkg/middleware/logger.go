// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"fmt"
	"mime"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
)

// LoggerConfig 定义日志中间件的配置
type LoggerConfig struct {
	AllowedPathPrefixes      []string //白名单
	SkippedPathPrefixes      []string //黑名单
	MaxOutputRequestBodyLen  int      // 请求体最大输出长度 B
	MaxOutputResponseBodyLen int      // 响应体最大输出长度 B
}

// DefaultLoggerConfig 默认日志配置
var DefaultLoggerConfig = LoggerConfig{
	MaxOutputRequestBodyLen:  1024 * 1024, // 1 MB
	MaxOutputResponseBodyLen: 1024 * 1024, // 1MB
}

// Logger 返回使用默认配置的日志中间件
func Logger() gin.HandlerFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}

// LoggerWithConfig 根据自定义配置创建日志中间件
func LoggerWithConfig(cfg LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 路径过滤:白名单 + 黑名单
		if !AllowedPathPrefixes(c, cfg.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, cfg.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		// 记录请求基本信息
		start := time.Now()
		contentType := c.Request.Header.Get("Content-Type")

		// 构建结构化日志字段（zap.Field）
		fields := []zap.Field{
			zap.String("client_ip", c.ClientIP()),                // 客户端真实 IP（支持 X-Forwarded-For）
			zap.String("method", c.Request.Method),               // HTTP 方法
			zap.String("path", c.Request.URL.Path),               // 请求路径
			zap.String("user_agent", c.Request.UserAgent()),      // User-Agent
			zap.String("referer", c.Request.Referer()),           // Referer
			zap.String("uri", c.Request.RequestURI),              // 完整 URI（含查询参数）
			zap.String("host", c.Request.Host),                   // Host
			zap.String("remote_addr", c.Request.RemoteAddr),      // TCP 连接远端地址
			zap.String("proto", c.Request.Proto),                 // HTTP 协议版本
			zap.Int64("content_length", c.Request.ContentLength), // 请求体长度
			zap.String("content_type", contentType),              // Content-Type
			zap.String("pragma", c.Request.Header.Get("Pragma")), // Pragma（缓存控制）
		}

		c.Next()

		//记录请求体
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			mediaType, _, _ := mime.ParseMediaType(contentType)
			if mediaType == "application/json" {
				//从 context 中获取 request
				if v, ok := c.Get(util.ReqBodyKey); ok {
					if b, ok := v.([]byte); ok && len(b) <= cfg.MaxOutputRequestBodyLen {
						fields = append(fields, zap.String("body", string(b)))
					}
				}
			}
		}
		// 记录响应信息
		cost := time.Since(start).Nanoseconds() / 1e6 // 耗时（毫秒）
		fields = append(fields,
			zap.Int64("cost", cost),                                              // 处理耗时
			zap.Int("status", c.Writer.Status()),                                 // HTTP 状态码
			zap.String("res_time", time.Now().Format("2006-01-02 15:04:05.999")), // 响应时间
			zap.Int("res_size", c.Writer.Size()),                                 // 响应体大小（字节）
		)
		// 记录响应体（同样受长度限制）
		if v, ok := c.Get(util.ResBodyKey); ok {
			if b, ok := v.([]byte); ok && len(b) <= cfg.MaxOutputResponseBodyLen {
				fields = append(fields, zap.String("res_body", string(b)))
			}
		}

		// 输出日志信息
		ctx := c.Request.Context()
		ctx = logging.NewTag(ctx, logging.TagKeyRequest)
		logging.Access(ctx, fmt.Sprintf("[HTTP] %s-%s-%d (%dms)",
			c.Request.URL.Path, c.Request.Method, c.Writer.Status(), cost), logging.SanitizeFields(fields)...)
	}
}
