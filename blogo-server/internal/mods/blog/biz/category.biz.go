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

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

type Category struct {
	Trans       *util.Trans
	CategoryDAL *dal.Category
	ArticleDAL  *dal.Article
}

func (c *Category) Query(ctx context.Context, params schema.CategoryQueryParam) (*schema.CategoryQueryResult, error) {
	params.Pagination = true
	result, err := c.CategoryDAL.Query(ctx, params, schema.CategoryQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "sort", Direction: util.ASC},
				{Field: "created_at", Direction: util.DESC},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Category) Get(ctx context.Context, id string) (*schema.Category, error) {
	category, err := c.CategoryDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if category == nil {
		return nil, errors.NotFound("", "Category not found")
	}
	return category, nil
}

func (c *Category) Create(ctx context.Context, categoryForm *schema.CategoryForm) (*schema.Category, error) {
	exists, err := c.CategoryDAL.ExistsName(ctx, categoryForm.Name)
	if err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Category name already exists")
	}
	if err := categoryForm.Validate(); err != nil {
		return nil, err
	}
	category := &schema.Category{ID: util.NewXID(), CreatedAt: time.Now()}
	categoryForm.FillTo(category)
	err = c.Trans.Exec(ctx, func(ctx context.Context) error {
		return c.CategoryDAL.Create(ctx, category)
	})
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, category.ID)
}

func (c *Category) Update(ctx context.Context, id string, categoryForm *schema.CategoryForm) error {
	oldCategory, err := c.CategoryDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if oldCategory == nil {
		return errors.NotFound("", "Category not found")
	}
	if oldCategory.Name != categoryForm.Name {
		exists, err := c.CategoryDAL.ExistsName(ctx, categoryForm.Name)
		if err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "Category name already exists")
		}
	}
	if err := categoryForm.Validate(); err != nil {
		return err
	}
	categoryForm.FillTo(oldCategory)
	oldCategory.UpdatedAt = time.Now()
	return c.Trans.Exec(ctx, func(ctx context.Context) error {
		return c.CategoryDAL.Update(ctx, oldCategory)
	})
}

// Delete 删除分类。有文章引用时拒绝，提示具体数量。
func (c *Category) Delete(ctx context.Context, id string) error {
	exists, err := c.CategoryDAL.ExistsID(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Category not found")
	}
	count, err := c.ArticleDAL.CountByCategoryID(ctx, id)
	if err != nil {
		return err
	} else if count > 0 {
		return errors.BadRequest("", "该分类下有 %d 篇文章，请先将它们移入其他分类后再删除", count)
	}
	return c.Trans.Exec(ctx, func(ctx context.Context) error {
		return c.CategoryDAL.Delete(ctx, id)
	})
}

func (c *Category) DeleteByIds(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		count, err := c.ArticleDAL.CountByCategoryID(ctx, id)
		if err != nil {
			return err
		} else if count > 0 {
			return errors.BadRequest("", "分类 %s 下有 %d 篇文章，请先移走后再删除", id, count)
		}
	}
	return c.Trans.Exec(ctx, func(ctx context.Context) error {
		return c.CategoryDAL.DeleteByIds(ctx, ids)
	})
}

func (c *Category) GetAll(ctx context.Context) (schema.Categories, error) {
	return c.CategoryDAL.GetAll(ctx)
}
