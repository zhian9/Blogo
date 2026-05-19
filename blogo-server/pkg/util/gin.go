// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package util

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/zhian9/blogo-server/pkg/encoding/json"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/logging"
	"go.uber.org/zap"
)

// GetToken 从请求中提取访问令牌（Access Token）。
// 优先级：
//  1. Authorization: Bearer <token>
//  2. Authorization: <token>（无前缀）
func GetToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	prefix := "Bearer "

	// 检查 Bearer Token 格式
	if auth != "" && strings.HasPrefix(auth, prefix) {
		return auth[len(prefix):]
	}
	return auth
}

// GetClientIP 从请求中提取真实客户端 IP。
// 优先级：X-Real-IP > X-Forwarded-For 首段 > RemoteAddr。
func GetClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if fwd := c.GetHeader("X-Forwarded-For"); fwd != "" {
		if idx := strings.IndexByte(fwd, ','); idx > 0 {
			return strings.TrimSpace(fwd[:idx])
		}
		return strings.TrimSpace(fwd)
	}
	return c.ClientIP()
}

// ParseTokenSub 轻量解析 JWT payload 中的 sub 声明。
//
// Deprecated: 此函数不验证 JWT 签名，存在身份伪造风险。
// 仅用于需要轻量级 userID 提取的非安全场景（如操作日志记录）。
// 任何涉及授权的场景必须使用 jwtx.Auther.ParseSubject() 进行完整验证。
func ParseTokenSub(tokenStr string) string {
	parts := strings.Split(tokenStr, ".")
	if len(parts) < 2 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	return claims.Sub
}

// GetBodyData 从上下文中获取请求体（由 CopyBody 中间件提前注入）。
// 依赖常量 ReqBodyKey（通常在中间件中设置）。
func GetBodyData(c *gin.Context) []byte {
	if v, ok := c.Get(ReqBodyKey); ok {
		if b, ok := v.([]byte); ok {
			return b
		}
	}
	return nil
}

// ParseJSON 解析 JSON 请求体到结构体。
// 使用 Gin 的 ShouldBindJSON，自动校验 binding 标签。
// 返回结构化错误（400 Bad Request）。
func ParseJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return errors.BadRequest("", "Failed to parse json: %s", err.Error())
	}
	return nil
}

// ParseQuery 解析 URL 查询参数到结构体。
// 使用 Gin 的 ShouldBindQuery，支持嵌套结构体。
func ParseQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return errors.BadRequest("", "Failed to parse query: %s", err.Error())
	}
	return nil
}

// ParseForm 解析表单数据（application/x-www-form-urlencoded）到结构体。
// 使用 Gin 的 ShouldBindWith + binding.Form。
func ParseForm(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return errors.BadRequest("", "Failed to parse form: %s", err.Error())
	}
	return nil
}

// ResJSON 发送 JSON 响应。
// 功能：
//   - 使用自定义 json.Marshal（支持 time.Time 等）
//   - 将响应体存入上下文（供 Logger 中间件记录）
//   - 设置 Content-Type 和状态码
//   - 调用 c.Abort() 阻止后续中间件执行
func ResJSON(c *gin.Context, status int, v interface{}) {
	buf, err := json.Marshal(v)
	if err != nil {
		// JSON 序列化失败属于严重错误，直接 panic
		panic(err)
	}

	// 存入上下文（供 Logger 中间件使用）
	c.Set(ResBodyKey, buf)

	// 发送响应
	c.Data(status, "application/json; charset=utf-8", buf)

	// 终止后续中间件（如 Logger 的 c.Next() 不会再执行）
	c.Abort()
}

// ResSuccess 发送成功响应（带数据）。
// 格式：{ "success": true, "data": ... }
func ResSuccess(c *gin.Context, v interface{}) {
	ResJSON(c, http.StatusOK, ResponseResult{
		Success: true,
		Data:    v,
	})
}

// ResOK 发送成功响应（无数据）。
// 格式：{ "success": true }
func ResOK(c *gin.Context) {
	ResJSON(c, http.StatusOK, ResponseResult{
		Success: true,
	})
}

// ResPage 发送分页成功响应。
// 处理空数据情况（避免返回 null）。
func ResPage(c *gin.Context, v interface{}, pr *PaginationResult) {
	var total int64
	if pr != nil {
		total = pr.Total
	}

	// 处理空切片：避免 JSON 返回 null
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	if reflectValue.IsNil() {
		v = make([]interface{}, 0)
	}

	ResJSON(c, http.StatusOK, ResponseResult{
		Success: true,
		Data:    v,
		Total:   total,
	})
}

// ResError 发送错误响应。
// 功能：
//   - 自动识别 errors.Error 类型
//   - 5xx 错误记录详细日志（含堆栈）
//   - 生产环境隐藏 5xx 错误详情（仅返回 "Internal Server Error"）
func ResError(c *gin.Context, err error, status ...int) {
	var ierr *errors.Error

	// 1. 尝试转换为结构化错误
	if e, ok := errors.As(err); ok {
		ierr = e
	} else {
		// 非结构化错误 → 转为 500
		ierr = errors.FromError(errors.InternalServerError("", err.Error()))
	}

	// 2. 确定 HTTP 状态码
	code := int(ierr.Code)
	if len(status) > 0 {
		code = status[0] // 允许覆盖状态码
	}

	// 3. 处理 5xx 错误（记录日志 + 隐藏详情）
	if code >= 500 {
		ctx := c.Request.Context()
		ctx = logging.NewTag(ctx, logging.TagKeySystem)      // 标记为系统错误
		ctx = logging.NewStack(ctx, fmt.Sprintf("%+v", err)) // 记录堆栈
		logging.Context(ctx).Error("Internal server error", zap.Error(err))

		// 生产环境隐藏敏感错误详情
		ierr.Detail = http.StatusText(http.StatusInternalServerError)
	}

	// 4. 更新状态码并返回
	ierr.Code = int32(code)
	ResJSON(c, code, ResponseResult{Error: ierr})
}
