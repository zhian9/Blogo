// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package logging

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

const (
	TagKeyMain     = "main"     // 主流程日志
	TagKeyRecovery = "recovery" // panic 恢复日志
	TagKeyRequest  = "request"  // HTTP 请求日志
	TagKeyLogin    = "login"    // 登录操作
	TagKeyLogout   = "logout"   // 退出操作
	TagKeySystem   = "system"   // 系统级
	TagKeyOperate  = "operate"  // 通用业务操作
	TagKeyAudit    = "audit"    // 后台审计操作
	TagKeySecurity = "security" // 安全事件
)

type (
	ctxLoggerKey  struct{}
	ctxTraceIDKey struct{}
	ctxUserIDKey  struct{}
	ctxTagKey     struct{}
	ctxStackKey   struct{}
)

// NewLogger 将zap.logger绑定到context
func NewLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

// FromLogger 从context中获取zap.logger
func FromLogger(ctx context.Context) *zap.Logger {
	v := ctx.Value(ctxLoggerKey{})
	if v != nil {
		if logger, ok := v.(*zap.Logger); ok {
			return logger
		}
	}
	// 安全处理
	return zap.L()
}

// NewTraceID 将trace_id 存到 context
func NewTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxTraceIDKey{}, traceID)
}

// FromTraceID 从 context 中获取 trace_id
func FromTraceID(ctx context.Context) string {
	v := ctx.Value(ctxTraceIDKey{})
	if v != nil {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return ""
}

// NewUserID 将用户 ID 存入到 context 中
func NewUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxUserIDKey{}, userID)
}

// FromUserID 从 context 中获取 user_id
func FromUserID(ctx context.Context) string {
	v := ctx.Value(ctxUserIDKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// NewTag 将日志分类标签存入到 context
func NewTag(ctx context.Context, tag string) context.Context {
	return context.WithValue(ctx, ctxTagKey{}, tag)
}

// FromTag 从 context 中获取日志标签
func FromTag(ctx context.Context) string {
	v := ctx.Value(ctxTagKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// NewStack 将错误堆栈信息存入 context 中
func NewStack(ctx context.Context, stack string) context.Context {
	return context.WithValue(ctx, ctxStackKey{}, stack)
}

// FromStack 从 context 中获取堆栈信息
func FromStack(ctx context.Context) string {
	v := ctx.Value(ctxStackKey{})
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// Context 返回一个带有 trace_id,user_id,tag,stack 等的zap.Logger
func Context(ctx context.Context) *zap.Logger {
	var fields []zap.Field

	//添加trace_id
	if v := FromTraceID(ctx); v != "" {
		fields = append(fields, zap.String("trace_id", v))
	}
	//添加user_id
	if v := FromUserID(ctx); v != "" {
		fields = append(fields, zap.String("user_id", v))
	}
	//添加tag
	if v := FromTag(ctx); v != "" {
		fields = append(fields, zap.String("tag", v))
	}
	//添加stack
	if v := FromStack(ctx); v != "" {
		fields = append(fields, zap.String("stack", v))
	}
	return FromLogger(ctx).With(fields...)
}

// PrintLogger 日志输出器
type PrintLogger struct{}

// Printf 将格式化字符串作为Info 级别日志输出到全局 zap logger
func (pl *PrintLogger) Printf(format string, args ...interface{}) {
	zap.L().Info(fmt.Sprintf(format, args...))
}
