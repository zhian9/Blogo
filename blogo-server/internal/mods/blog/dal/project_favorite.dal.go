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

func GetProjectFavoriteDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ProjectFavorite))
}

type ProjectFavorite struct {
	DB *gorm.DB
}

func (pf *ProjectFavorite) Create(ctx context.Context, fav *schema.ProjectFavorite) error {
	result := GetProjectFavoriteDB(ctx, pf.DB).Create(fav)
	return errors.WithStack(result.Error)
}

func (pf *ProjectFavorite) Delete(ctx context.Context, userID, projectID string) error {
	result := GetProjectFavoriteDB(ctx, pf.DB).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Delete(new(schema.ProjectFavorite))
	return errors.WithStack(result.Error)
}

func (pf *ProjectFavorite) Exists(ctx context.Context, userID, projectID string) (bool, error) {
	ok, err := util.Exists(ctx, GetProjectFavoriteDB(ctx, pf.DB).
		Where("user_id = ? AND project_id = ?", userID, projectID))
	return ok, errors.WithStack(err)
}

func (pf *ProjectFavorite) CountByProjectID(ctx context.Context, projectID string) (int64, error) {
	var count int64
	err := GetProjectFavoriteDB(ctx, pf.DB).
		Where("project_id = ?", projectID).
		Count(&count).Error
	return count, errors.WithStack(err)
}

func (pf *ProjectFavorite) DeleteByProjectIDs(ctx context.Context, projectIDs []string) error {
	if len(projectIDs) == 0 {
		return nil
	}
	result := GetProjectFavoriteDB(ctx, pf.DB).
		Where("project_id IN ?", projectIDs).
		Delete(new(schema.ProjectFavorite))
	return errors.WithStack(result.Error)
}
