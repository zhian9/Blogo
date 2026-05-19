// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
// Project repository: https://github.com/zhian9/go-admin

package dal

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

func GetArticleLikeDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ArticleLike))
}

type ArticleLike struct {
	DB *gorm.DB
}

// Create 创建点赞记录（原子操作：先查后插，避免重复）
func (a *ArticleLike) Create(ctx context.Context, like *schema.ArticleLike) error {
	result := GetArticleLikeDB(ctx, a.DB).Create(like)
	return errors.WithStack(result.Error)
}

// Delete 删除点赞记录（取消点赞）
func (a *ArticleLike) Delete(ctx context.Context, userID, articleID string) error {
	result := GetArticleLikeDB(ctx, a.DB).
		Where("user_id = ? AND article_id = ?", userID, articleID).
		Delete(new(schema.ArticleLike))
	return errors.WithStack(result.Error)
}

// Exists 检查用户是否已点赞
func (a *ArticleLike) Exists(ctx context.Context, userID, articleID string) (bool, error) {
	ok, err := util.Exists(ctx, GetArticleLikeDB(ctx, a.DB).
		Where("user_id = ? AND article_id = ?", userID, articleID))
	return ok, errors.WithStack(err)
}

// CountByArticleID 统计文章点赞数（原子查询）
func (a *ArticleLike) CountByArticleID(ctx context.Context, articleID string) (int64, error) {
	var count int64
	err := GetArticleLikeDB(ctx, a.DB).
		Where("article_id = ?", articleID).
		Count(&count).Error
	return count, errors.WithStack(err)
}

// GetLikedArticleIDs 获取用户点赞的文章ID列表
func (a *ArticleLike) GetLikedArticleIDs(ctx context.Context, userID string) ([]string, error) {
	var ids []string
	err := GetArticleLikeDB(ctx, a.DB).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Pluck("article_id", &ids).Error
	return ids, errors.WithStack(err)
}

// DeleteByArticleIDs 级联清理：文章删除时删除相关点赞
func (a *ArticleLike) DeleteByArticleIDs(ctx context.Context, articleIDs []string) error {
	if len(articleIDs) == 0 {
		return nil
	}
	result := GetArticleLikeDB(ctx, a.DB).
		Where("article_id IN ?", articleIDs).
		Delete(new(schema.ArticleLike))
	return errors.WithStack(result.Error)
}
