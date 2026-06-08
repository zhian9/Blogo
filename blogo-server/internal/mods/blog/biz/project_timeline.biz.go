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

type ProjectTimeline struct {
	Trans              *util.Trans
	ProjectTimelineDAL *dal.ProjectTimeline
	ProjectDAL         *dal.Project
}

// GetByProjectID 获取项目的时间线
func (pt *ProjectTimeline) GetByProjectID(ctx context.Context, projectID string) (schema.ProjectTimelines, error) {
	return pt.ProjectTimelineDAL.GetByProjectID(ctx, projectID)
}

// Get 获取单个时间线条目
func (pt *ProjectTimeline) Get(ctx context.Context, id string) (*schema.ProjectTimeline, error) {
	tl, err := pt.ProjectTimelineDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if tl == nil {
		return nil, errors.NotFound(config.ErrNotFound, "Timeline entry not found")
	}
	return tl, nil
}

// Create 创建时间线条目
func (pt *ProjectTimeline) Create(ctx context.Context, projectID string, form *schema.ProjectTimelineForm) (*schema.ProjectTimeline, error) {
	// 校验项目存在
	project, err := pt.ProjectDAL.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.NotFound(config.ErrNotFound, "Project not found")
	}

	uid := userID(ctx)
	if project.AuthorID != uid && !util.FromIsRootUser(ctx) {
		return nil, errors.Forbidden("", "您无权编辑此项目的历程")
	}

	tl := &schema.ProjectTimeline{
		ID:        util.NewXID(),
		ProjectID: projectID,
		CreatedAt: time.Now(),
	}
	if err := form.FillTo(tl); err != nil {
		return nil, err
	}

	if err := pt.ProjectTimelineDAL.Create(ctx, tl); err != nil {
		return nil, err
	}
	return tl, nil
}

// Update 更新时间线条目
func (pt *ProjectTimeline) Update(ctx context.Context, id string, form *schema.ProjectTimelineForm) error {
	tl, err := pt.ProjectTimelineDAL.Get(ctx, id)
	if err != nil {
		return err
	}
	if tl == nil {
		return errors.NotFound(config.ErrNotFound, "Timeline entry not found")
	}

	project, err := pt.ProjectDAL.Get(ctx, tl.ProjectID)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.NotFound(config.ErrNotFound, "Project not found")
	}

	uid := userID(ctx)
	if project.AuthorID != uid && !util.FromIsRootUser(ctx) {
		return errors.Forbidden("", "您无权编辑此项目的历程")
	}

	if err := form.FillTo(tl); err != nil {
		return err
	}
	tl.UpdatedAt = time.Now()

	return pt.ProjectTimelineDAL.Update(ctx, tl)
}

// Delete 删除时间线条目
func (pt *ProjectTimeline) Delete(ctx context.Context, id string) error {
	tl, err := pt.ProjectTimelineDAL.Get(ctx, id)
	if err != nil {
		return err
	}
	if tl == nil {
		return errors.NotFound(config.ErrNotFound, "Timeline entry not found")
	}

	project, err := pt.ProjectDAL.Get(ctx, tl.ProjectID)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.NotFound(config.ErrNotFound, "Project not found")
	}

	uid := userID(ctx)
	if project.AuthorID != uid && !util.FromIsRootUser(ctx) {
		return errors.Forbidden("", "您无权删除此项目的历程")
	}

	return pt.ProjectTimelineDAL.Delete(ctx, id)
}
