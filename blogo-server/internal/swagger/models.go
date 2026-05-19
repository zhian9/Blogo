// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

// Package swagger provides shared Swagger/OpenAPI model definitions
// referenced by handler annotations via `@Success` / `@Failure` tags.
//
// These types are NOT used at runtime — they exist solely to give swag
// concrete struct types to embed in the generated swagger.json.
package swagger

import (
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	rbac "github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
)

// ═══════════════════════════════════════════════════════
// Generic response wrappers (match util.ResponseResult)
// ═══════════════════════════════════════════════════════

// SuccessResponse is the standard success response: { success: true, data: T }
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResponse is the standard paginated response: { success: true, data: [...], total: N }
type PageResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
	Total   int64       `json:"total" example:"42"`
}

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Success bool         `json:"success" example:"false"`
	Error   *errors.Error `json:"error,omitempty"`
}

// ═══════════════════════════════════════════════════════
// Concrete response types for swag to expand generics
// Each type wraps one schema entity so swag can emit
// the full $ref definition instead of "object".
// ═══════════════════════════════════════════════════════

// ── Auth ───────────────────────────────────────────────

type CaptchaResponse struct {
	Success bool          `json:"success"`
	Data    rbac.Captcha `json:"data"`
}

type LoginTokenResponse struct {
	Success bool             `json:"success"`
	Data    rbac.LoginToken `json:"data"`
}

// ── User ───────────────────────────────────────────────

type UserResponse struct {
	Success bool       `json:"success"`
	Data    rbac.User  `json:"data"`
}

type UserListResponse struct {
	Success bool        `json:"success"`
	Data    []rbac.User `json:"data"`
}

// ── Article ────────────────────────────────────────────

type ArticleResponse struct {
	Success bool           `json:"success"`
	Data    schema.Article `json:"data"`
}

type ArticleListResponse struct {
	Success bool            `json:"success"`
	Data    []schema.Article `json:"data"`
}

// ── Comment ────────────────────────────────────────────

type CommentResponse struct {
	Success bool           `json:"success"`
	Data    schema.Comment `json:"data"`
}

type CommentListResponse struct {
	Success bool             `json:"success"`
	Data    []schema.Comment `json:"data"`
}

// ── Category / Tag ─────────────────────────────────────

type CategoryListResponse struct {
	Success bool              `json:"success"`
	Data    []schema.Category `json:"data"`
}

type TagListResponse struct {
	Success bool         `json:"success"`
	Data    []schema.Tag `json:"data"`
}

// ── Statistics ─────────────────────────────────────────

type StatListResponse struct {
	Success bool               `json:"success"`
	Data    []schema.Statistics `json:"data"`
}

// ── Menu ───────────────────────────────────────────────

type MenuListResponse struct {
	Success bool        `json:"success"`
	Data    []rbac.Menu `json:"data"`
}

// ── Role ───────────────────────────────────────────────

type RoleListResponse struct {
	Success bool        `json:"success"`
	Data    []rbac.Role `json:"data"`
}

// ── OperationLog ───────────────────────────────────────

type OperationLogListResponse struct {
	Success bool                  `json:"success"`
	Data    []rbac.OperationLog   `json:"data"`
}
