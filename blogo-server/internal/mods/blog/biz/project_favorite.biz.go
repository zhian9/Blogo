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

type ProjectFavorite struct {
	Trans              *util.Trans
	ProjectFavoriteDAL *dal.ProjectFavorite
	ProjectDAL         *dal.Project
}

// Create 收藏项目（幂等）
func (pf *ProjectFavorite) Create(ctx context.Context, userID string, form *schema.ProjectFavoriteForm) error {
	project, err := pf.ProjectDAL.Get(ctx, form.ProjectID)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.BadRequest(config.ErrBadRequest, "Project not found")
	}
	exists, err := pf.ProjectFavoriteDAL.Exists(ctx, userID, form.ProjectID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	fav := &schema.ProjectFavorite{ProjectID: form.ProjectID, UserID: userID, CreatedAt: time.Now()}
	return pf.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := pf.ProjectFavoriteDAL.Create(ctx, fav); err != nil {
			return err
		}
		return pf.ProjectDAL.IncFavoriteCount(ctx, form.ProjectID, 1)
	})
}

// Delete 取消收藏（幂等）
func (pf *ProjectFavorite) Delete(ctx context.Context, userID, projectID string) error {
	exists, err := pf.ProjectFavoriteDAL.Exists(ctx, userID, projectID)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return pf.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := pf.ProjectFavoriteDAL.Delete(ctx, userID, projectID); err != nil {
			return err
		}
		return pf.ProjectDAL.IncFavoriteCount(ctx, projectID, -1)
	})
}

// IsFavorite 检查是否已收藏
func (pf *ProjectFavorite) IsFavorite(ctx context.Context, userID, projectID string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	return pf.ProjectFavoriteDAL.Exists(ctx, userID, projectID)
}

// Count 获取项目收藏数
func (pf *ProjectFavorite) Count(ctx context.Context, projectID string) (int64, error) {
	return pf.ProjectFavoriteDAL.CountByProjectID(ctx, projectID)
}
