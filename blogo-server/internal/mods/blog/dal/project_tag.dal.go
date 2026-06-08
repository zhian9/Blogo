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

func GetProjectTagDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ProjectTag))
}

type ProjectTag struct {
	DB *gorm.DB
}

func (pt *ProjectTag) CreateBatch(ctx context.Context, items []schema.ProjectTag) error {
	if len(items) == 0 {
		return nil
	}
	result := GetProjectTagDB(ctx, pt.DB).Create(&items)
	return errors.WithStack(result.Error)
}

func (pt *ProjectTag) DeleteByProjectID(ctx context.Context, projectID string) error {
	result := GetProjectTagDB(ctx, pt.DB).Where("project_id = ?", projectID).Delete(new(schema.ProjectTag))
	return errors.WithStack(result.Error)
}

func (pt *ProjectTag) DeleteByProjectIDs(ctx context.Context, projectIDs []string) error {
	if len(projectIDs) == 0 {
		return nil
	}
	result := GetProjectTagDB(ctx, pt.DB).Where("project_id IN ?", projectIDs).Delete(new(schema.ProjectTag))
	return errors.WithStack(result.Error)
}
