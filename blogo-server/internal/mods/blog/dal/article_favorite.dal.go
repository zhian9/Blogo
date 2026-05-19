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

func GetArticleFavoriteDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ArticleFavorite))
}

type ArticleFavorite struct {
	DB *gorm.DB
}

func (a *ArticleFavorite) Query(ctx context.Context, params schema.ArticleFavoriteQueryParam, opts ...schema.ArticleFavoriteQueryOptions) (*schema.ArticleFavoriteQueryResult, error) {
	var opt schema.ArticleFavoriteQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	db := GetArticleFavoriteDB(ctx, a.DB)
	if params.UserID != "" {
		db = db.Where("user_id = ?", params.UserID)
	}
	if v := params.ArticleID; v != "" {
		db = db.Where("article_id = ?", v)
	}
	if v := params.CreatedAtGte; v != nil {
		db = db.Where("created_at >= ?", *v)
	}
	if v := params.CreatedAtLte; v != nil {
		db = db.Where("created_at <= ?", *v)
	}
	if opt.WithArticle {
		db = db.Preload("Article")
	}
	var list schema.ArticleFavorites
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &schema.ArticleFavoriteQueryResult{Data: list, PageResult: pageResult}, nil
}

func (a *ArticleFavorite) Exists(ctx context.Context, userID, articleID string) (bool, error) {
	ok, err := util.Exists(ctx, GetArticleFavoriteDB(ctx, a.DB).
		Where("user_id = ? AND article_id = ?", userID, articleID))
	return ok, errors.WithStack(err)
}

func (a *ArticleFavorite) Create(ctx context.Context, fav *schema.ArticleFavorite) error {
	result := GetArticleFavoriteDB(ctx, a.DB).Create(fav)
	return errors.WithStack(result.Error)
}

func (a *ArticleFavorite) Delete(ctx context.Context, userID, articleID string) error {
	result := GetArticleFavoriteDB(ctx, a.DB).
		Where("user_id = ? AND article_id = ?", userID, articleID).
		Delete(new(schema.ArticleFavorite))
	return errors.WithStack(result.Error)
}

func (a *ArticleFavorite) CountByUserID(ctx context.Context, userID string) (int64, error) {
	if userID == "" {
		return 0, nil
	}
	var count int64
	err := GetArticleFavoriteDB(ctx, a.DB).Where("user_id = ?", userID).Count(&count).Error
	return count, errors.WithStack(err)
}

func (a *ArticleFavorite) GetArticleIDsByUserID(ctx context.Context, userID string, limit int) ([]string, error) {
	if userID == "" {
		return nil, nil
	}
	var ids []string
	db := GetArticleFavoriteDB(ctx, a.DB).Select("article_id").Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	err := db.Pluck("article_id", &ids).Error
	return ids, errors.WithStack(err)
}

func (a *ArticleFavorite) DeleteByArticleIDs(ctx context.Context, userID string, articleIDs []string) error {
	if len(articleIDs) == 0 {
		return nil
	}
	db := GetArticleFavoriteDB(ctx, a.DB).Where("article_id IN ?", articleIDs)
	if userID != "" {
		db = db.Where("user_id = ?", userID)
	}
	result := db.Delete(new(schema.ArticleFavorite))
	return errors.WithStack(result.Error)
}
