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

type Image struct {
	ImageBIZ *biz.Image
}

// @Tags ImageAPI
// @Security BearerAuth
// @Summary Upload image file
// @Description Upload an image file (multipart/form-data). Creates DB record and saves to disk.
// @Param file formData file true "Image file"
// @Param category formData string false "Category" Enums(avatar,article_cover,ad_image,friend_link_logo)
// @Success 200 {object} util.ResponseResult{data=schema.Image}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/images/upload [post]
func (i *Image) Upload(c *gin.Context) {
	ctx := c.Request.Context()
	file, err := c.FormFile("file")
	if err != nil {
		util.ResError(c, err)
		return
	}

	category := c.PostForm("category")
	if category == "" {
		category = "avatar"
	}

	result, err := i.ImageBIZ.Upload(ctx, file, category)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags ImageAPI
// @Security BearerAuth
// @Summary Query image list
// @Param current query int false "pagination index" default(1)
// @Param pageSize query int false "pagination size" default(10)
// @Param url query string false "Image URL (fuzzy)"
// @Param category query string false "Category" Enums(article_cover,ad_image,friend_link_logo)
// @Param type query string false "MIME type" Enums(image/png,image/jpeg,image/gif)
// @Success 200 {object} util.ResponseResult{data=[]schema.Image}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/images [get]
func (i *Image) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.ImageQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := i.ImageBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags ImageAPI
// @Security BearerAuth
// @Summary Get image by ID
// @Param id path string true "Image ID"
// @Success 200 {object} util.ResponseResult{data=schema.Image}
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/images/{id} [get]
func (i *Image) Get(c *gin.Context) {
	ctx := c.Request.Context()
	image, err := i.ImageBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, image)
}

// ServeFile 重定向到图片的真实 URL（用于 <img src> 标签显示图片）
// GET /api/v1/images/:id/file
func (i *Image) ServeFile(c *gin.Context) {
	ctx := c.Request.Context()
	image, err := i.ImageBIZ.Get(ctx, c.Param("id"))
	if err != nil || image == nil {
		util.ResError(c, err)
		return
	}
	c.Redirect(302, image.URL)
}

// @Tags ImageAPI
// @Security BearerAuth
// @Summary Create image record
// @Description This only creates DB record. File upload should be handled separately.
// @Param body body schema.ImageForm true "Image form"
// @Success 200 {object} util.ResponseResult{data=schema.Image}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/images [post]
func (i *Image) Create(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.ImageForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := i.ImageBIZ.Create(ctx, form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags ImageAPI
// @Security BearerAuth
// @Summary Update image
// @Param id path string true "Image ID"
// @Param body body schema.ImageForm true "Image form"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/images/{id} [put]
func (i *Image) Update(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.ImageForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	err := i.ImageBIZ.Update(ctx, c.Param("id"), form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags ImageAPI
// @Security BearerAuth
// @Summary Delete image
// @Param id path string true "Image ID"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/images/{id} [delete]
func (i *Image) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := i.ImageBIZ.Delete(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags ImageAPI
// @Security BearerAuth
// @Summary Batch delete images
// @Param body body schema.BatchIDsForm true "Image IDs"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/images/batch-delete [delete]
func (i *Image) DeleteBatch(c *gin.Context) {
	ctx := c.Request.Context()
	var req struct {
		IDs []string `json:"ids" binding:"required,min=1"`
	}
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	err := i.ImageBIZ.DeleteByIds(ctx, req.IDs)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags ImageAPI
// @Summary Get images by category (public)
// @Param category path string true "Category" Enums(article_cover,ad_image,friend_link_logo)
// @Success 200 {object} util.ResponseResult{data=[]schema.Image}
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/images/category/{category} [get]
func (i *Image) GetByCategory(c *gin.Context) {
	ctx := c.Request.Context()
	category := c.Param("category")
	images, err := i.ImageBIZ.GetByCategory(ctx, category)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, images)
}
