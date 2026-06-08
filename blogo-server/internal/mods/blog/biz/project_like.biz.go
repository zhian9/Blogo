// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

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

type ProjectLike struct {
	Trans          *util.Trans
	ProjectLikeDAL *dal.ProjectLike
	ProjectDAL     *dal.Project
}

// Like 点赞项目（幂等）
func (pl *ProjectLike) Like(ctx context.Context, userID string, form *schema.ProjectLikeForm) error {
	project, err := pl.ProjectDAL.Get(ctx, form.ProjectID)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.BadRequest(config.ErrBadRequest, "Project not found")
	}
	if project.Status != schema.ProjectStatusPublished {
		return errors.BadRequest(config.ErrBadRequest, "Cannot like unpublished project")
	}
	exists, err := pl.ProjectLikeDAL.Exists(ctx, userID, form.ProjectID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	like := &schema.ProjectLike{ProjectID: form.ProjectID, UserID: userID, CreatedAt: time.Now()}
	return pl.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := pl.ProjectLikeDAL.Create(ctx, like); err != nil {
			return err
		}
		return pl.ProjectDAL.IncLikeCount(ctx, form.ProjectID, 1)
	})
}

// UnLike 取消点赞（幂等）
func (pl *ProjectLike) UnLike(ctx context.Context, userID, projectID string) error {
	exists, err := pl.ProjectLikeDAL.Exists(ctx, userID, projectID)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return pl.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := pl.ProjectLikeDAL.Delete(ctx, userID, projectID); err != nil {
			return err
		}
		return pl.ProjectDAL.IncLikeCount(ctx, projectID, -1)
	})
}

// GetLikeStatus 获取项目的点赞状态和总数
func (pl *ProjectLike) GetLikeStatus(ctx context.Context, userID, projectID string) (*schema.ProjectLikeCountResult, error) {
	count, err := pl.ProjectLikeDAL.CountByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	liked := false
	if userID != "" {
		liked, err = pl.ProjectLikeDAL.Exists(ctx, userID, projectID)
		if err != nil {
			return nil, err
		}
	}
	return &schema.ProjectLikeCountResult{Count: count, Liked: liked}, nil
}
