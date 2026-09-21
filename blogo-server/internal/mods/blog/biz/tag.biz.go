// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by a [MIT/Apache/BSD] style license
// that can be found in the LICENSE file.

package biz

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
)

type Tag struct {
	Trans         *util.Trans     // 事务管理器
	TagDAL        *dal.Tag        // 标签数据访问层
	ArticleDAL    *dal.Article    // 文章数据访问层（用于删除前检查）
	ArticleTagDAL *dal.ArticleTag // 中间表（用于清理关联）
}

// Query 查询标签列表（附带每个标签的引用文章数，便于后台判断能否删除）
func (t *Tag) Query(ctx context.Context, params schema.TagQueryParam) (*schema.TagQueryResult, error) {
	params.Pagination = true
	result, err := t.TagDAL.Query(ctx, params, schema.TagQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "created_at", Direction: util.DESC},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	t.fillArticleCount(ctx, result.Data)
	return result, nil
}

// fillArticleCount 批量填充标签的引用文章数
func (t *Tag) fillArticleCount(ctx context.Context, tags schema.Tags) {
	if len(tags) == 0 {
		return
	}

	counts, err := t.ArticleTagDAL.CountByTagIDs(ctx, tags.ToIDs())
	if err != nil {
		logging.Context(ctx).Error("fill tag article count failed", zap.Error(err))
		return
	}
	for _, tag := range tags {
		tag.ArticleCount = counts[tag.ID]
	}
}

// GetReferences 查询标签被哪些文章引用（后台删除标签前的提示用）
func (t *Tag) GetReferences(ctx context.Context, id string) (*schema.TagReferenceResult, error) {
	tag, err := t.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	articleIDs, err := t.ArticleTagDAL.GetArticlesByTagID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &schema.TagReferenceResult{TagID: tag.ID, TagName: tag.Name}
	if len(articleIDs) == 0 {
		return result, nil
	}

	articles, err := t.ArticleDAL.GetBriefByIDs(ctx, articleIDs)
	if err != nil {
		return nil, err
	}
	for _, article := range articles {
		result.Articles = append(result.Articles, &schema.TagReference{
			ID:          article.ID,
			Title:       article.Title,
			Slug:        article.Slug,
			Status:      article.Status,
			PublishedAt: &article.PublishedAt,
		})
	}
	result.Total = int64(len(articleIDs))
	return result, nil
}

// Get 获取单个标签
func (t *Tag) Get(ctx context.Context, id string) (*schema.Tag, error) {
	tag, err := t.TagDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if tag == nil {
		return nil, errors.NotFound("", "Tag not found")
	}
	return tag, nil
}

// Create 创建新标签
func (t *Tag) Create(ctx context.Context, form *schema.TagForm) (*schema.Tag, error) {
	// 1. 名称唯一性校验
	exists, err := t.TagDAL.ExistsName(ctx, form.Name)
	if err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Tag name already exists")
	}

	if err := form.Validate(); err != nil {
		return nil, err
	}

	// 3. 创建标签（在事务外生成 ID）
	newTag := &schema.Tag{
		ID:        util.NewXID(),
		CreatedAt: time.Now(),
	}
	form.FillTo(newTag)

	err = t.Trans.Exec(ctx, func(ctx context.Context) error {
		return t.TagDAL.Create(ctx, newTag)
	})
	if err != nil {
		return nil, err
	}
	return newTag, nil
}

// Update 更新标签
func (t *Tag) Update(ctx context.Context, id string, form *schema.TagForm) error {
	tag, err := t.TagDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if tag == nil {
		return errors.NotFound("", "Tag not found")
	}

	if tag.Name != form.Name {
		exists, err := t.TagDAL.ExistsName(ctx, form.Name)
		if err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "Tag name already exists")
		}
	}

	if err := form.Validate(); err != nil {
		return err
	}

	form.FillTo(tag)
	tag.UpdatedAt = time.Now()

	return t.Trans.Exec(ctx, func(ctx context.Context) error {
		return t.TagDAL.Update(ctx, tag)
	})
}

// Delete 删除标签（需确保无文章引用）
func (t *Tag) Delete(ctx context.Context, id string) error {
	exists, err := t.TagDAL.ExistsID(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Tag not found")
	}

	articleIDs, _ := t.ArticleTagDAL.GetArticlesByTagID(ctx, id)
	if n := len(articleIDs); n > 0 {
		return errors.BadRequest("", "该标签被 %d 篇文章引用，请先解除关联后再删除", n)
	}
	return t.Trans.Exec(ctx, func(ctx context.Context) error {
		return t.TagDAL.Delete(ctx, id)
	})
}

// GetAll 获取所有标签（用于标签云）
func (t *Tag) GetAll(ctx context.Context) (schema.Tags, error) {
	return t.TagDAL.GetAll(ctx)
}
