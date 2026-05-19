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

func GetArticleDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Article))
}

type Article struct {
	DB *gorm.DB
}

func (a *Article) loadCoverImages(ctx context.Context, articles schema.Articles) error {
	var ids []string
	for _, art := range articles {
		if art.CoverImageID != nil && *art.CoverImageID != "" {
			ids = append(ids, *art.CoverImageID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var images []schema.Image
	if err := util.GetDB(ctx, a.DB).Model(new(schema.Image)).Where("id IN ?", ids).Find(&images).Error; err != nil {
		return errors.WithStack(err)
	}
	imgMap := make(map[string]*schema.Image, len(images))
	for i := range images {
		imgMap[images[i].ID] = &images[i]
	}
	for _, art := range articles {
		if art.CoverImageID != nil {
			art.CoverImage = imgMap[*art.CoverImageID]
		}
	}
	return nil
}

func (a *Article) loadAuthors(ctx context.Context, articles schema.Articles) error {
	var ids []string
	for _, art := range articles {
		if art.AuthorID != "" {
			ids = append(ids, art.AuthorID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var users []rschema.User
	if err := util.GetDB(ctx, a.DB).Model(new(rschema.User)).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return errors.WithStack(err)
	}
	userMap := make(map[string]*rschema.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}
	for _, art := range articles {
		if art.AuthorID != "" {
			art.Author = userMap[art.AuthorID]
		}
	}
	return nil
}

func (a *Article) loadVisibleUsers(ctx context.Context, articles schema.Articles) error {
	var ids []string
	for _, art := range articles {
		if art.Visibility == schema.ArticleVisibilityPartialVisible {
			ids = append(ids, art.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var vuList []*schema.ArticleVisibleUser
	if err := util.GetDB(ctx, a.DB).Model(new(schema.ArticleVisibleUser)).Where("article_id IN ?", ids).Find(&vuList).Error; err != nil {
		return errors.WithStack(err)
	}
	vuMap := make(map[string][]*schema.ArticleVisibleUser, len(ids))
	for _, vu := range vuList {
		vuMap[vu.ArticleID] = append(vuMap[vu.ArticleID], vu)
	}
	for _, art := range articles {
		if art.Visibility == schema.ArticleVisibilityPartialVisible {
			art.VisibleUsers = vuMap[art.ID]
		}
	}
	return nil
}

func (a *Article) Query(ctx context.Context, params schema.ArticleQueryParam, opts ...schema.ArticleQueryOptions) (*schema.ArticleQueryResult, error) {
	var opt schema.ArticleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetArticleDB(ctx, a.DB)

	if opt.UserID != "" {
		// 已登录用户可见范围：
		//   - 已发布 + 公开
		//   - 自己的文章（任意可见性）
		//   - 已发布 + 部分可见 且 user_id 在可见用户列表中
		visibleSubQuery := a.DB.Model(&schema.ArticleVisibleUser{}).
			Select("article_id").
			Where("user_id = ?", opt.UserID)
		db = db.Where(
			"(status = ? AND visibility = ?) OR author_id = ? OR (status = ? AND visibility = ? AND id IN (?))",
			schema.ArticleStatusPublished, schema.ArticleVisibilityPublic,
			opt.UserID,
			schema.ArticleStatusPublished, schema.ArticleVisibilityPartialVisible,
			visibleSubQuery,
		)
	} else {
		db = db.Where("status = ? AND visibility = ?",
			schema.ArticleStatusPublished, schema.ArticleVisibilityPublic)
	}

	if v := params.Title; len(v) > 0 {
		db = db.Where("title LIKE ?", "%"+v+"%")
	}
	if v := params.CategoryID; len(v) > 0 {
		db = db.Where("category_id = ?", v)
	}
	if v := params.TagName; len(v) > 0 {
		// GORM model subquery — table names auto-prefixed via TableName()
		subQuery := a.DB.Model(&schema.ArticleTag{}).
			Select("article_id").
			Where("tag_id IN (?)",
				a.DB.Model(&schema.Tag{}).Select("id").Where("name = ?", v),
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
	if v := params.IsTop; v != nil {
		db = db.Where("is_top = ?", *v)
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

	var list schema.Articles
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if opt.WithCoverImage && len(list) > 0 {
		if err := a.loadCoverImages(ctx, list); err != nil {
			return nil, err
		}
	}
	if opt.WithAuthor && len(list) > 0 {
		if err := a.loadAuthors(ctx, list); err != nil {
			return nil, err
		}
	}
	if opt.WithVisibleUsers && len(list) > 0 {
		if err := a.loadVisibleUsers(ctx, list); err != nil {
			return nil, err
		}
	}

	return &schema.ArticleQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

func (a *Article) Get(ctx context.Context, id string, opts ...schema.ArticleQueryOptions) (*schema.Article, error) {
	var opt schema.ArticleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	article := new(schema.Article)
	db := GetArticleDB(ctx, a.DB).Where("id = ?", id)
	if opt.WithCategory {
		db = db.Preload("Category")
	}
	if opt.WithTags {
		db = db.Preload("Tags")
	}

	ok, err := util.FindOne(ctx, db, opt.QueryOptions, article)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}

	if opt.WithCoverImage {
		if err := a.loadCoverImages(ctx, schema.Articles{article}); err != nil {
			return nil, err
		}
	}
	if opt.WithAuthor {
		if err := a.loadAuthors(ctx, schema.Articles{article}); err != nil {
			return nil, err
		}
	}
	if opt.WithVisibleUsers {
		if err := a.loadVisibleUsers(ctx, schema.Articles{article}); err != nil {
			return nil, err
		}
	}
	return article, nil
}

func (a *Article) GetBySlug(ctx context.Context, slug string, opts ...schema.ArticleQueryOptions) (*schema.Article, error) {
	var opt schema.ArticleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	article := new(schema.Article)
	db := GetArticleDB(ctx, a.DB).Where("slug = ?", slug)
	if opt.UserID == "" {
		db = db.Where("status = ? AND visibility = ?", schema.ArticleStatusPublished, schema.ArticleVisibilityPublic)
	} else {
		visibleSubQuery := a.DB.Model(&schema.ArticleVisibleUser{}).
			Select("article_id").
			Where("user_id = ?", opt.UserID)
		db = db.Where(
			"(status = ? AND visibility = ?) OR author_id = ? OR (status = ? AND visibility = ? AND id IN (?))",
			schema.ArticleStatusPublished, schema.ArticleVisibilityPublic,
			opt.UserID,
			schema.ArticleStatusPublished, schema.ArticleVisibilityPartialVisible,
			visibleSubQuery,
		)
	}
	if opt.WithCategory {
		db = db.Preload("Category")
	}
	if opt.WithTags {
		db = db.Preload("Tags")
	}

	ok, err := util.FindOne(ctx, db, opt.QueryOptions, article)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}

	if opt.WithCoverImage {
		if err := a.loadCoverImages(ctx, schema.Articles{article}); err != nil {
			return nil, err
		}
	}
	if opt.WithAuthor {
		if err := a.loadAuthors(ctx, schema.Articles{article}); err != nil {
			return nil, err
		}
	}
	if opt.WithVisibleUsers {
		if err := a.loadVisibleUsers(ctx, schema.Articles{article}); err != nil {
			return nil, err
		}
	}
	return article, nil
}

func (a *Article) canView(article *schema.Article, userID string) bool {
	if article == nil {
		return false
	}
	// 已发布 + 公开 → 任何人可读
	if article.Status == schema.ArticleStatusPublished && article.Visibility == schema.ArticleVisibilityPublic {
		return true
	}
	// 未登录 → 不可读非公开文章
	if userID == "" {
		return false
	}
	// 作者本人 → 可读
	if article.AuthorID == userID {
		return true
	}
	// 部分可见 → 查表
	if article.Status == schema.ArticleStatusPublished && article.Visibility == schema.ArticleVisibilityPartialVisible {
		var count int64
		a.DB.Model(&schema.ArticleVisibleUser{}).
			Where("article_id = ? AND user_id = ?", article.ID, userID).
			Count(&count)
		return count > 0
	}
	return false
}

func (a *Article) ExistsID(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetArticleDB(ctx, a.DB).Where("id = ?", id))
	return ok, errors.WithStack(err)
}
func (a *Article) ExistsSlug(ctx context.Context, slug string) (bool, error) {
	ok, err := util.Exists(ctx, GetArticleDB(ctx, a.DB).Where("slug = ?", slug))
	return ok, errors.WithStack(err)
}
func (a *Article) Create(ctx context.Context, article *schema.Article) error {
	result := GetArticleDB(ctx, a.DB).Create(article)
	return errors.WithStack(result.Error)
}
func (a *Article) Update(ctx context.Context, article *schema.Article, selectFields ...string) error {
	db := GetArticleDB(ctx, a.DB).Where("id = ?", article.ID)
	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("id", "created_at")
	}
	result := db.Updates(article)
	return errors.WithStack(result.Error)
}
func (a *Article) Delete(ctx context.Context, id string) error {
	result := GetArticleDB(ctx, a.DB).Where("id = ?", id).Delete(new(schema.Article))
	return errors.WithStack(result.Error)
}
func (a *Article) DeleteByIds(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	result := GetArticleDB(ctx, a.DB).Where("id IN ?", ids).Delete(new(schema.Article))
	return errors.WithStack(result.Error)
}
func (a *Article) UpdateStatus(ctx context.Context, ids []string, status string) error {
	if len(ids) == 0 {
		return nil
	}
	result := GetArticleDB(ctx, a.DB).Where("id IN ?", ids).Update("status", status)
	return errors.WithStack(result.Error)
}
func (a *Article) IncViews(ctx context.Context, id string, num int64) error {
	result := GetArticleDB(ctx, a.DB).Where("id = ?", id).Update("views", gorm.Expr("views + ?", num))
	return errors.WithStack(result.Error)
}
func (a *Article) GetArchive(ctx context.Context) ([]schema.ArchiveItem, error) {
	var items []schema.ArchiveItem
	err := GetArticleDB(ctx, a.DB).
		Select("YEAR(published_at) as year, MONTH(published_at) as month, COUNT(*) as count").
		Where("status = ? AND visibility = ?", schema.ArticleStatusPublished, schema.ArticleVisibilityPublic).
		Group("year, month").Order("year DESC, month DESC").Find(&items).Error
	return items, errors.WithStack(err)
}
func (a *Article) CountByCategoryID(ctx context.Context, categoryID string) (int64, error) {
	if categoryID == "" {
		return 0, nil
	}
	var count int64
	err := GetArticleDB(ctx, a.DB).Where("category_id = ?", categoryID).Count(&count).Error
	return count, errors.WithStack(err)
}
