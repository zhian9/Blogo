// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package blog

import (
	"github.com/google/wire"
	"github.com/zhian9/blogo-server/internal/mods/blog/api"
	"github.com/zhian9/blogo-server/internal/mods/blog/biz"
	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/msg"
)

// Set 是博客模块的 Wire 依赖注入集合。
var Set = wire.NewSet(
	wire.Struct(new(Blog), "*"), // 聚合根

	// Article 相关
	wire.Struct(new(dal.Article), "*"),
	wire.Struct(new(biz.Article), "*"),
	wire.Struct(new(api.Article), "*"),

	// Category 相关
	wire.Struct(new(dal.Category), "*"),
	wire.Struct(new(biz.Category), "*"),
	wire.Struct(new(api.Category), "*"),

	// Tag 相关
	wire.Struct(new(dal.Tag), "*"),
	wire.Struct(new(biz.Tag), "*"),
	wire.Struct(new(api.Tag), "*"),

	// Article-Tag 中间表
	wire.Struct(new(dal.ArticleTag), "*"),

	// ArticleVisibleUser 可见用户关联
	wire.Struct(new(dal.ArticleVisibleUser), "*"),

	// ArticlePermission 文章阅读权限
	wire.Struct(new(biz.ArticlePermission), "*"),

	// Contribution 贡献数据
	wire.Struct(new(dal.Contribution), "*"),

	// Comment 相关
	wire.Struct(new(dal.Comment), "*"),
	wire.Struct(new(biz.Comment), "*"),
	wire.Struct(new(api.Comment), "*"),

	// Page 相关
	wire.Struct(new(dal.Page), "*"),
	wire.Struct(new(biz.Page), "*"),
	wire.Struct(new(api.Page), "*"),

	// FriendLink 相关
	wire.Struct(new(dal.FriendLink), "*"),
	wire.Struct(new(biz.FriendLink), "*"),
	wire.Struct(new(api.FriendLink), "*"),

	// Setting 相关
	wire.Struct(new(dal.Setting), "*"),
	wire.Struct(new(biz.Setting), "*"),
	wire.Struct(new(api.Setting), "*"),

	// Notification 相关
	wire.Struct(new(dal.Notification), "*"),
	wire.Struct(new(biz.Notification), "*"),
	wire.Struct(new(api.Notification), "*"),

	// Image 相关
	wire.Struct(new(dal.Image), "*"),
	wire.Struct(new(biz.Image), "*"),
	wire.Struct(new(api.Image), "*"),

	// Statistics 相关
	wire.Struct(new(dal.Statistics), "*"),
	wire.Struct(new(biz.Statistics), "*"),
	wire.Struct(new(api.Statistics), "*"),

	// ArticleLike 相关
	wire.Struct(new(dal.ArticleLike), "*"),
	wire.Struct(new(biz.ArticleLike), "*"),
	wire.Struct(new(api.ArticleLike), "*"),

	// ArticleFavorite 相关
	wire.Struct(new(dal.ArticleFavorite), "*"),
	wire.Struct(new(biz.ArticleFavorite), "*"),
	wire.Struct(new(api.ArticleFavorite), "*"),

	// Subscriber 相关
	wire.Struct(new(dal.Subscriber), "*"),
	wire.Struct(new(biz.Subscriber), "*"),
	wire.Struct(new(api.Subscriber), "*"),

	// MailWorker (constructor — avoid exposing unexported fields to Wire)
	msg.NewMailWorker,

	// Profile 相关（跨模块聚合）
	wire.Struct(new(biz.Profile), "*"),
	wire.Struct(new(api.Profile), "*"),
	// SEO 相关
	wire.Struct(new(api.SEO), "*"),
)
