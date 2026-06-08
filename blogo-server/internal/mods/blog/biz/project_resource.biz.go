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

type ProjectResource struct {
	Trans              *util.Trans
	ProjectResourceDAL *dal.ProjectResource
	ProjectDAL         *dal.Project
}

// GetByProjectID 获取项目的相关资源
func (pr *ProjectResource) GetByProjectID(ctx context.Context, projectID string) (schema.ProjectResources, error) {
	return pr.ProjectResourceDAL.GetByProjectID(ctx, projectID)
}

// Create 创建资源
func (pr *ProjectResource) Create(ctx context.Context, projectID string, form *schema.ProjectResourceForm) (*schema.ProjectResource, error) {
	project, err := pr.ProjectDAL.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.NotFound(config.ErrNotFound, "Project not found")
	}

	uid := userID(ctx)
	if project.AuthorID != uid && !util.FromIsRootUser(ctx) {
		return nil, errors.Forbidden("", "您无权编辑此项目的资源")
	}

	res := &schema.ProjectResource{
		ID:        util.NewXID(),
		ProjectID: projectID,
		CreatedAt: time.Now(),
	}
	if err := form.FillTo(res); err != nil {
		return nil, err
	}

	if err := pr.ProjectResourceDAL.Create(ctx, res); err != nil {
		return nil, err
	}
	return res, nil
}

// Update 更新资源
func (pr *ProjectResource) Update(ctx context.Context, id string, form *schema.ProjectResourceForm) error {
	target, err := pr.ProjectResourceDAL.Get(ctx, id)
	if err != nil {
		return err
	}
	if target == nil {
		return errors.NotFound(config.ErrNotFound, "Resource not found")
	}

	project, err := pr.ProjectDAL.Get(ctx, target.ProjectID)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.NotFound(config.ErrNotFound, "Project not found")
	}

	uid := userID(ctx)
	if project.AuthorID != uid && !util.FromIsRootUser(ctx) {
		return errors.Forbidden("", "您无权编辑此项目的资源")
	}

	if err := form.FillTo(target); err != nil {
		return err
	}
	target.UpdatedAt = time.Now()

	return pr.ProjectResourceDAL.Update(ctx, target)
}

// Delete 删除资源
func (pr *ProjectResource) Delete(ctx context.Context, projectID, id string) error {
	project, err := pr.ProjectDAL.Get(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.NotFound(config.ErrNotFound, "Project not found")
	}

	uid := userID(ctx)
	if project.AuthorID != uid && !util.FromIsRootUser(ctx) {
		return errors.Forbidden("", "您无权删除此项目的资源")
	}

	return pr.ProjectResourceDAL.Delete(ctx, id)
}
