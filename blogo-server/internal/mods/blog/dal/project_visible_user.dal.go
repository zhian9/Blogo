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

func GetProjectVisibleUserDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ProjectVisibleUser))
}

type ProjectVisibleUser struct {
	DB *gorm.DB
}

func (pv *ProjectVisibleUser) BatchCreate(ctx context.Context, items []*schema.ProjectVisibleUser) error {
	if len(items) == 0 {
		return nil
	}
	result := GetProjectVisibleUserDB(ctx, pv.DB).Create(&items)
	return errors.WithStack(result.Error)
}

func (pv *ProjectVisibleUser) DeleteByProjectID(ctx context.Context, projectID string) error {
	result := GetProjectVisibleUserDB(ctx, pv.DB).Where("project_id = ?", projectID).Delete(new(schema.ProjectVisibleUser))
	return errors.WithStack(result.Error)
}

func (pv *ProjectVisibleUser) DeleteByProjectIDs(ctx context.Context, projectIDs []string) error {
	if len(projectIDs) == 0 {
		return nil
	}
	result := GetProjectVisibleUserDB(ctx, pv.DB).Where("project_id IN ?", projectIDs).Delete(new(schema.ProjectVisibleUser))
	return errors.WithStack(result.Error)
}

func (pv *ProjectVisibleUser) IsUserVisible(ctx context.Context, projectID, userID string) (bool, error) {
	ok, err := util.Exists(ctx, GetProjectVisibleUserDB(ctx, pv.DB).
		Where("project_id = ? AND user_id = ?", projectID, userID))
	return ok, errors.WithStack(err)
}

func (pv *ProjectVisibleUser) BatchGetByProjectIDs(ctx context.Context, projectIDs []string) (map[string][]*schema.ProjectVisibleUser, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	var list []*schema.ProjectVisibleUser
	if err := GetProjectVisibleUserDB(ctx, pv.DB).Where("project_id IN ?", projectIDs).Find(&list).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	result := make(map[string][]*schema.ProjectVisibleUser, len(projectIDs))
	for _, item := range list {
		result[item.ProjectID] = append(result[item.ProjectID], item)
	}
	return result, nil
}

func (pv *ProjectVisibleUser) ValidateUsersExist(ctx context.Context, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var existing []string
	err := util.GetDB(ctx, pv.DB).Model(new(schema.ProjectVisibleUser)).
		Distinct("user_id").Where("user_id IN ?", userIDs).Pluck("user_id", &existing).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	existSet := make(map[string]bool, len(existing))
	for _, id := range existing {
		existSet[id] = true
	}
	var missing []string
	for _, id := range userIDs {
		if !existSet[id] {
			missing = append(missing, id)
		}
	}
	return missing, nil
}
