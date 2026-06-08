// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// 评论管理所需的角色码
var commentModeratorRoles = []string{"super_admin", "admin", "content_manager", "comment_moderator"}

// Comment 是评论管理业务的核心对象。
type Comment struct {
	Trans       *util.Trans   // 事务管理器
	CommentDAL  *dal.Comment  // 评论数据访问层
	ArticleDAL  *dal.Article  // 文章数据访问层（用于校验）
	ProjectDAL  *dal.Project  // 项目数据访问层（用于校验）
}

// canModerateComments 检查当前用户是否有评论管理权限。
func canModerateComments(ctx context.Context) bool {
	isRoot := util.FromIsRootUser(ctx)
	userCache := util.FromUserCache(ctx)
	fmt.Printf("[DEBUG canModerate] isRoot=%v roleIDs=%v roleCodes=%v\n", isRoot, userCache.RoleIDs, userCache.RoleCodes)
	if isRoot {
		return true
	}
	if len(userCache.RoleIDs) == 0 {
		return false
	}
	hasRole := userCache.HasAnyRoleCode(commentModeratorRoles)
	fmt.Printf("[DEBUG canModerate] hasRole=%v moderatorRoles=%v\n", hasRole, commentModeratorRoles)
	return hasRole
}

// Query 查询评论列表（支持嵌套、分页）。
func (c *Comment) Query(ctx context.Context, params schema.CommentQueryParam) (*schema.CommentQueryResult, error) {
	params.Pagination = true

	result, err := c.CommentDAL.Query(ctx, params, schema.CommentQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "created_at", Direction: util.DESC},
			},
		},
		WithArticle: true,
		WithParent:  params.ParentID != "",
		WithUser:    true,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get 获取单条评论。
func (c *Comment) Get(ctx context.Context, id string) (*schema.Comment, error) {
	comment, err := c.CommentDAL.Get(ctx, id, schema.CommentQueryOptions{

		WithArticle: true,
		WithParent:  true,
		WithUser:    true,
	})
	if err != nil {
		return nil, err
	} else if comment == nil {
		return nil, errors.NotFound("", "Comment not found")
	}
	return comment, nil
}

// Create 创建新评论（支持游客和登录用户）。
// 流程：
//  1. 校验文章/项目是否存在且已发布
//  2. 校验父评论（如果提供）
//  3. 事务内创建评论
func (c *Comment) Create(ctx context.Context, commentForm *schema.CommentForm, ip, userAgent string) (*schema.Comment, error) {
	// 1. 校验文章或项目
	if commentForm.ArticleID != "" {
		article, err := c.ArticleDAL.Get(ctx, commentForm.ArticleID)
		if err != nil {
			return nil, err
		} else if article == nil || article.Status != schema.ArticleStatusPublished {
			return nil, errors.BadRequest("", "Article not found or not published")
		}
	} else if commentForm.ProjectID != "" {
		project, err := c.ProjectDAL.Get(ctx, commentForm.ProjectID)
		if err != nil {
			return nil, err
		} else if project == nil || project.Status != schema.ProjectStatusPublished {
			return nil, errors.BadRequest("", "Project not found or not published")
		}
	}

	// 2. 校验父评论（如果提供）
	if commentForm.ParentID != "" {
		parent, err := c.CommentDAL.Get(ctx, commentForm.ParentID)
		if err != nil {
			return nil, err
		} else if parent == nil {
			return nil, errors.BadRequest("", "Parent comment not found")
		}
		// 可选：限制嵌套层级（如只允许一级回复）
	}

	// 3. 表单验证
	if err := commentForm.Validate(); err != nil {
		return nil, err
	}

	// 4. 初始化评论实体
	comment := &schema.Comment{
		ID:        util.NewXID(),
		UserID:    util.FromUserID(ctx),
		CreatedAt: time.Now(),
		// Status 由 FillTo 根据 config.General.CommentModeration 动态设置
	}

	// 5. 填充数据（含 IP 和 UA）
	if err := commentForm.FillTo(comment, ip, userAgent); err != nil {
		return nil, err
	}

	// 6. 事务内创建
	if err := c.Trans.Exec(ctx, func(ctx context.Context) error {
		return c.CommentDAL.Create(ctx, comment)
	}); err != nil {
		return nil, err
	}

	return comment, nil
}

// Update 更新评论（需评论管理权限）。
func (c *Comment) Update(ctx context.Context, id string, updateFields map[string]interface{}) error {
	if !canModerateComments(ctx) {
		return errors.Forbidden("", "仅评论管理员可执行此操作")
	}
	exists, err := c.CommentDAL.ExistsID(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Comment not found")
	}

	return c.Trans.Exec(ctx, func(ctx context.Context) error {
		return c.CommentDAL.UpdateByMap(ctx, id, updateFields)
	})
}

// Delete 删除评论（需评论管理权限）。
func (c *Comment) Delete(ctx context.Context, id string) error {
	if !canModerateComments(ctx) {
		return errors.Forbidden("", "仅评论管理员可执行此操作")
	}
	exists, err := c.CommentDAL.ExistsID(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Comment not found")
	}

	return c.Trans.Exec(ctx, func(ctx context.Context) error {
		return c.CommentDAL.Delete(ctx, id)
	})
}

// DeleteByArticleID 批量删除某篇文章的所有评论（用于文章删除）。
func (c *Comment) DeleteByArticleID(ctx context.Context, articleID string) error {
	if articleID == "" {
		return nil
	}
	return c.CommentDAL.DeleteByArticleID(ctx, articleID)
}

// Approve 批准评论（需评论管理权限）。
func (c *Comment) Approve(ctx context.Context, id string) error {
	if !canModerateComments(ctx) {
		return errors.Forbidden("", "仅评论管理员可执行此操作")
	}
	return c.Update(ctx, id, map[string]interface{}{"status": schema.CommentStatusApproved})
}

// Reject 拒绝评论（需评论管理权限）。
func (c *Comment) Reject(ctx context.Context, id string) error {
	if !canModerateComments(ctx) {
		return errors.Forbidden("", "仅评论管理员可执行此操作")
	}
	return c.Update(ctx, id, map[string]interface{}{"status": schema.CommentStatusRejected})
}

// GetByArticleID 获取文章的所有已通过评论（含用户信息 + 父评论预加载）。
// 返回平铺列表，前端根据 parent_id 自行构建树形结构。
func (c *Comment) GetByArticleID(ctx context.Context, articleID string) (schema.Comments, error) {
	return c.CommentDAL.QueryByArticleID(ctx, articleID)
}

// GetByProjectID 获取项目的所有已通过评论
func (c *Comment) GetByProjectID(ctx context.Context, projectID string) (schema.Comments, error) {
	return c.CommentDAL.QueryByProjectID(ctx, projectID)
}

// Stats 获取全站评论统计
func (c *Comment) Stats(ctx context.Context) (*dal.CommentStats, error) {
	return c.CommentDAL.Stats(ctx)
}
