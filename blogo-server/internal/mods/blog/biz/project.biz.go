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
	rschema "github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// Project 是项目管理业务的核心对象
type Project struct {
	Trans                 *util.Trans
	ProjectDAL            *dal.Project
	CategoryDAL           *dal.Category
	TagDAL                *dal.Tag
	ProjectTagDAL         *dal.ProjectTag
	CommentDAL            *dal.Comment
	ProjectVisibleUserDAL *dal.ProjectVisibleUser
	ProjectTimelineDAL    *dal.ProjectTimeline
	ProjectResourceDAL    *dal.ProjectResource
	ProjectLikeDAL        *dal.ProjectLike
	ProjectFavoriteDAL    *dal.ProjectFavorite
	ContributionDAL       *dal.Contribution
	ProjectPermission     *ProjectPermission
}

func (p *Project) hasRoleCode(ctx context.Context, codes ...string) bool {
	if util.FromIsRootUser(ctx) {
		return true
	}
	userCache := util.FromUserCache(ctx)
	if len(userCache.RoleIDs) == 0 {
		return false
	}
	if len(userCache.RoleCodes) > 0 {
		for _, rc := range userCache.RoleCodes {
			for _, c := range codes {
				if rc == c {
					return true
				}
			}
		}
		return false
	}
	var count int64
	p.Trans.DB.Model(&rschema.Role{}).
		Where("id IN ? AND code IN ?", userCache.RoleIDs, codes).
		Count(&count)
	return count > 0
}

func (p *Project) canSetTop(ctx context.Context) bool {
	return p.hasRoleCode(ctx, "super_admin", "admin", "content_manager")
}

func (p *Project) canSetFeatured(ctx context.Context) bool {
	return p.hasRoleCode(ctx, "super_admin", "admin")
}

// Query 查询项目列表
func (p *Project) Query(ctx context.Context, params schema.ProjectQueryParam) (*schema.ProjectQueryResult, error) {
	params.Pagination = true

	// 排序
	orderFields := []util.OrderByParam{
		{Field: "is_top", Direction: util.DESC},
	}
	switch params.SortBy {
	case "hot":
		orderFields = append(orderFields, util.OrderByParam{Field: "views", Direction: util.DESC})
	case "most_liked":
		orderFields = append(orderFields, util.OrderByParam{Field: "like_count", Direction: util.DESC})
	default:
		orderFields = append(orderFields, util.OrderByParam{Field: "published_at", Direction: util.DESC})
	}

	result, err := p.ProjectDAL.Query(ctx, params, schema.ProjectQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: orderFields,
		},
		WithCategory:     true,
		WithTags:         true,
		WithCoverImage:   true,
		WithAuthor:       true,
		WithVisibleUsers: true,
		UserID:           userID(ctx),
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get 获取单个项目（含权限校验）
func (p *Project) Get(ctx context.Context, id string) (*schema.Project, error) {
	project, err := p.ProjectDAL.Get(ctx, id, schema.ProjectQueryOptions{
		WithCategory:     true,
		WithTags:         true,
		WithCoverImage:   true,
		WithAuthor:       true,
		WithVisibleUsers: true,
		WithTimeline:     true,
		WithResources:    true,
		UserID:           userID(ctx),
	})
	if err != nil {
		return nil, err
	} else if project == nil {
		return nil, errors.NotFound(config.ErrNotFound, "Project not found")
	}
	if !p.ProjectPermission.CanUserReadProject(ctx, userID(ctx), project) {
		return nil, errors.Forbidden(config.ErrProjectPermissionDenied, "无权限访问该项目")
	}
	return project, nil
}

// GetBySlug 根据 Slug 获取项目
func (p *Project) GetBySlug(ctx context.Context, slug string) (*schema.Project, error) {
	project, err := p.ProjectDAL.GetBySlug(ctx, slug, schema.ProjectQueryOptions{
		WithCategory:     true,
		WithTags:         true,
		WithCoverImage:   true,
		WithAuthor:       true,
		WithVisibleUsers: true,
		WithTimeline:     true,
		WithResources:    true,
		UserID:           userID(ctx),
	})
	if err != nil {
		return nil, err
	} else if project == nil {
		return nil, errors.NotFound(config.ErrNotFound, "Project not found or not published")
	}
	if !p.ProjectPermission.CanUserReadProject(ctx, userID(ctx), project) {
		return nil, errors.Forbidden(config.ErrProjectPermissionDenied, "无权限访问该项目")
	}
	return project, nil
}

// GetFeatured 获取精选项目列表
func (p *Project) GetFeatured(ctx context.Context) (schema.Projects, error) {
	isFeatured := true
	result, err := p.ProjectDAL.Query(ctx, schema.ProjectQueryParam{
		PaginationParam: util.PaginationParam{Current: 1, PageSize: 10},
		IsFeatured:      &isFeatured,
		Status:          schema.ProjectStatusPublished,
	}, schema.ProjectQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "featured_order", Direction: util.ASC},
				{Field: "published_at", Direction: util.DESC},
			},
		},
		WithCategory:   true,
		WithTags:       true,
		WithCoverImage: true,
		WithAuthor:     true,
		UserID:         userID(ctx),
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// Create 创建新项目
func (p *Project) Create(ctx context.Context, projectForm *schema.ProjectForm) (*schema.Project, error) {
	if p.hasRoleCode(ctx, "guest") && !util.FromIsRootUser(ctx) {
		return nil, errors.Forbidden("", "游客无权发布项目，请先登录或联系管理员")
	}

	existsSlug, err := p.ProjectDAL.ExistsSlug(ctx, projectForm.Slug)
	if err != nil {
		return nil, err
	} else if existsSlug {
		return nil, errors.BadRequest(config.ErrBadRequest, "Slug already exists")
	}

	validUserIDs, err := p.validateVisibleUsers(ctx, projectForm)
	if err != nil {
		return nil, err
	}

	categoryID := projectForm.CategoryID
	if categoryID != "" {
		exists, err := p.CategoryDAL.ExistsID(ctx, categoryID)
		if err != nil {
			return nil, err
		}
		if !exists {
			categoryID, err = p.CategoryDAL.FindOrCreateByName(ctx, categoryID)
			if err != nil {
				return nil, err
			}
		}
	}
	projectForm.CategoryID = categoryID

	var tagIDs []string
	if len(projectForm.TagIDs) > 0 {
		existingTags, notExistsNames, err := p.TagDAL.GetByNames(ctx, projectForm.TagIDs)
		if err != nil {
			return nil, err
		}
		for _, tag := range existingTags {
			tagIDs = append(tagIDs, tag.ID)
		}
		for _, name := range notExistsNames {
			newTag := &schema.Tag{
				ID:        util.NewXID(),
				Name:      name,
				CreatedAt: time.Now(),
			}
			if err := p.TagDAL.Create(ctx, newTag); err != nil {
				return nil, err
			}
			tagIDs = append(tagIDs, newTag.ID)
		}
	}

	project := &schema.Project{
		ID:        util.NewXID(),
		AuthorID:  userID(ctx),
		CreatedAt: time.Now(),
	}

	if !p.canSetTop(ctx) {
		projectForm.IsTop = false
	}
	if !p.canSetFeatured(ctx) {
		projectForm.IsFeatured = false
	}

	if err := projectForm.FillTo(project); err != nil {
		return nil, err
	}

	err = p.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := p.ProjectDAL.Create(ctx, project); err != nil {
			return err
		}
		var projectTags []schema.ProjectTag
		for _, tagID := range tagIDs {
			projectTags = append(projectTags, schema.ProjectTag{
				ProjectID: project.ID,
				TagID:     tagID,
				CreatedAt: time.Now(),
			})
		}
		if len(projectTags) > 0 {
			if err := p.ProjectTagDAL.CreateBatch(ctx, projectTags); err != nil {
				return err
			}
		}
		if len(validUserIDs) > 0 {
			var vuList []*schema.ProjectVisibleUser
			for _, uid := range validUserIDs {
				vuList = append(vuList, &schema.ProjectVisibleUser{
					ID:        util.NewXID(),
					ProjectID: project.ID,
					UserID:    uid,
					CreatedAt: time.Now(),
				})
			}
			if err := p.ProjectVisibleUserDAL.BatchCreate(ctx, vuList); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if projectForm.Status == schema.ProjectStatusPublished {
		_ = p.ContributionDAL.RecordPublish(ctx, project.AuthorID)
	}

	return p.Get(ctx, project.ID)
}

// Update 更新项目
func (p *Project) Update(ctx context.Context, id string, projectForm *schema.ProjectForm) error {
	project, err := p.ProjectDAL.Get(ctx, id, schema.ProjectQueryOptions{
		UserID: userID(ctx),
	})
	if err != nil {
		return err
	} else if project == nil {
		return errors.NotFound(config.ErrNotFound, "Project not found")
	}

	uid := userID(ctx)
	if project.AuthorID != "" && project.AuthorID != uid && !util.FromIsRootUser(ctx) {
		return errors.Forbidden("", "您无权编辑他人发布的项目")
	}

	if project.Slug != projectForm.Slug {
		existsSlug, err := p.ProjectDAL.ExistsSlug(ctx, projectForm.Slug)
		if err != nil {
			return err
		} else if existsSlug {
			return errors.BadRequest(config.ErrBadRequest, "Slug already exists")
		}
	}

	validUserIDs, err := p.validateVisibleUsers(ctx, projectForm)
	if err != nil {
		return err
	}

	if projectForm.CategoryID != "" {
		exists, err := p.CategoryDAL.ExistsID(ctx, projectForm.CategoryID)
		if err != nil {
			return err
		}
		if !exists {
			categoryID, err := p.CategoryDAL.FindOrCreateByName(ctx, projectForm.CategoryID)
			if err != nil {
				return err
			}
			projectForm.CategoryID = categoryID
		}
	}

	var tagIDs []string
	if len(projectForm.TagIDs) > 0 {
		existingTags, notExistsNames, err := p.TagDAL.GetByNames(ctx, projectForm.TagIDs)
		if err != nil {
			return err
		}
		for _, tag := range existingTags {
			tagIDs = append(tagIDs, tag.ID)
		}
		for _, name := range notExistsNames {
			newTag := &schema.Tag{
				ID:        util.NewXID(),
				Name:      name,
				CreatedAt: time.Now(),
			}
			if err := p.TagDAL.Create(ctx, newTag); err != nil {
				return err
			}
			tagIDs = append(tagIDs, newTag.ID)
		}
	}

	if !p.canSetTop(ctx) {
		projectForm.IsTop = project.IsTop
	}
	if !p.canSetFeatured(ctx) {
		projectForm.IsFeatured = project.IsFeatured
		projectForm.FeaturedOrder = project.FeaturedOrder
	}

	wasDraft := project.Status != schema.ProjectStatusPublished

	if err := projectForm.FillTo(project); err != nil {
		return err
	}
	project.ID = id
	project.UpdatedAt = time.Now()

	err = p.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := p.ProjectDAL.Update(ctx, project); err != nil {
			return err
		}
		if err := p.ProjectTagDAL.DeleteByProjectID(ctx, id); err != nil {
			return err
		}
		var projectTags []schema.ProjectTag
		for _, tagID := range tagIDs {
			projectTags = append(projectTags, schema.ProjectTag{
				ProjectID: id,
				TagID:     tagID,
				CreatedAt: time.Now(),
			})
		}
		if len(projectTags) > 0 {
			if err := p.ProjectTagDAL.CreateBatch(ctx, projectTags); err != nil {
				return err
			}
		}
		if err := p.ProjectVisibleUserDAL.DeleteByProjectID(ctx, id); err != nil {
			return err
		}
		if len(validUserIDs) > 0 {
			var vuList []*schema.ProjectVisibleUser
			for _, uid := range validUserIDs {
				vuList = append(vuList, &schema.ProjectVisibleUser{
					ID:        util.NewXID(),
					ProjectID: id,
					UserID:    uid,
					CreatedAt: time.Now(),
				})
			}
			if err := p.ProjectVisibleUserDAL.BatchCreate(ctx, vuList); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	_ = p.ContributionDAL.RecordEdit(ctx, uid)
	if wasDraft && projectForm.Status == schema.ProjectStatusPublished {
		_ = p.ContributionDAL.RecordPublish(ctx, uid)
	}
	return nil
}

// Delete 删除项目（级联删除评论、标签关联、可见用户、时间线、资源、点赞、收藏）
func (p *Project) Delete(ctx context.Context, id string) error {
	project, err := p.ProjectDAL.Get(ctx, id, schema.ProjectQueryOptions{
		UserID: userID(ctx),
	})
	if err != nil {
		return err
	} else if project == nil {
		return errors.NotFound(config.ErrNotFound, "Project not found")
	}
	uid := userID(ctx)
	if project.AuthorID != "" && project.AuthorID != uid && !util.FromIsRootUser(ctx) {
		return errors.Forbidden("", "您无权删除他人发布的项目")
	}

	return p.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := p.ProjectDAL.Delete(ctx, id); err != nil {
			return err
		}
		if err := p.CommentDAL.DeleteByProjectID(ctx, id); err != nil {
			return err
		}
		if err := p.ProjectTagDAL.DeleteByProjectID(ctx, id); err != nil {
			return err
		}
		if err := p.ProjectVisibleUserDAL.DeleteByProjectID(ctx, id); err != nil {
			return err
		}
		if err := p.ProjectTimelineDAL.DeleteByProjectID(ctx, id); err != nil {
			return err
		}
		if err := p.ProjectResourceDAL.DeleteByProjectID(ctx, id); err != nil {
			return err
		}
		if err := p.ProjectLikeDAL.DeleteByProjectIDs(ctx, []string{id}); err != nil {
			return err
		}
		if err := p.ProjectFavoriteDAL.DeleteByProjectIDs(ctx, []string{id}); err != nil {
			return err
		}
		return nil
	})
}

// ToggleTop 切换项目置顶状态
func (p *Project) ToggleTop(ctx context.Context, id string, isTop bool) error {
	if !p.canSetTop(ctx) {
		return errors.Forbidden("", "您无权修改项目置顶状态")
	}
	return p.ProjectDAL.Update(ctx, &schema.Project{
		ID:    id,
		IsTop: isTop,
	}, "is_top")
}

// ToggleFeatured 切换项目精选状态
func (p *Project) ToggleFeatured(ctx context.Context, id string, isFeatured bool) error {
	if !p.canSetFeatured(ctx) {
		return errors.Forbidden("", "您无权修改项目精选状态")
	}
	return p.ProjectDAL.Update(ctx, &schema.Project{
		ID:         id,
		IsFeatured: isFeatured,
	}, "is_featured")
}

// UpdateStatus 批量更新项目状态
func (p *Project) UpdateStatus(ctx context.Context, ids []string, status string) error {
	if len(ids) == 0 {
		return nil
	}
	return p.ProjectDAL.UpdateStatus(ctx, ids, status)
}

// IncViews 增加浏览量
func (p *Project) IncViews(ctx context.Context, id string, num int64) error {
	return p.ProjectDAL.IncViews(ctx, id, num)
}

func (p *Project) validateVisibleUsers(ctx context.Context, form *schema.ProjectForm) ([]string, error) {
	if form.Visibility != schema.ProjectVisibilityPartialVisible {
		return nil, nil
	}
	seen := make(map[string]bool, len(form.VisibleUserIDs))
	var unique []string
	for _, id := range form.VisibleUserIDs {
		if id == "" {
			continue
		}
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	if len(unique) == 0 {
		return nil, errors.BadRequest(config.ErrProjectVisibleUsersEmpty, "partial_visible 可见范围必须指定至少一个用户")
	}
	missing, err := p.ProjectVisibleUserDAL.ValidateUsersExist(ctx, unique)
	if err != nil {
		return nil, err
	}
	if len(missing) > 0 {
		return nil, errors.BadRequest(config.ErrBadRequest, "用户不存在: %v", missing)
	}
	return unique, nil
}
