// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
// Project repository: https://github.com/zhian9/go-admin

package biz

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

type ArticleLike struct {
	Trans          *util.Trans
	ArticleLikeDAL *dal.ArticleLike
	ArticleDAL     *dal.Article
}

// Like 点赞文章（幂等：已点赞则忽略）
func (al *ArticleLike) Like(ctx context.Context, userID string, form *schema.ArticleLikeForm) error {
	article, err := al.ArticleDAL.Get(ctx, form.ArticleID)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.BadRequest(config.ErrBadRequest, "Article not found")
	}
	if article.Status != schema.ArticleStatusPublished {
		return errors.BadRequest(config.ErrBadRequest, "Cannot like unpublished article")
	}
	exists, err := al.ArticleLikeDAL.Exists(ctx, userID, form.ArticleID)
	if err != nil {
		return err
	}
	if exists {
		return nil // 幂等
	}
	like := &schema.ArticleLike{ArticleID: form.ArticleID, UserID: userID, CreatedAt: time.Now()}
	return al.Trans.Exec(ctx, func(ctx context.Context) error {
		return al.ArticleLikeDAL.Create(ctx, like)
	})
}

// UnLike 取消点赞（幂等）
func (al *ArticleLike) UnLike(ctx context.Context, userID, articleID string) error {
	exists, err := al.ArticleLikeDAL.Exists(ctx, userID, articleID)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return al.Trans.Exec(ctx, func(ctx context.Context) error {
		return al.ArticleLikeDAL.Delete(ctx, userID, articleID)
	})
}

// GetLikeStatus 获取文章的点赞状态和总数
func (al *ArticleLike) GetLikeStatus(ctx context.Context, userID, articleID string) (*schema.ArticleLikeCountResult, error) {
	count, err := al.ArticleLikeDAL.CountByArticleID(ctx, articleID)
	if err != nil {
		return nil, err
	}
	liked := false
	if userID != "" {
		liked, err = al.ArticleLikeDAL.Exists(ctx, userID, articleID)
		if err != nil {
			return nil, err
		}
	}
	return &schema.ArticleLikeCountResult{Count: count, Liked: liked}, nil
}
