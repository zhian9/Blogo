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

func GetProjectLikeDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ProjectLike))
}

type ProjectLike struct {
	DB *gorm.DB
}

func (pl *ProjectLike) Create(ctx context.Context, like *schema.ProjectLike) error {
	result := GetProjectLikeDB(ctx, pl.DB).Create(like)
	return errors.WithStack(result.Error)
}

func (pl *ProjectLike) Delete(ctx context.Context, userID, projectID string) error {
	result := GetProjectLikeDB(ctx, pl.DB).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Delete(new(schema.ProjectLike))
	return errors.WithStack(result.Error)
}

func (pl *ProjectLike) Exists(ctx context.Context, userID, projectID string) (bool, error) {
	ok, err := util.Exists(ctx, GetProjectLikeDB(ctx, pl.DB).
		Where("user_id = ? AND project_id = ?", userID, projectID))
	return ok, errors.WithStack(err)
}

func (pl *ProjectLike) CountByProjectID(ctx context.Context, projectID string) (int64, error) {
	var count int64
	err := GetProjectLikeDB(ctx, pl.DB).
		Where("project_id = ?", projectID).
		Count(&count).Error
	return count, errors.WithStack(err)
}

func (pl *ProjectLike) DeleteByProjectIDs(ctx context.Context, projectIDs []string) error {
	if len(projectIDs) == 0 {
		return nil
	}
	result := GetProjectLikeDB(ctx, pl.DB).
		Where("project_id IN ?", projectIDs).
		Delete(new(schema.ProjectLike))
	return errors.WithStack(result.Error)
}
