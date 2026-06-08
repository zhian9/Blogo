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

func GetProjectTimelineDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ProjectTimeline))
}

type ProjectTimeline struct {
	DB *gorm.DB
}

func (pt *ProjectTimeline) Query(ctx context.Context, params schema.ProjectTimelineQueryParam, opts ...util.QueryOptions) (*schema.ProjectTimelineQueryResult, error) {
	var opt util.QueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetProjectTimelineDB(ctx, pt.DB)

	if v := params.ProjectID; len(v) > 0 {
		db = db.Where("project_id = ?", v)
	}
	if v := params.Type; len(v) > 0 {
		db = db.Where("type = ?", v)
	}

	var list schema.ProjectTimelines
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schema.ProjectTimelineQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

func (pt *ProjectTimeline) Get(ctx context.Context, id string) (*schema.ProjectTimeline, error) {
	tl := new(schema.ProjectTimeline)
	ok, err := util.FindOne(ctx, GetProjectTimelineDB(ctx, pt.DB).Where("id = ?", id), util.QueryOptions{}, tl)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return tl, nil
}

func (pt *ProjectTimeline) GetByProjectID(ctx context.Context, projectID string) (schema.ProjectTimelines, error) {
	var list schema.ProjectTimelines
	if err := GetProjectTimelineDB(ctx, pt.DB).Where("project_id = ?", projectID).Order("event_date DESC").Find(&list).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return list, nil
}

func (pt *ProjectTimeline) Create(ctx context.Context, tl *schema.ProjectTimeline) error {
	result := GetProjectTimelineDB(ctx, pt.DB).Create(tl)
	return errors.WithStack(result.Error)
}

func (pt *ProjectTimeline) Update(ctx context.Context, tl *schema.ProjectTimeline, selectFields ...string) error {
	db := GetProjectTimelineDB(ctx, pt.DB).Where("id = ?", tl.ID)
	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("id", "created_at")
	}
	result := db.Updates(tl)
	return errors.WithStack(result.Error)
}

func (pt *ProjectTimeline) Delete(ctx context.Context, id string) error {
	result := GetProjectTimelineDB(ctx, pt.DB).Where("id = ?", id).Delete(new(schema.ProjectTimeline))
	return errors.WithStack(result.Error)
}

func (pt *ProjectTimeline) DeleteByProjectID(ctx context.Context, projectID string) error {
	result := GetProjectTimelineDB(ctx, pt.DB).Where("project_id = ?", projectID).Delete(new(schema.ProjectTimeline))
	return errors.WithStack(result.Error)
}

func (pt *ProjectTimeline) DeleteByProjectIDs(ctx context.Context, projectIDs []string) error {
	if len(projectIDs) == 0 {
		return nil
	}
	result := GetProjectTimelineDB(ctx, pt.DB).Where("project_id IN ?", projectIDs).Delete(new(schema.ProjectTimeline))
	return errors.WithStack(result.Error)
}
