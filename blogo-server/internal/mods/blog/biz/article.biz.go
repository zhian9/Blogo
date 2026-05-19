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
	"github.com/zhian9/blogo-server/internal/mods/blog/msg"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	rschema "github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// Article 是文章管理业务的核心对象
type Article struct {
	Trans                 *util.Trans
	ArticleDAL            *dal.Article
	CategoryDAL           *dal.Category
	TagDAL                *dal.Tag
	ArticleTagDAL         *dal.ArticleTag
	CommentDAL            *dal.Comment
	ArticleVisibleUserDAL *dal.ArticleVisibleUser
	ContributionDAL       *dal.Contribution
	ArticlePermission     *ArticlePermission
	MailWorker            *msg.MailWorker
}

// userID 安全提取当前用户 ID，未登录返回空字符串
func userID(ctx context.Context) string {
	return util.FromUserID(ctx)
}

// hasRoleCode 检查当前用户是否拥有指定角色码之一（root 用户视为 admin）
func (a *Article) hasRoleCode(ctx context.Context, codes ...string) bool {
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
	// fallback: 查询数据库获取角色码
	var count int64
	a.Trans.DB.Model(&rschema.Role{}).
		Where("id IN ? AND code IN ?", userCache.RoleIDs, codes).
		Count(&count)
	return count > 0
}

// canSetTop 检查当前用户是否有置顶权限
func (a *Article) canSetTop(ctx context.Context) bool {
	return a.hasRoleCode(ctx, "super_admin", "admin", "content_manager")
}

// Query 查询文章列表。
// 已登录用户可看自己的 private/draft；未登录仅看 published+public
func (a *Article) Query(ctx context.Context, params schema.ArticleQueryParam) (*schema.ArticleQueryResult, error) {
	params.Pagination = true

	result, err := a.ArticleDAL.Query(ctx, params, schema.ArticleQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "is_top", Direction: util.DESC},
				{Field: "published_at", Direction: util.DESC},
			},
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

// Get 获取单篇文章（含权限校验）。
// 无权限 → 403；不存在 → 404
func (a *Article) Get(ctx context.Context, id string) (*schema.Article, error) {
	article, err := a.ArticleDAL.Get(ctx, id, schema.ArticleQueryOptions{
		WithCategory:     true,
		WithTags:         true,
		WithCoverImage:   true,
		WithAuthor:       true,
		WithVisibleUsers: true,
		UserID:           userID(ctx),
	})
	if err != nil {
		return nil, err
	} else if article == nil {
		return nil, errors.NotFound(config.ErrNotFound, "Article not found")
	}
	if !a.ArticlePermission.CanUserReadArticle(ctx, userID(ctx), article) {
		return nil, errors.Forbidden(config.ErrArticlePermissionDenied, "无权限访问该文章")
	}
	return article, nil
}

// GetBySlug 根据 Slug 获取文章（公开访问 + 权限控制）
func (a *Article) GetBySlug(ctx context.Context, slug string) (*schema.Article, error) {
	article, err := a.ArticleDAL.GetBySlug(ctx, slug, schema.ArticleQueryOptions{
		WithCategory:     true,
		WithTags:         true,
		WithCoverImage:   true,
		WithAuthor:       true,
		WithVisibleUsers: true,
		UserID:           userID(ctx),
	})
	if err != nil {
		return nil, err
	} else if article == nil {
		return nil, errors.NotFound(config.ErrNotFound, "Article not found or not published")
	}
	if !a.ArticlePermission.CanUserReadArticle(ctx, userID(ctx), article) {
		return nil, errors.Forbidden(config.ErrArticlePermissionDenied, "无权限访问该文章")
	}
	return article, nil
}

// Create 创建新文章。
// 自动记录 AuthorID；处理可见用户。guest 角色禁止发文。
func (a *Article) Create(ctx context.Context, articleForm *schema.ArticleForm) (*schema.Article, error) {
	if a.hasRoleCode(ctx, "guest") && !util.FromIsRootUser(ctx) {
		return nil, errors.Forbidden("", "游客无权发布文章，请先登录或联系管理员")
	}

	// 1. Slug 唯一性校验
	existsSlug, err := a.ArticleDAL.ExistsSlug(ctx, articleForm.Slug)
	if err != nil {
		return nil, err
	} else if existsSlug {
		return nil, errors.BadRequest(config.ErrBadRequest, "Slug already exists")
	}

	// 2. 校验可见用户
	validUserIDs, err := a.validateVisibleUsers(ctx, articleForm)
	if err != nil {
		return nil, err
	}

	// 3. 分类 FindOrCreate：先按 ID 查，失败则按名称查，都不存在则自动创建
	categoryID := articleForm.CategoryID
	if categoryID != "" {
		exists, err := a.CategoryDAL.ExistsID(ctx, categoryID)
		if err != nil {
			return nil, err
		}
		if !exists {
			categoryID, err = a.CategoryDAL.FindOrCreateByName(ctx, categoryID)
			if err != nil {
				return nil, err
			}
		}
	}
	articleForm.CategoryID = categoryID

	// 4. 处理标签：FindOrCreate（存在则复用ID，不存在则按名称创建）
	var tagIDs []string
	if len(articleForm.TagIDs) > 0 {
		existingTags, notExistsNames, err := a.TagDAL.GetByNames(ctx, articleForm.TagIDs)
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
			if err := a.TagDAL.Create(ctx, newTag); err != nil {
				return nil, err
			}
			tagIDs = append(tagIDs, newTag.ID)
		}
	}

	// 5. 初始化文章实体 + 记录作者
	article := &schema.Article{
		ID:          util.NewXID(),
		AuthorID:    userID(ctx),
		CreatedAt:   time.Now(),
		PublishedAt: time.Now(),
	}

	// 6. 普通用户禁止置顶
	if !a.canSetTop(ctx) {
		articleForm.IsTop = false
	}

	// 7. FillTo：自动存稿 + MD→HTML
	if err := articleForm.FillTo(article); err != nil {
		return nil, err
	}

	// 7. 事务：创建文章 + 标签关联 + 可见用户关联
	err = a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.ArticleDAL.Create(ctx, article); err != nil {
			return err
		}
		var articleTags []schema.ArticleTag
		for _, tagID := range tagIDs {
			articleTags = append(articleTags, schema.ArticleTag{
				ArticleID: article.ID,
				TagID:     tagID,
				CreatedAt: time.Now(),
			})
		}
		if len(articleTags) > 0 {
			if err := a.ArticleTagDAL.CreateBatch(ctx, articleTags); err != nil {
				return err
			}
		}
		// 创建可见用户关联
		if len(validUserIDs) > 0 {
			var vuList []*schema.ArticleVisibleUser
			for _, uid := range validUserIDs {
				vuList = append(vuList, &schema.ArticleVisibleUser{
					ID:        util.NewXID(),
					ArticleID: article.ID,
					UserID:    uid,
					CreatedAt: time.Now(),
				})
			}
			if err := a.ArticleVisibleUserDAL.BatchCreate(ctx, vuList); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 异步通知订阅者（仅发布文章时触发）
	if articleForm.Status == schema.ArticleStatusPublished {
		a.MailWorker.Send(msg.MailMessage{
			Type:         msg.MailTypeNewArticle,
			ArticleTitle: article.Title,
			ArticleSlug:  article.Slug,
			SiteURL:      config.C.General.SiteURL,
		})
		// 记录发布贡献
		_ = a.ContributionDAL.RecordPublish(ctx, article.AuthorID)
	}

	return a.Get(ctx, article.ID)
}

// Update 更新文章信息。
// 事务：更新文章 + 标签 + 可见用户
func (a *Article) Update(ctx context.Context, id string, articleForm *schema.ArticleForm) error {
	article, err := a.ArticleDAL.Get(ctx, id, schema.ArticleQueryOptions{
		UserID: userID(ctx),
	})
	if err != nil {
		return err
	} else if article == nil {
		return errors.NotFound(config.ErrNotFound, "Article not found")
	}

	uid := userID(ctx)
	if article.AuthorID != "" && article.AuthorID != uid && !util.FromIsRootUser(ctx) {
		return errors.Forbidden("", "您无权编辑他人发布的文章")
	}

	if article.Slug != articleForm.Slug {
		existsSlug, err := a.ArticleDAL.ExistsSlug(ctx, articleForm.Slug)
		if err != nil {
			return err
		} else if existsSlug {
			return errors.BadRequest(config.ErrBadRequest, "Slug already exists")
		}
	}

	// 校验可见用户
	validUserIDs, err := a.validateVisibleUsers(ctx, articleForm)
	if err != nil {
		return err
	}

	if articleForm.CategoryID != "" {
		exists, err := a.CategoryDAL.ExistsID(ctx, articleForm.CategoryID)
		if err != nil {
			return err
		}
		if !exists {
			categoryID, err := a.CategoryDAL.FindOrCreateByName(ctx, articleForm.CategoryID)
			if err != nil {
				return err
			}
			articleForm.CategoryID = categoryID
		}
	}

	var tagIDs []string
	if len(articleForm.TagIDs) > 0 {
		existingTags, notExistsNames, err := a.TagDAL.GetByNames(ctx, articleForm.TagIDs)
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
			if err := a.TagDAL.Create(ctx, newTag); err != nil {
				return err
			}
			tagIDs = append(tagIDs, newTag.ID)
		}
	}

	// 普通用户禁止修改置顶状态
	if !a.canSetTop(ctx) {
		articleForm.IsTop = article.IsTop // 保持原有置顶状态
	}

	if err := articleForm.FillTo(article); err != nil {
		return err
	}
	// 保留原作者
	article.ID = id
	article.UpdatedAt = time.Now()

	err = a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.ArticleDAL.Update(ctx, article); err != nil {
			return err
		}
		if err := a.ArticleTagDAL.DeleteByArticleID(ctx, id); err != nil {
			return err
		}
		var articleTags []schema.ArticleTag
		for _, tagID := range tagIDs {
			articleTags = append(articleTags, schema.ArticleTag{
				ArticleID: id,
				TagID:     tagID,
				CreatedAt: time.Now(),
			})
		}
		if len(articleTags) > 0 {
			if err := a.ArticleTagDAL.CreateBatch(ctx, articleTags); err != nil {
				return err
			}
		}
		// 重建可见用户关联（先删后插）
		if err := a.ArticleVisibleUserDAL.DeleteByArticleID(ctx, id); err != nil {
			return err
		}
		if len(validUserIDs) > 0 {
			var vuList []*schema.ArticleVisibleUser
			for _, uid := range validUserIDs {
				vuList = append(vuList, &schema.ArticleVisibleUser{
					ID:        util.NewXID(),
					ArticleID: id,
					UserID:    uid,
					CreatedAt: time.Now(),
				})
			}
			if err := a.ArticleVisibleUserDAL.BatchCreate(ctx, vuList); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	_ = a.ContributionDAL.RecordEdit(ctx, uid)
	return nil
}

// Delete 删除文章（级联删除评论、标签关联、可见用户关联）
func (a *Article) Delete(ctx context.Context, id string) error {
	article, err := a.ArticleDAL.Get(ctx, id, schema.ArticleQueryOptions{
		UserID: userID(ctx),
	})
	if err != nil {
		return err
	} else if article == nil {
		return errors.NotFound(config.ErrNotFound, "Article not found")
	}
	uid := userID(ctx)
	if article.AuthorID != "" && article.AuthorID != uid && !util.FromIsRootUser(ctx) {
		return errors.Forbidden("", "您无权删除他人发布的文章")
	}

	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.ArticleDAL.Delete(ctx, id); err != nil {
			return err
		}
		if err := a.CommentDAL.DeleteByArticleID(ctx, id); err != nil {
			return err
		}
		if err := a.ArticleTagDAL.DeleteByArticleID(ctx, id); err != nil {
			return err
		}
		if err := a.ArticleVisibleUserDAL.DeleteByArticleID(ctx, id); err != nil {
			return err
		}
		return nil
	})
}

// ToggleTop 切换文章置顶状态（仅 admin / content_manager 可操作）
func (a *Article) ToggleTop(ctx context.Context, id string, isTop bool) error {
	if !a.canSetTop(ctx) {
		return errors.Forbidden("", "您无权修改文章置顶状态")
	}

	if err := a.ArticleDAL.Update(ctx, &schema.Article{
		ID:    id,
		IsTop: isTop,
	}, "is_top"); err != nil {
		return err
	}
	return nil
}

// validateVisibleUsers 校验表单中的可见用户配置。
// public / private → VisibleUserIDs 被忽略（返回空列表）
// partial_visible → 必须提供 visible_user_ids，且所有用户必须存在
func (a *Article) validateVisibleUsers(ctx context.Context, form *schema.ArticleForm) ([]string, error) {
	if form.Visibility != schema.ArticleVisibilityPartialVisible {
		return nil, nil
	}
	// 去重
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
		return nil, errors.BadRequest(config.ErrArticleVisibleUsersEmpty, "partial_visible 可见范围必须指定至少一个用户")
	}
	// 校验用户是否存在
	missing, err := a.ArticleVisibleUserDAL.ValidateUsersExist(ctx, unique)
	if err != nil {
		return nil, err
	}
	if len(missing) > 0 {
		return nil, errors.BadRequest(config.ErrBadRequest, "用户不存在: %v", missing)
	}
	return unique, nil
}

func (a *Article) UpdateStatus(ctx context.Context, ids []string, status string) error {
	if len(ids) == 0 {
		return nil
	}
	return a.ArticleDAL.UpdateStatus(ctx, ids, status)
}

func (a *Article) IncViews(ctx context.Context, id string, num int64) error {
	return a.ArticleDAL.IncViews(ctx, id, num)
}

func (a *Article) GetArchive(ctx context.Context) ([]schema.ArchiveItem, error) {
	return a.ArticleDAL.GetArchive(ctx)
}
