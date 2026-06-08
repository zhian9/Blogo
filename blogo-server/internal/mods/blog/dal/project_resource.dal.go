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

func GetProjectResourceDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ProjectResource))
}

type ProjectResource struct {
	DB *gorm.DB
}

func (pr *ProjectResource) Get(ctx context.Context, id string) (*schema.ProjectResource, error) {
	res := new(schema.ProjectResource)
	ok, err := util.FindOne(ctx, GetProjectResourceDB(ctx, pr.DB).Where("id = ?", id), util.QueryOptions{}, res)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return res, nil
}

func (pr *ProjectResource) GetByProjectID(ctx context.Context, projectID string) (schema.ProjectResources, error) {
	var list schema.ProjectResources
	if err := GetProjectResourceDB(ctx, pr.DB).Where("project_id = ?", projectID).Order("sort_order ASC").Find(&list).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return list, nil
}

func (pr *ProjectResource) Create(ctx context.Context, res *schema.ProjectResource) error {
	result := GetProjectResourceDB(ctx, pr.DB).Create(res)
	return errors.WithStack(result.Error)
}

func (pr *ProjectResource) Update(ctx context.Context, res *schema.ProjectResource, selectFields ...string) error {
	db := GetProjectResourceDB(ctx, pr.DB).Where("id = ?", res.ID)
	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("id", "created_at")
	}
	result := db.Updates(res)
	return errors.WithStack(result.Error)
}

func (pr *ProjectResource) Delete(ctx context.Context, id string) error {
	result := GetProjectResourceDB(ctx, pr.DB).Where("id = ?", id).Delete(new(schema.ProjectResource))
	return errors.WithStack(result.Error)
}

func (pr *ProjectResource) DeleteByProjectID(ctx context.Context, projectID string) error {
	result := GetProjectResourceDB(ctx, pr.DB).Where("project_id = ?", projectID).Delete(new(schema.ProjectResource))
	return errors.WithStack(result.Error)
}

func (pr *ProjectResource) DeleteByProjectIDs(ctx context.Context, projectIDs []string) error {
	if len(projectIDs) == 0 {
		return nil
	}
	result := GetProjectResourceDB(ctx, pr.DB).Where("project_id IN ?", projectIDs).Delete(new(schema.ProjectResource))
	return errors.WithStack(result.Error)
}
