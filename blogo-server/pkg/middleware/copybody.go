// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// CopyBodyConfig 定义请求体复制的配置。
type CopyBodyConfig struct {
	// AllowedPathPrefixes: 白名单路径前缀（仅这些路径复制请求体）
	// 空列表表示所有路径都复制。
	AllowedPathPrefixes []string

	// SkippedPathPrefixes: 黑名单路径前缀（这些路径跳过复制）
	SkippedPathPrefixes []string

	// MaxContentLen: 请求体最大长度（字节）
	// 超过此长度将返回 413 Request Entity Too Large
	MaxContentLen int64
}

// DefaultCopyBodyConfig 默认配置（32MB 限制）
var DefaultCopyBodyConfig = CopyBodyConfig{
	MaxContentLen: 32 << 20, // 32 * 1024 * 1024 = 33,554,432 字节
}

// CopyBody 返回使用默认配置的请求体复制中间件。
func CopyBody() gin.HandlerFunc {
	return CopyBodyWithConfig(DefaultCopyBodyConfig)
}

// CopyBodyWithConfig 根据配置创建请求体复制中间件。
// 功能：
//   - 路径过滤（白名单 + 黑名单）
//   - 请求体大小限制（防止 OOM）
//   - Gzip 自动解压（支持压缩请求体）
//   - 请求体存入上下文（供 Logger/Auth 等中间件使用）
func CopyBodyWithConfig(config CopyBodyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 路径过滤
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			c.Request.Body == nil {
			c.Next()
			return
		}

		// 安全读取请求体（带大小限制）
		var (
			requestBody []byte
			err         error
		)

		// 使用 http.MaxBytesReader 限制读取大小（防止内存溢出）
		safe := http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxContentLen)

		// 检查是否为 Gzip 压缩请求体
		isGzip := false
		if c.GetHeader("Content-Encoding") == "gzip" {
			// 尝试创建 Gzip 解压 reader
			if reader, ierr := gzip.NewReader(safe); ierr == nil {
				isGzip = true
				// 读取并解压全部内容
				requestBody, err = io.ReadAll(reader)
				reader.Close() // 确保关闭解压器
			}
		}

		// 非 Gzip 请求体：直接读取
		if !isGzip {
			requestBody, err = io.ReadAll(safe)
		}

		// 处理读取错误

		if err != nil {
			// 可能原因：
			//   - 请求体超过 MaxContentLen（http.MaxBytesReader 返回错误）
			//   - Gzip 解压失败
			util.ResError(c, errors.RequestEntityTooLarge("", "Request body too large, limit %d byte", config.MaxContentLen))
			return
		}

		// 重置请求体并存入上下文
		// a) 关闭原始 Body（避免资源泄漏）
		c.Request.Body.Close()

		// b) 创建新的可读 Body（bytes.Buffer + io.NopCloser）
		bf := bytes.NewBuffer(requestBody)
		c.Request.Body = io.NopCloser(bf)

		// c) 存入上下文（供 Logger/Auth 等中间件使用）
		c.Set(util.ReqBodyKey, requestBody)

		// 继续执行后续中间件/Handler
		c.Next()
	}
}
