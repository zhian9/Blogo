// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package schema

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema" // 引入 rbac 的 schema 包
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

const (
	CommentStatusPending  = "pending"  // 待审核（可选）
	CommentStatusApproved = "approved" // 已通过
	CommentStatusRejected = "rejected" // 已拒绝
)

// Comment 评论表
type Comment struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`                // 评论唯一ID
	ArticleID string    `json:"article_id" gorm:"size:20;index;not null"`     // 关联文章ID
	UserID    string    `json:"user_id" gorm:"size:20;index;"`                // 评论者ID（可为空，支持游客评论）
	Username  string    `json:"username" gorm:"size:64;"`                     // 评论者名称（若为游客，则存储输入的名称）
	Email     string    `json:"email" gorm:"size:128;"`                       // 评论者邮箱（用于通知，可选）
	Content   string    `json:"content" gorm:"type:text;not null;"`           // 评论内容
	IP        string    `json:"ip" gorm:"size:45;"`                           // 评论者IP（IPv4/IPv6）
	UserAgent string    `json:"user_agent" gorm:"type:text;"`                 // 浏览器UA
	Status    string    `json:"status" gorm:"size:20;index;default:pending;"` // 评论状态（pending→审核后approved/rejected）
	ParentID  string    `json:"parent_id" gorm:"size:20;index;"`              // 父评论ID（用于回复评论，实现嵌套）
	IsTop     bool      `json:"is_top" gorm:"default:false;"`                 // 是否置顶（管理员可置顶优质评论）
	CreatedAt time.Time `json:"created_at" gorm:"index;"`                     // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`                     // 更新时间

	// 关联字段（全部手动加载，避免外键约束）
	User              *schema.User `json:"user,omitempty" gorm:"-"`               // 评论者信息（跨模块）
	Article           *Article     `json:"article,omitempty" gorm:"-"`            // 所属文章
	Parent            *Comment     `json:"parent,omitempty" gorm:"-"`             // 父评论
	ModerationMessage string       `json:"moderation_message,omitempty" gorm:"-"` // 审核提示
}

func (c *Comment) TableName() string {
	return config.C.FormatTableName("comment")
}

// CommentQueryParam 评论查询参数
type CommentQueryParam struct {
	util.PaginationParam
	ArticleID    string     `form:"article_id"`                                          // 按文章ID查询
	UserID       string     `form:"user_id"`                                             // 按用户ID查询
	Status       string     `form:"status" binding:"oneof=approved pending rejected ''"` // 状态筛选
	ParentID     string     `form:"parent_id"`                                           // 父评论ID（查子评论）
	IsTop        *bool      `form:"is_top"`                                              // 是否置顶
	CreatedAtGte *time.Time `form:"created_at_gte"`                                      // 创建时间 >=
	CreatedAtLte *time.Time `form:"created_at_lte"`                                      // 创建时间 <=
}

// CommentQueryOptions 查询选项
type CommentQueryOptions struct {
	util.QueryOptions
	WithUser    bool // 是否预加载用户信息
	WithArticle bool // 是否预加载文章信息
	WithParent  bool // 是否预加载父评论
}

// CommentQueryResult 查询结果
type CommentQueryResult struct {
	Data       Comments
	PageResult *util.PaginationResult
}

// Comments 评论切片
type Comments []*Comment

// ToIDs 返回评论ID列表
func (c Comments) ToIDs() []string {
	var ids []string
	for _, comment := range c {
		ids = append(ids, comment.ID)
	}
	return ids
}

// CommentForm 评论表单（用于创建评论）
type CommentForm struct {
	ArticleID string `json:"article_id" binding:"required"`       // 文章ID
	Content   string `json:"content" binding:"required,max=2000"` // 评论内容
	ParentID  string `json:"parent_id" binding:"omitempty"`       // 父评论ID（可选）
	Username  string `json:"username"`                            // 游客需填名称
	Email     string `json:"email" binding:"omitempty,email"`     // 邮箱（可选，但若填则需合法）
	UserID    string `json:"user_id" binding:"omitempty"`         // 已登录用户ID（由 token 解析，前端不应传）
}

// Validate 验证评论表单
func (cf *CommentForm) Validate() error {
	// 如果 UserID 为空（游客），则 Username 必填
	if cf.UserID == "" && cf.Username == "" {
		return errors.BadRequest("", "Username is required for guest comments")
	}

	// 验证邮箱格式（如果提供了）
	if cf.Email != "" {
		if err := validator.New().Var(cf.Email, "email"); err != nil {
			return errors.BadRequest("", "Invalid email address")
		}
	}

	// 内容敏感词过滤（可选，此处仅示意）
	if util.ContainsSensitiveWords(cf.Content) {
		return errors.BadRequest("", "Comment contains sensitive words")
	}

	return nil
}

// FillTo 将表单数据填充到 Comment 模型
func (cf *CommentForm) FillTo(comment *Comment, ip, userAgent string) error {
	comment.ArticleID = cf.ArticleID
	comment.Content = cf.Content
	comment.ParentID = cf.ParentID
	comment.IP = ip
	comment.UserAgent = userAgent

	// 已登录用户：user_id 从 JWT Token 解析（在 biz 层赋值），不信任前端传参
	// 游客评论：user_id 为空
	if cf.UserID == "" {
		comment.Username = cf.Username
		comment.Email = cf.Email
	}

	// 动态审核策略：从全局配置读取 comment_moderation
	if config.C.General.CommentModeration {
		comment.Status = CommentStatusPending
		comment.ModerationMessage = "评论发表成功，已进入审核队列"
	} else {
		comment.Status = CommentStatusApproved
	}

	return nil
}
