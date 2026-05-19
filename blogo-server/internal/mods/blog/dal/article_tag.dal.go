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
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// GetArticleTagDB 返回文章-标签中间表的 GORM 查询实例
func GetArticleTagDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ArticleTag))
}

// ArticleTag 文章-标签中间表数据访问对象
type ArticleTag struct {
	DB *gorm.DB
}

// Create 创建单条文章-标签关联
func (at *ArticleTag) Create(ctx context.Context, item *schema.ArticleTag) error {
	result := GetArticleTagDB(ctx, at.DB).Create(item)
	return errors.WithStack(result.Error)
}

// CreateBatch 批量创建文章-标签关联
func (at *ArticleTag) CreateBatch(ctx context.Context, items []schema.ArticleTag) error {
	if len(items) == 0 {
		return nil
	}
	result := GetArticleTagDB(ctx, at.DB).CreateInBatches(items, 100)
	return errors.WithStack(result.Error)
}

// DeleteByArticleID 删除某篇文章的所有标签关联
func (at *ArticleTag) DeleteByArticleID(ctx context.Context, articleID string) error {
	if articleID == "" {
		return nil
	}
	result := GetArticleTagDB(ctx, at.DB).Where("article_id = ?", articleID).Delete(new(schema.ArticleTag))
	return errors.WithStack(result.Error)
}

// DeleteByTagID 删除某个标签的所有文章关联（谨慎使用）
func (at *ArticleTag) DeleteByTagID(ctx context.Context, tagID string) error {
	if tagID == "" {
		return nil
	}
	result := GetArticleTagDB(ctx, at.DB).Where("tag_id = ?", tagID).Delete(new(schema.ArticleTag))
	return errors.WithStack(result.Error)
}

// GetTagsByArticleID 根据文章 ID 获取所有关联的标签 ID 列表
func (at *ArticleTag) GetTagsByArticleID(ctx context.Context, articleID string) ([]string, error) {
	var tagIDs []string
	err := GetArticleTagDB(ctx, at.DB).
		Where("article_id = ?", articleID).
		Pluck("tag_id", &tagIDs).Error
	return tagIDs, errors.WithStack(err)
}

// GetArticlesByTagID 根据标签 ID 获取所有关联的文章 ID 列表
func (at *ArticleTag) GetArticlesByTagID(ctx context.Context, tagID string) ([]string, error) {
	var articleIDs []string
	err := GetArticleTagDB(ctx, at.DB).
		Where("tag_id = ?", tagID).
		Pluck("article_id", &articleIDs).Error
	return articleIDs, errors.WithStack(err)
}

// Exists 检查某篇文章是否已关联某个标签
func (at *ArticleTag) Exists(ctx context.Context, articleID, tagID string) (bool, error) {
	ok, err := util.Exists(ctx, GetArticleTagDB(ctx, at.DB).Where("article_id = ? AND tag_id = ?", articleID, tagID))
	return ok, errors.WithStack(err)
}
