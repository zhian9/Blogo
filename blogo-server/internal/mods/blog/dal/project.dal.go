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
	rschema "github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

func GetProjectDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Project))
}

type Project struct {
	DB *gorm.DB
}

func (p *Project) loadCoverImages(ctx context.Context, projects schema.Projects) error {
	var ids []string
	for _, proj := range projects {
		if proj.CoverImageID != nil && *proj.CoverImageID != "" {
			ids = append(ids, *proj.CoverImageID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var images []schema.Image
	if err := util.GetDB(ctx, p.DB).Model(new(schema.Image)).Where("id IN ?", ids).Find(&images).Error; err != nil {
		return errors.WithStack(err)
	}
	imgMap := make(map[string]*schema.Image, len(images))
	for i := range images {
		imgMap[images[i].ID] = &images[i]
	}
	for _, proj := range projects {
		if proj.CoverImageID != nil {
			proj.CoverImage = imgMap[*proj.CoverImageID]
		}
	}
	return nil
}

func (p *Project) loadAuthors(ctx context.Context, projects schema.Projects) error {
	var ids []string
	for _, proj := range projects {
		if proj.AuthorID != "" {
			ids = append(ids, proj.AuthorID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var users []rschema.User
	if err := util.GetDB(ctx, p.DB).Model(new(rschema.User)).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return errors.WithStack(err)
	}
	userMap := make(map[string]*rschema.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}
	for _, proj := range projects {
		if proj.AuthorID != "" {
			proj.Author = userMap[proj.AuthorID]
		}
	}
	return nil
}

func (p *Project) loadVisibleUsers(ctx context.Context, projects schema.Projects) error {
	var ids []string
	for _, proj := range projects {
		if proj.Visibility == schema.ProjectVisibilityPartialVisible {
			ids = append(ids, proj.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var vuList []*schema.ProjectVisibleUser
	if err := util.GetDB(ctx, p.DB).Model(new(schema.ProjectVisibleUser)).Where("project_id IN ?", ids).Find(&vuList).Error; err != nil {
		return errors.WithStack(err)
	}
	vuMap := make(map[string][]*schema.ProjectVisibleUser, len(ids))
	for _, vu := range vuList {
		vuMap[vu.ProjectID] = append(vuMap[vu.ProjectID], vu)
	}
	for _, proj := range projects {
		if proj.Visibility == schema.ProjectVisibilityPartialVisible {
			proj.VisibleUsers = vuMap[proj.ID]
		}
	}
	return nil
}

func (p *Project) loadTimelines(ctx context.Context, projects schema.Projects) error {
	var ids []string
	for _, proj := range projects {
		ids = append(ids, proj.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	var timelines []*schema.ProjectTimeline
	if err := util.GetDB(ctx, p.DB).Model(new(schema.ProjectTimeline)).Where("project_id IN ?", ids).Order("event_date DESC").Find(&timelines).Error; err != nil {
		return errors.WithStack(err)
	}
	tlMap := make(map[string][]*schema.ProjectTimeline, len(ids))
	for _, tl := range timelines {
		tlMap[tl.ProjectID] = append(tlMap[tl.ProjectID], tl)
	}
	for _, proj := range projects {
		proj.Timeline = tlMap[proj.ID]
	}
	return nil
}

func (p *Project) loadResources(ctx context.Context, projects schema.Projects) error {
	var ids []string
	for _, proj := range projects {
		ids = append(ids, proj.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	var resources []*schema.ProjectResource
	if err := util.GetDB(ctx, p.DB).Model(new(schema.ProjectResource)).Where("project_id IN ?", ids).Order("sort_order ASC").Find(&resources).Error; err != nil {
		return errors.WithStack(err)
	}
	resMap := make(map[string][]*schema.ProjectResource, len(ids))
	for _, res := range resources {
		resMap[res.ProjectID] = append(resMap[res.ProjectID], res)
	}
	for _, proj := range projects {
		proj.Resources = resMap[proj.ID]
	}
	return nil
}

func (p *Project) Query(ctx context.Context, params schema.ProjectQueryParam, opts ...schema.ProjectQueryOptions) (*schema.ProjectQueryResult, error) {
	var opt schema.ProjectQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetProjectDB(ctx, p.DB)

	if opt.UserID != "" {
		visibleSubQuery := p.DB.Model(&schema.ProjectVisibleUser{}).
			Select("project_id").
			Where("user_id = ?", opt.UserID)
		db = db.Where(
			"(status = ? AND visibility = ?) OR author_id = ? OR (status = ? AND visibility = ? AND id IN (?))",
			schema.ProjectStatusPublished, schema.ProjectVisibilityPublic,
			opt.UserID,
			schema.ProjectStatusPublished, schema.ProjectVisibilityPartialVisible,
			visibleSubQuery,
		)
	} else {
		db = db.Where("status = ? AND visibility = ?",
			schema.ProjectStatusPublished, schema.ProjectVisibilityPublic)
	}

	if v := params.Title; len(v) > 0 {
		db = db.Where("title LIKE ?", "%"+v+"%")
	}
	if v := params.CategoryID; len(v) > 0 {
		db = db.Where("category_id = ?", v)
	}
	if v := params.TagName; len(v) > 0 {
		subQuery := p.DB.Model(&schema.ProjectTag{}).
			Select("project_id").
			Where("tag_id IN (?)",
				p.DB.Model(&schema.Tag{}).Select("id").Where("name = ?", v),
			)
		db = db.Where("id IN (?)", subQuery)
	}
	if v := params.AuthorID; len(v) > 0 {
		db = db.Where("author_id = ?", v)
	}
	if v := params.Status; len(v) > 0 {
		db = db.Where("status = ?", v)
	}
	if v := params.Visibility; len(v) > 0 {
		db = db.Where("visibility = ?", v)
	}
	if v := params.ProjectState; len(v) > 0 {
		db = db.Where("project_state = ?", v)
	}
	if v := params.IsTop; v != nil {
		db = db.Where("is_top = ?", *v)
	}
	if v := params.IsFeatured; v != nil {
		db = db.Where("is_featured = ?", *v)
	}
	if v := params.PublishedAtGte; v != nil {
		db = db.Where("published_at >= ?", *v)
	}
	if v := params.PublishedAtLte; v != nil {
		db = db.Where("published_at <= ?", *v)
	}

	if opt.WithCategory {
		db = db.Preload("Category")
	}
	if opt.WithTags {
		db = db.Preload("Tags")
	}

	var list schema.Projects
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if opt.WithCoverImage && len(list) > 0 {
		if err := p.loadCoverImages(ctx, list); err != nil {
			return nil, err
		}
	}
	if opt.WithAuthor && len(list) > 0 {
		if err := p.loadAuthors(ctx, list); err != nil {
			return nil, err
		}
	}
	if opt.WithVisibleUsers && len(list) > 0 {
		if err := p.loadVisibleUsers(ctx, list); err != nil {
			return nil, err
		}
	}

	return &schema.ProjectQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

func (p *Project) Get(ctx context.Context, id string, opts ...schema.ProjectQueryOptions) (*schema.Project, error) {
	var opt schema.ProjectQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	project := new(schema.Project)
	db := GetProjectDB(ctx, p.DB).Where("id = ?", id)
	if opt.WithCategory {
		db = db.Preload("Category")
	}
	if opt.WithTags {
		db = db.Preload("Tags")
	}

	ok, err := util.FindOne(ctx, db, opt.QueryOptions, project)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}

	if opt.WithCoverImage {
		if err := p.loadCoverImages(ctx, schema.Projects{project}); err != nil {
			return nil, err
		}
	}
	if opt.WithAuthor {
		if err := p.loadAuthors(ctx, schema.Projects{project}); err != nil {
			return nil, err
		}
	}
	if opt.WithVisibleUsers {
		if err := p.loadVisibleUsers(ctx, schema.Projects{project}); err != nil {
			return nil, err
		}
	}
	if opt.WithTimeline {
		if err := p.loadTimelines(ctx, schema.Projects{project}); err != nil {
			return nil, err
		}
	}
	if opt.WithResources {
		if err := p.loadResources(ctx, schema.Projects{project}); err != nil {
			return nil, err
		}
	}
	return project, nil
}

func (p *Project) GetBySlug(ctx context.Context, slug string, opts ...schema.ProjectQueryOptions) (*schema.Project, error) {
	var opt schema.ProjectQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	project := new(schema.Project)
	db := GetProjectDB(ctx, p.DB).Where("slug = ?", slug)
	if opt.UserID == "" {
		db = db.Where("status = ? AND visibility = ?", schema.ProjectStatusPublished, schema.ProjectVisibilityPublic)
	} else {
		visibleSubQuery := p.DB.Model(&schema.ProjectVisibleUser{}).
			Select("project_id").
			Where("user_id = ?", opt.UserID)
		db = db.Where(
			"(status = ? AND visibility = ?) OR author_id = ? OR (status = ? AND visibility = ? AND id IN (?))",
			schema.ProjectStatusPublished, schema.ProjectVisibilityPublic,
			opt.UserID,
			schema.ProjectStatusPublished, schema.ProjectVisibilityPartialVisible,
			visibleSubQuery,
		)
	}
	if opt.WithCategory {
		db = db.Preload("Category")
	}
	if opt.WithTags {
		db = db.Preload("Tags")
	}

	ok, err := util.FindOne(ctx, db, opt.QueryOptions, project)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}

	if opt.WithCoverImage {
		if err := p.loadCoverImages(ctx, schema.Projects{project}); err != nil {
			return nil, err
		}
	}
	if opt.WithAuthor {
		if err := p.loadAuthors(ctx, schema.Projects{project}); err != nil {
			return nil, err
		}
	}
	if opt.WithVisibleUsers {
		if err := p.loadVisibleUsers(ctx, schema.Projects{project}); err != nil {
			return nil, err
		}
	}
	if opt.WithTimeline {
		if err := p.loadTimelines(ctx, schema.Projects{project}); err != nil {
			return nil, err
		}
	}
	if opt.WithResources {
		if err := p.loadResources(ctx, schema.Projects{project}); err != nil {
			return nil, err
		}
	}
	return project, nil
}

func (p *Project) ExistsID(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetProjectDB(ctx, p.DB).Where("id = ?", id))
	return ok, errors.WithStack(err)
}
func (p *Project) ExistsSlug(ctx context.Context, slug string) (bool, error) {
	ok, err := util.Exists(ctx, GetProjectDB(ctx, p.DB).Where("slug = ?", slug))
	return ok, errors.WithStack(err)
}
func (p *Project) Create(ctx context.Context, project *schema.Project) error {
	result := GetProjectDB(ctx, p.DB).Create(project)
	return errors.WithStack(result.Error)
}
func (p *Project) Update(ctx context.Context, project *schema.Project, selectFields ...string) error {
	db := GetProjectDB(ctx, p.DB).Where("id = ?", project.ID)
	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("id", "created_at")
	}
	result := db.Updates(project)
	return errors.WithStack(result.Error)
}
func (p *Project) Delete(ctx context.Context, id string) error {
	result := GetProjectDB(ctx, p.DB).Where("id = ?", id).Delete(new(schema.Project))
	return errors.WithStack(result.Error)
}
func (p *Project) DeleteByIds(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	result := GetProjectDB(ctx, p.DB).Where("id IN ?", ids).Delete(new(schema.Project))
	return errors.WithStack(result.Error)
}
func (p *Project) UpdateStatus(ctx context.Context, ids []string, status string) error {
	if len(ids) == 0 {
		return nil
	}
	result := GetProjectDB(ctx, p.DB).Where("id IN ?", ids).Update("status", status)
	return errors.WithStack(result.Error)
}
func (p *Project) IncViews(ctx context.Context, id string, num int64) error {
	result := GetProjectDB(ctx, p.DB).Where("id = ?", id).Update("views", gorm.Expr("views + ?", num))
	return errors.WithStack(result.Error)
}
func (p *Project) IncLikeCount(ctx context.Context, id string, delta int64) error {
	result := GetProjectDB(ctx, p.DB).Where("id = ?", id).Update("like_count", gorm.Expr("like_count + ?", delta))
	return errors.WithStack(result.Error)
}
func (p *Project) IncFavoriteCount(ctx context.Context, id string, delta int64) error {
	result := GetProjectDB(ctx, p.DB).Where("id = ?", id).Update("favorite_count", gorm.Expr("favorite_count + ?", delta))
	return errors.WithStack(result.Error)
}
func (p *Project) IncCommentCount(ctx context.Context, id string, delta int64) error {
	result := GetProjectDB(ctx, p.DB).Where("id = ?", id).Update("comment_count", gorm.Expr("comment_count + ?", delta))
	return errors.WithStack(result.Error)
}
func (p *Project) CountByCategoryID(ctx context.Context, categoryID string) (int64, error) {
	if categoryID == "" {
		return 0, nil
	}
	var count int64
	err := GetProjectDB(ctx, p.DB).Where("category_id = ?", categoryID).Count(&count).Error
	return count, errors.WithStack(err)
}
