// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/mods/blog/biz"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/util"
)

type Article struct {
	ArticleBIZ *biz.Article
}

// @Tags ArticleAPI
// @Security BearerAuth
// @Summary Query article list
// @Param current query int false "pagination index" default(1)
// @Param pageSize query int false "pagination size" default(10)
// @Param title query string false "Article title (fuzzy)"
// @Param category_id query string false "Category ID"
// @Param status query string false "Status (draft/published)" Enums(draft,published)
// @Param is_top query boolean false "Is top"
// @Param published_at_gte query string false "Published at >= (YYYY-MM-DD)"
// @Param published_at_lte query string false "Published at <= (YYYY-MM-DD)"
// @Success 200 {object} util.ResponseResult{data=[]schema.Article}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/articles [get]
func (a *Article) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.ArticleQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := a.ArticleBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags ArticleAPI
// @Security BearerAuth
// @Summary Get article by ID
// @Param id path string true "Article ID"
// @Success 200 {object} util.ResponseResult{data=schema.Article}
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/articles/{id} [get]
func (a *Article) Get(c *gin.Context) {
	ctx := c.Request.Context()
	article, err := a.ArticleBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, article)
}

// @Tags ArticleAPI
// @Summary Get published article by slug (public)
// @Description Get article by slug, only published articles are accessible
// @Param slug path string true "Article slug"
// @Success 200 {object} util.ResponseResult{data=schema.Article}
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/articles/slug/{slug} [get]
func (a *Article) GetBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	article, err := a.ArticleBIZ.GetBySlug(ctx, c.Param("slug"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, article)
}

// @Tags ArticleAPI
// @Security BearerAuth
// @Summary Create article
// @Param body body schema.ArticleForm true "Article form"
// @Success 200 {object} util.ResponseResult{data=schema.Article}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/articles [post]
func (a *Article) Create(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.ArticleForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := a.ArticleBIZ.Create(ctx, form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags ArticleAPI
// @Security BearerAuth
// @Summary Update article by ID
// @Param id path string true "Article ID"
// @Param body body schema.ArticleForm true "Article form"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/articles/{id} [put]
func (a *Article) Update(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.ArticleForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	err := a.ArticleBIZ.Update(ctx, c.Param("id"), form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags ArticleAPI
// @Security BearerAuth
// @Summary Delete article by ID
// @Param id path string true "Article ID"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/articles/{id} [delete]
func (a *Article) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := a.ArticleBIZ.Delete(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags ArticleAPI
// @Security BearerAuth
// @Summary Batch update article status
// @Param body body schema.BatchUpdateStatusForm true "IDs and status"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/articles/status [patch]
func (a *Article) UpdateStatus(c *gin.Context) {
	ctx := c.Request.Context()
	var req struct {
		IDs    []string `json:"ids" binding:"required,min=1"`
		Status string   `json:"status" binding:"required,oneof=draft published"`
	}
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	err := a.ArticleBIZ.UpdateStatus(ctx, req.IDs, req.Status)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags ArticleAPI
// @Summary Get article archive (public)
// @Description Get article archive grouped by year and month
// @Success 200 {object} util.ResponseResult{data=[]schema.ArchiveItem}
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/archives [get]
func (a *Article) GetArchive(c *gin.Context) {
	ctx := c.Request.Context()
	items, err := a.ArticleBIZ.GetArchive(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, items)
}

// @Tags ArticleAPI
// @Security BearerAuth
// @Summary Toggle article top status
// @Param id path string true "Article ID"
// @Param body body object{is_top=bool} true "is_top flag"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/articles/{id}/top [patch]
func (a *Article) ToggleTop(c *gin.Context) {
	ctx := c.Request.Context()
	var req struct {
		IsTop bool `json:"is_top"`
	}
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	if err := a.ArticleBIZ.ToggleTop(ctx, c.Param("id"), req.IsTop); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags ArticleAPI
// @Summary Increment article views (public)
// @Description Atomic increment article views by ID
// @Param id path string true "Article ID"
// @Param num query int false "Increment number" default(1)
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/articles/{id}/views [post]
func (a *Article) IncViews(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	num := util.ToInt64(c.DefaultQuery("num", "1"))
	if num <= 0 {
		num = 1
	}

	err := a.ArticleBIZ.IncViews(ctx, id, num)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}
