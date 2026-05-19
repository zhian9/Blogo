// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
// Project repository: https://github.com/zhian9/go-admin

package biz

import (
	"context"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

type ArticleFavorite struct {
	Trans              *util.Trans
	ArticleFavoriteDAL *dal.ArticleFavorite
	ArticleDAL         *dal.Article
}

func (af *ArticleFavorite) Query(ctx context.Context, params schema.ArticleFavoriteQueryParam) (*schema.ArticleFavoriteQueryResult, error) {
	params.Pagination = true
	result, err := af.ArticleFavoriteDAL.Query(ctx, params, schema.ArticleFavoriteQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{{Field: "created_at", Direction: util.DESC}},
		},
		WithArticle: true,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Create 收藏文章（幂等）
func (af *ArticleFavorite) Create(ctx context.Context, userID string, form *schema.ArticleFavoriteForm) error {
	article, err := af.ArticleDAL.Get(ctx, form.ArticleID)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.BadRequest(config.ErrBadRequest, "Article not found")
	}
	if article.Status != schema.ArticleStatusPublished {
		return errors.BadRequest(config.ErrBadRequest, "Cannot favorite unpublished article")
	}
	exists, err := af.ArticleFavoriteDAL.Exists(ctx, userID, form.ArticleID)
	if err != nil {
		return err
	}
	if exists {
		return nil // 幂等
	}
	fav := &schema.ArticleFavorite{ArticleID: form.ArticleID, UserID: userID}
	return af.Trans.Exec(ctx, func(ctx context.Context) error {
		return af.ArticleFavoriteDAL.Create(ctx, fav)
	})
}

// Delete 取消收藏（幂等）
func (af *ArticleFavorite) Delete(ctx context.Context, userID, articleID string) error {
	exists, err := af.ArticleFavoriteDAL.Exists(ctx, userID, articleID)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return af.Trans.Exec(ctx, func(ctx context.Context) error {
		return af.ArticleFavoriteDAL.Delete(ctx, userID, articleID)
	})
}

func (af *ArticleFavorite) IsFavorite(ctx context.Context, userID, articleID string) (bool, error) {
	return af.ArticleFavoriteDAL.Exists(ctx, userID, articleID)
}

func (af *ArticleFavorite) CountByUserID(ctx context.Context, userID string) (int64, error) {
	return af.ArticleFavoriteDAL.CountByUserID(ctx, userID)
}

func (af *ArticleFavorite) GetArticleIDsByUserID(ctx context.Context, userID string, limit int) ([]string, error) {
	return af.ArticleFavoriteDAL.GetArticleIDsByUserID(ctx, userID, limit)
}
