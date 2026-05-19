// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"fmt"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
)

// RecoveryConfig 定义 Panic 恢复中间件的配置。
type RecoveryConfig struct {
	// Skip: 调用栈跳过层数
	// 默认值 3：跳过 recovery 中间件自身的调用栈
	Skip int
}

// DefaultRecoveryConfig 默认配置（跳过 3 层）
var DefaultRecoveryConfig = RecoveryConfig{
	Skip: 3,
}

// Recovery 返回使用默认配置的 Panic 恢复中间件。
func Recovery() gin.HandlerFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig)
}

// RecoveryWithConfig 根据配置创建 Panic 恢复中间件。
// 功能：
//   - 捕获 panic 避免服务崩溃
//   - 记录带堆栈的详细错误日志
//   - 安全处理敏感头（如 Authorization）
//   - 返回标准 500 错误响应
func RecoveryWithConfig(config RecoveryConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			// 捕获 panic
			if rv := recover(); rv != nil {
				// 1. 设置日志上下文（标记为 recovery 日志）
				ctx := c.Request.Context()
				ctx = logging.NewTag(ctx, logging.TagKeyRecovery)

				// 2. 构建日志字段
				var fields []zap.Field

				// 2.1 记录 panic 值
				fields = append(fields, zap.Strings("error", []string{fmt.Sprintf("%v", rv)}))

				// 2.2 记录调用栈（跳过指定层数）
				fields = append(fields, zap.StackSkip("stack", config.Skip))

				// 2.3 调试模式下记录请求头（脱敏处理）
				if gin.IsDebugging() {
					// 获取原始 HTTP 请求（不含 Body）
					httpRequest, _ := httputil.DumpRequest(c.Request, false)
					headers := strings.Split(string(httpRequest), "\r\n")

					// 脱敏敏感头（如 Authorization）
					for idx, header := range headers {
						current := strings.SplitN(header, ":", 2) // 只分割一次
						if len(current) == 2 && current[0] == "Authorization" {
							headers[idx] = current[0] + ": *"
						}
					}
					fields = append(fields, zap.Strings("headers", headers))
				}

				// 3. 记录错误日志
				logging.Context(ctx).Error(
					fmt.Sprintf("[Recovery] %s panic recovered", time.Now().Format("2006/01/02 - 15:04:05")),
					fields...,
				)

				// 4. 返回 500 错误响应（隐藏敏感信息）
				util.ResError(c, errors.InternalServerError("", "Internal server error, please try again later"))
			}
		}()

		// 执行后续中间件/Handler
		c.Next()
	}
}
