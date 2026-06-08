// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	rschema "github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// GetCommentDB 根据上下文返回评论表的 GORM 查询实例
func GetCommentDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Comment))
}

// Comment 评论数据访问对象
type Comment struct {
	DB *gorm.DB
}

// loadArticles 手动加载评论关联的文章信息
func (c *Comment) loadArticles(ctx context.Context, comments schema.Comments) error {
	var ids []string
	for _, cm := range comments {
		if cm.ArticleID != "" {
			ids = append(ids, cm.ArticleID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var articles []*schema.Article
	if err := util.GetDB(ctx, c.DB).Model(new(schema.Article)).Where("id IN ?", ids).Find(&articles).Error; err != nil {
		return errors.WithStack(err)
	}
	m := make(map[string]*schema.Article, len(articles))
	for _, a := range articles {
		m[a.ID] = a
	}
	for _, cm := range comments {
		cm.Article = m[cm.ArticleID]
	}
	return nil
}

// loadUsers 手动加载评论者的用户信息（User 字段标记为 gorm:"-" 无法自动 Preload）
func (c *Comment) loadUsers(ctx context.Context, comments schema.Comments) error {
	var ids []string
	for _, cm := range comments {
		if cm.UserID != "" {
			ids = append(ids, cm.UserID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var users []rschema.User
	if err := util.GetDB(ctx, c.DB).Model(new(rschema.User)).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return errors.WithStack(err)
	}
	userMap := make(map[string]*rschema.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}
	for _, cm := range comments {
		if cm.UserID != "" {
			cm.User = userMap[cm.UserID]
		}
	}
	return nil
}

// Query 根据查询参数分页查询评论列表
// 支持：文章ID、用户ID、状态、父评论ID、置顶、时间范围等筛选
func (c *Comment) Query(ctx context.Context, params schema.CommentQueryParam, opts ...schema.CommentQueryOptions) (*schema.CommentQueryResult, error) {
	var opt schema.CommentQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetCommentDB(ctx, c.DB)

	// ===== 条件查询 =====
	if v := params.ArticleID; len(v) > 0 {
		db = db.Where("article_id = ?", v)
	}
	if v := params.ProjectID; len(v) > 0 {
		db = db.Where("project_id = ?", v)
	}
	if v := params.UserID; len(v) > 0 {
		db = db.Where("user_id = ?", v)
	}
	if v := params.Status; len(v) > 0 {
		db = db.Where("status = ?", v)
	}
	if v := params.ParentID; len(v) > 0 {
		db = db.Where("parent_id = ?", v) // 查询子评论
	} else {
		db = db.Where("parent_id = '' OR parent_id IS NULL") // 默认只查顶级评论
	}
	if v := params.IsTop; v != nil {
		db = db.Where("is_top = ?", *v)
	}
	if v := params.CreatedAtGte; v != nil {
		db = db.Where("created_at >= ?", *v)
	}
	if v := params.CreatedAtLte; v != nil {
		db = db.Where("created_at <= ?", *v)
	}

	// ===== 按创建时间倒序（最新评论在前）=====
	db = db.Order("created_at DESC")

	// ===== 执行分页查询 =====
	var list schema.Comments
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if opt.WithArticle && len(list) > 0 {
		if err := c.loadArticles(ctx, list); err != nil {
			return nil, err
		}
	}
	if opt.WithUser && len(list) > 0 {
		if err := c.loadUsers(ctx, list); err != nil {
			return nil, err
		}
	}

	return &schema.CommentQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

// Get 根据 ID 获取单条评论
func (c *Comment) Get(ctx context.Context, id string, opts ...schema.CommentQueryOptions) (*schema.Comment, error) {
	var opt schema.CommentQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	comment := new(schema.Comment)
	db := GetCommentDB(ctx, c.DB).Where("id = ?", id)

	ok, err := util.FindOne(ctx, db, opt.QueryOptions, comment)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	if opt.WithArticle {
		if err := c.loadArticles(ctx, schema.Comments{comment}); err != nil {
			return nil, err
		}
	}
	if opt.WithUser {
		if err := c.loadUsers(ctx, schema.Comments{comment}); err != nil {
			return nil, err
		}
	}
	return comment, nil
}

// ExistsID 检查评论 ID 是否存在
func (c *Comment) ExistsID(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetCommentDB(ctx, c.DB).Where("id = ?", id))
	return ok, errors.WithStack(err)
}

// Create 创建新评论
func (c *Comment) Create(ctx context.Context, comment *schema.Comment) error {
	result := GetCommentDB(ctx, c.DB).Create(comment)
	return errors.WithStack(result.Error)
}

// Update 更新评论信息
// selectFields: 指定更新字段，为空则更新所有字段（除 created_at）
func (c *Comment) Update(ctx context.Context, comment *schema.Comment, selectFields ...string) error {
	db := GetCommentDB(ctx, c.DB).Where("id = ?", comment.ID)

	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("id", "created_at", "user_id", "article_id") // 保护关键字段
	}

	result := db.Updates(comment)
	return errors.WithStack(result.Error)
}

// UpdateByMap 根据 ID 和字段映射更新评论（仅允许更新安全字段）
func (c *Comment) UpdateByMap(ctx context.Context, id string, updates map[string]interface{}) error {
	// 白名单：允许更新的字段
	allowedFields := map[string]bool{
		"content":   true,
		"status":    true,
		"is_top":    true,
		"parent_id": true, // 谨慎开放
	}

	// 过滤非法字段
	safeUpdates := make(map[string]interface{})
	for k, v := range updates {
		if allowedFields[k] {
			safeUpdates[k] = v
		}
	}

	if len(safeUpdates) == 0 {
		return nil // 无有效字段可更新
	}

	result := GetCommentDB(ctx, c.DB).
		Where("id = ?", id).
		Updates(safeUpdates)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除评论（硬删除）
// 注意：应级联删除其子评论（在 Service 层实现）
func (c *Comment) Delete(ctx context.Context, id string) error {
	result := GetCommentDB(ctx, c.DB).Where("id = ?", id).Delete(new(schema.Comment))
	return errors.WithStack(result.Error)
}

// DeleteByArticleID 根据文章 ID 删除所有评论（用于删除文章时清理）
func (c *Comment) DeleteByArticleID(ctx context.Context, articleID string) error {
	if articleID == "" {
		return nil
	}
	result := GetCommentDB(ctx, c.DB).Where("article_id = ?", articleID).Delete(new(schema.Comment))
	return errors.WithStack(result.Error)
}

// DeleteByProjectID 根据项目 ID 删除所有评论（用于删除项目时清理）
func (c *Comment) DeleteByProjectID(ctx context.Context, projectID string) error {
	if projectID == "" {
		return nil
	}
	result := GetCommentDB(ctx, c.DB).Where("project_id = ?", projectID).Delete(new(schema.Comment))
	return errors.WithStack(result.Error)
}

// CountByProjectID 统计某个项目的评论总数（仅顶级评论）
func (c *Comment) CountByProjectID(ctx context.Context, projectID string) (int64, error) {
	var count int64
	err := GetCommentDB(ctx, c.DB).
		Where("project_id = ?", projectID).
		Where("parent_id = '' OR parent_id IS NULL").
		Count(&count).Error
	return count, errors.WithStack(err)
}

// QueryByProjectID 获取某个项目的所有已通过评论
func (c *Comment) QueryByProjectID(ctx context.Context, projectID string) (schema.Comments, error) {
	var list schema.Comments
	err := GetCommentDB(ctx, c.DB).
		Where("project_id = ? AND status = ?", projectID, schema.CommentStatusApproved).
		Order("is_top DESC, created_at ASC").
		Find(&list).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(list) > 0 {
		parentIDs := make([]string, 0)
		for _, cm := range list {
			if cm.ParentID != "" {
				parentIDs = append(parentIDs, cm.ParentID)
			}
		}
		if len(parentIDs) > 0 {
			var parents schema.Comments
			GetCommentDB(ctx, c.DB).Where("id IN ?", parentIDs).Find(&parents)
			parentMap := make(map[string]*schema.Comment)
			for _, p := range parents {
				parentMap[p.ID] = p
			}
			for _, cm := range list {
				if p, ok := parentMap[cm.ParentID]; ok {
					cm.Parent = p
				}
			}
		}
		if err := c.loadUsers(ctx, list); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return list, nil
}

// CommentStats 评论统计
type CommentStats struct {
	Total    int64 `json:"total"`
	Pending  int64 `json:"pending"`
	Approved int64 `json:"approved"`
	Rejected int64 `json:"rejected"`
}

// Stats 获取全站评论状态统计
func (c *Comment) Stats(ctx context.Context) (*CommentStats, error) {
	s := &CommentStats{}
	base := GetCommentDB(ctx, c.DB)
	// 每次 Count 前创建新 Session，防止 WHERE 条件累积
	if err := base.Session(&gorm.Session{}).Count(&s.Total).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	if err := base.Session(&gorm.Session{}).Where("status = ?", schema.CommentStatusPending).Count(&s.Pending).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	if err := base.Session(&gorm.Session{}).Where("status = ?", schema.CommentStatusApproved).Count(&s.Approved).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	if err := base.Session(&gorm.Session{}).Where("status = ?", schema.CommentStatusRejected).Count(&s.Rejected).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return s, nil
}

// CountByArticleID 统计某篇文章的评论总数（仅顶级评论）
func (c *Comment) CountByArticleID(ctx context.Context, articleID string) (int64, error) {
	var count int64
	err := GetCommentDB(ctx, c.DB).
		Where("article_id = ?", articleID).
		Where("parent_id = '' OR parent_id IS NULL").
		Count(&count).Error
	return count, errors.WithStack(err)
}

// QueryByArticleID 获取某篇文章的所有已通过评论（含父评论预加载）。
// 返回平铺列表，前端自行构建嵌套结构。评论者信息通过 username/user_id 字段直接返回。
func (c *Comment) QueryByArticleID(ctx context.Context, articleID string) (schema.Comments, error) {
	var list schema.Comments
	err := GetCommentDB(ctx, c.DB).
		Where("article_id = ? AND status = ?", articleID, schema.CommentStatusApproved).
		Order("is_top DESC, created_at ASC").
		Find(&list).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// 批量加载父评论信息（用于展示 "回复 @xxx"）
	if len(list) > 0 {
		parentIDs := make([]string, 0)
		for _, cm := range list {
			if cm.ParentID != "" {
				parentIDs = append(parentIDs, cm.ParentID)
			}
		}
		if len(parentIDs) > 0 {
			var parents schema.Comments
			GetCommentDB(ctx, c.DB).Where("id IN ?", parentIDs).Find(&parents)
			parentMap := make(map[string]*schema.Comment)
			for _, p := range parents {
				parentMap[p.ID] = p
			}
			for _, cm := range list {
				if p, ok := parentMap[cm.ParentID]; ok {
					cm.Parent = p
				}
			}
		}
	}
	// 加载用户信息
	if len(list) > 0 {
		if err := c.loadUsers(ctx, list); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return list, nil
}

// GetReplies 获取某条评论的所有直接子评论（用于嵌套回复）
func (c *Comment) GetReplies(ctx context.Context, parentID string, limit int) (schema.Comments, error) {
	var replies schema.Comments
	db := GetCommentDB(ctx, c.DB).
		Where("parent_id = ?", parentID).
		Order("created_at ASC") // 按时间正序（最早回复在前）

	if limit > 0 {
		db = db.Limit(limit)
	}

	err := db.Find(&replies).Error
	return replies, errors.WithStack(err)
}
