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
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
)

type Comment struct {
	CommentBIZ *biz.Comment
}

// @Tags CommentAPI
// @Security BearerAuth
// @Summary Query comment list (admin)
// @Param current query int false "pagination index" default(1)
// @Param pageSize query int false "pagination size" default(10)
// @Param article_id query string false "Article ID"
// @Param user_id query string false "User ID"
// @Param status query string false "Status" Enums(approved,pending,rejected)
// @Param parent_id query string false "Parent comment ID"
// @Param is_top query boolean false "Is top"
// @Param created_at_gte query string false "Created at >= (YYYY-MM-DD)"
// @Param created_at_lte query string false "Created at <= (YYYY-MM-DD)"
// @Success 200 {object} util.ResponseResult{data=[]schema.Comment}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/comments [get]
func (c *Comment) Query(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	var params schema.CommentQueryParam
	if err := util.ParseQuery(cxt, &params); err != nil {
		util.ResError(cxt, err)
		return
	}

	result, err := c.CommentBIZ.Query(ctx, params)
	if err != nil {
		logging.Context(ctx).Error("comment query failed", zap.Error(err), zap.Any("params", params))
		util.ResError(cxt, err)
		return
	}
	util.ResPage(cxt, result.Data, result.PageResult)
}

// @Tags CommentAPI
// @Security BearerAuth
// @Summary Get comment by ID
// @Param id path string true "Comment ID"
// @Success 200 {object} util.ResponseResult{data=schema.Comment}
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/comments/{id} [get]
func (c *Comment) Get(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	comment, err := c.CommentBIZ.Get(ctx, cxt.Param("id"))
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResSuccess(cxt, comment)
}

// @Tags CommentAPI
// @Summary Create comment (public)
// @Description Create a new comment, supports guest and logged-in users.
// @Param body body schema.CommentForm true "Comment form"
// @Success 200 {object} util.ResponseResult{data=schema.Comment}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/comments [post]
// GetByArticleID 获取文章的所有已通过评论（公开接口，供博客前端使用）
// @Router /api/v1/articles/{id}/comments [get]
func (c *Comment) GetByArticleID(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	comments, err := c.CommentBIZ.GetByArticleID(ctx, cxt.Param("id"))
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResSuccess(cxt, comments)
}

func (c *Comment) Create(cxt *gin.Context) {
	ctx := cxt.Request.Context()

	userID := util.FromUserID(ctx)
	form := new(schema.CommentForm)
	if err := util.ParseJSON(cxt, form); err != nil {
		util.ResError(cxt, err)
		return
	}
	form.UserID = userID

	if err := form.Validate(); err != nil {
		util.ResError(cxt, err)
		return
	}

	// 获取客户端 IP 和 User-Agent
	ip := cxt.ClientIP()
	userAgent := cxt.GetHeader("User-Agent")

	result, err := c.CommentBIZ.Create(ctx, form, ip, userAgent)
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResSuccess(cxt, result)
}

// @Tags CommentAPI
// @Security BearerAuth
// @Summary Update comment by ID
// @Param id path string true "Comment ID"
// @Param body body map[string]interface{} true "Fields to update (e.g. {\"content\": \"new\", \"status\": \"approved\"})"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/comments/{id} [put]
func (c *Comment) Update(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	var updates map[string]interface{}
	if err := util.ParseJSON(cxt, &updates); err != nil {
		util.ResError(cxt, err)
		return
	}

	// 白名单校验（可选，Biz 层已做）
	allowed := map[string]bool{
		"content": true,
		"status":  true,
		"is_top":  true,
	}
	for k := range updates {
		if !allowed[k] {
			util.ResError(cxt, errors.NewBadRequest("Invalid field: %s", k))
			return
		}
	}

	err := c.CommentBIZ.Update(ctx, cxt.Param("id"), updates)
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResOK(cxt)
}

// @Tags CommentAPI
// @Security BearerAuth
// @Summary Delete comment by ID
// @Param id path string true "Comment ID"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/comments/{id} [delete]
func (c *Comment) Delete(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	err := c.CommentBIZ.Delete(ctx, cxt.Param("id"))
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResOK(cxt)
}

// @Tags CommentAPI
// @Security BearerAuth
// @Summary Approve comment
// @Param id path string true "Comment ID"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/comments/{id}/approve [patch]
func (c *Comment) Approve(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	err := c.CommentBIZ.Approve(ctx, cxt.Param("id"))
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResOK(cxt)
}

// @Tags CommentAPI
// @Security BearerAuth
// @Summary Reject comment
// @Param id path string true "Comment ID"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/comments/{id}/reject [patch]
func (c *Comment) Reject(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	err := c.CommentBIZ.Reject(ctx, cxt.Param("id"))
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResOK(cxt)
}

// Stats 全站评论统计
// GET /api/v1/comments/stats
func (c *Comment) Stats(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	stats, err := c.CommentBIZ.Stats(ctx)
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResSuccess(cxt, stats)
}
