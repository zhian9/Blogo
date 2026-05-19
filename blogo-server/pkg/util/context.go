// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package util

import (
	"context"

	"github.com/zhian9/blogo-server/pkg/encoding/json"

	"gorm.io/gorm"
)

// 使用私有 struct 作为 context.Key，确保唯一性（避免与其他包冲突）
type (
	traceIDCtx    struct{} // Trace ID（链路追踪）
	transCtx      struct{} // 数据库事务
	rowLockCtx    struct{} // 行锁标识（用于并发控制）
	userIDCtx     struct{} // 用户 ID
	userTokenCtx  struct{} // 用户 Token
	isRootUserCtx struct{} // 是否超级管理员
	userCacheCtx  struct{} // 用户缓存（角色等）
)

// TraceID 相关操作

// NewTraceID 将 TraceID 存入 context。
// 通常由中间件（如 TraceMiddleware）在请求入口设置。
func NewTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDCtx{}, traceID)
}

// FromTraceID 从 context 获取 TraceID。
func FromTraceID(ctx context.Context) string {
	v := ctx.Value(traceIDCtx{})
	if v != nil {
		return v.(string)
	}
	return ""
}

// 数据库事务相关操作

// NewTrans 将 GORM 事务 DB 实例存入 context。
// 用于跨多个 DAO 方法共享同一事务。
func NewTrans(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, transCtx{}, db)
}

// FromTrans 从 context 获取事务 DB 实例。
// 返回 (*gorm.DB, bool)，bool 表示是否存在事务。
func FromTrans(ctx context.Context) (*gorm.DB, bool) {
	v := ctx.Value(transCtx{})
	if v != nil {
		return v.(*gorm.DB), true
	}
	return nil, false
}

//  行锁标识操作

// NewRowLock 标记当前操作需要行级锁（如 SELECT FOR UPDATE）。
func NewRowLock(ctx context.Context) context.Context {
	return context.WithValue(ctx, rowLockCtx{}, true)
}

// FromRowLock 检查是否需要行锁。
func FromRowLock(ctx context.Context) bool {
	v := ctx.Value(rowLockCtx{})
	return v != nil && v.(bool)
}

// 用户身份相关操作

// NewUserID 设置当前用户 ID。
func NewUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDCtx{}, userID)
}

// FromUserID 获取当前用户 ID。
func FromUserID(ctx context.Context) string {
	v := ctx.Value(userIDCtx{})
	if v != nil {
		return v.(string)
	}
	return ""
}

// NewUserToken 设置用户 Token（JWT）。
func NewUserToken(ctx context.Context, userToken string) context.Context {
	return context.WithValue(ctx, userTokenCtx{}, userToken)
}

// FromUserToken 获取用户 Token。
func FromUserToken(ctx context.Context) string {
	v := ctx.Value(userTokenCtx{})
	if v != nil {
		return v.(string)
	}
	return ""
}

// NewIsRootUser 标记当前用户为超级管理员。
func NewIsRootUser(ctx context.Context) context.Context {
	return context.WithValue(ctx, isRootUserCtx{}, true)
}

// FromIsRootUser 检查是否为超级管理员。
func FromIsRootUser(ctx context.Context) bool {
	v := ctx.Value(isRootUserCtx{})
	return v != nil && v.(bool)
}

// 用户缓存操作

// UserCache 定义用户缓存结构（通常存储角色 ID 列表）。
type UserCache struct {
	RoleIDs   []string `json:"rids"`   // 角色 ID 列表
	RoleCodes []string `json:"rcodes"` // 角色 Code 列表（如 admin, content_manager）
}

// ParseUserCache 从 JSON 字符串解析 UserCache。
// 用于从缓存（如 Redis）中反序列化用户信息。
func ParseUserCache(s string) UserCache {
	var a UserCache
	if s == "" {
		return a
	}
	_ = json.Unmarshal([]byte(s), &a)
	return a
}

// HasAnyRoleCode 检查 UserCache 中是否包含任意目标角色码。
func (a UserCache) HasAnyRoleCode(codes []string) bool {
	for _, rc := range a.RoleCodes {
		for _, c := range codes {
			if rc == c {
				return true
			}
		}
	}
	return false
}

// String 将 UserCache 序列化为 JSON 字符串。
// 用于存储到缓存（如 Redis）。
func (a UserCache) String() string {
	return json.MarshalToString(a)
}

// NewUserCache 将用户缓存存入 context。
func NewUserCache(ctx context.Context, userCache UserCache) context.Context {
	return context.WithValue(ctx, userCacheCtx{}, userCache)
}

// FromUserCache 从 context 获取用户缓存。
func FromUserCache(ctx context.Context) UserCache {
	v := ctx.Value(userCacheCtx{})
	if v != nil {
		return v.(UserCache)
	}
	return UserCache{}
}
