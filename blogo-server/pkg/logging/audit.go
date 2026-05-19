// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package logging

import (
	"context"
	"strings"

	"go.uber.org/zap"
)

// Audit 记录后台管理审计日志。
// 写入 audit.log (若启用), 同时写入主日志。
func Audit(ctx context.Context, msg string, fields ...zap.Field) {
	if AuditLogger != nil {
		AuditLogger.Info("[AUDIT] "+msg, append(fields, ctxFields(ctx)...)...)
	}
	ctx = NewTag(ctx, TagKeyAudit)
	Context(ctx).Info("[AUDIT] "+msg, fields...)
}

// Security 记录安全事件日志。
// 写入 security.log (若启用), 同时写入主日志。
func Security(ctx context.Context, msg string, fields ...zap.Field) {
	if SecurityLogger != nil {
		SecurityLogger.Warn("[SECURITY] "+msg, append(fields, ctxFields(ctx)...)...)
	}
	ctx = NewTag(ctx, TagKeySecurity)
	Context(ctx).Warn("[SECURITY] "+msg, fields...)
}

// Access 记录 HTTP 访问日志。
// 写入 access.log (若启用), 同时写入主日志。
func Access(ctx context.Context, msg string, fields ...zap.Field) {
	if AccessLogger != nil {
		AccessLogger.Info(msg, append(fields, ctxFields(ctx)...)...)
	}
	Context(ctx).Info(msg, fields...)
}

func ctxFields(ctx context.Context) []zap.Field {
	var fields []zap.Field
	if v := FromTraceID(ctx); v != "" {
		fields = append(fields, zap.String("trace_id", v))
	}
	if v := FromUserID(ctx); v != "" {
		fields = append(fields, zap.String("user_id", v))
	}
	return fields
}

// ── 敏感数据掩码 ───────────────────────────────────

var SensitiveFields = map[string]bool{
	"password":      true,
	"pass":          true,
	"token":         true,
	"jwt":           true,
	"secret":        true,
	"authorization": true,
	"cookie":        true,
	"set-cookie":    true,
	"access_token":  true,
	"refresh_token": true,
}

func SanitizeField(f zap.Field) zap.Field {
	if SensitiveFields[strings.ToLower(f.Key)] {
		return zap.String(f.Key, "***")
	}
	return f
}

func SanitizeFields(fields []zap.Field) []zap.Field {
	for i, f := range fields {
		fields[i] = SanitizeField(f)
	}
	return fields
}
