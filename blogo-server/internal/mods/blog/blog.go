// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package blog

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/blog/api"
	"github.com/zhian9/blogo-server/internal/mods/blog/msg"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// Blog 是博客模块的核心聚合对象，包含所有子组件。
type Blog struct {
	DB              *gorm.DB             // 数据库连接
	ArticleAPI      *api.Article         // 文章 API 控制器
	CategoryAPI     *api.Category        // 分类 API 控制器
	TagAPI          *api.Tag             // 标签 API 控制器
	CommentAPI      *api.Comment         // 评论 API 控制器
	PageAPI         *api.Page            // 页面 API 控制器
	FriendLinkAPI   *api.FriendLink      // 友情链接 API 控制器
	SettingAPI      *api.Setting         // 系统配置 API 控制器
	NotificationAPI *api.Notification    // 通知 API 控制器
	ImageAPI        *api.Image           // 图片 API 控制器
	StatisticsAPI   *api.Statistics      // 统计 API 控制器
	ArticleLikeAPI  *api.ArticleLike     // 点赞 API 控制器
	FavoriteAPI     *api.ArticleFavorite // 收藏 API 控制器
	ProfileAPI      *api.Profile         // 个人主页 API 控制器
	SEOAPI          *api.SEO             // SEO API 控制器
	SubscriberAPI   *api.Subscriber      // 订阅者 API 控制器
	MailWorker      *msg.MailWorker      // 异步邮件 Worker
}

// AutoMigrate 自动创建或更新博客模块相关数据库表。
func (b *Blog) AutoMigrate(ctx context.Context) error {
	return b.DB.AutoMigrate(
		new(schema.Article),
		new(schema.Category),
		new(schema.Tag),
		new(schema.Comment),
		new(schema.Page),
		new(schema.FriendLink),
		new(schema.Setting),
		new(schema.Notification),
		new(schema.Image),
		new(schema.Statistics),
		new(schema.ArticleTag),         // 文章-标签中间表
		new(schema.ArticleLike),        // 点赞表
		new(schema.ArticleFavorite),    // 收藏表
		new(schema.Subscriber),         // 邮件订阅者表
		new(schema.ArticleVisibleUser), // 文章可见用户关联表
		new(schema.UserContribution),   // 用户贡献记录表
	)
}

// Init 初始化博客模块。
// - 执行自动迁移（如果启用）
// - 初始化默认页面（如“关于我”）
// - 初始化默认系统配置（如站点标题、SEO 信息）
func (b *Blog) Init(ctx context.Context) error {
	// 1. 自动迁移（受全局开关控制）
	if config.C.Storage.DB.AutoMigrate {
		if err := b.AutoMigrate(ctx); err != nil {
			return err
		}
	}

	// 2. 初始化默认页面（仅当页面表为空时）
	if err := b.initDefaultPages(ctx); err != nil {
		return err
	}

	// 3. 初始化默认系统配置（仅当配置表为空时）
	if err := b.initDefaultSettings(ctx); err != nil {
		return err
	}

	// 4. 启动异步邮件 Worker
	b.MailWorker.Start(ctx)

	return nil
}

// initDefaultPages 初始化默认页面（如“关于我”）
func (b *Blog) initDefaultPages(ctx context.Context) error {
	var count int64
	if err := b.DB.Model(&schema.Page{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // 已有页面，跳过
	}

	defaultPages := []schema.Page{
		{
			ID:          util.NewXID(),
			Title:       "关于我",
			Slug:        "about",
			Content:     "这是关于我的介绍页面，你可以在这里写一些个人简介、技术栈、项目经历等。",
			IsPublished: true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          util.NewXID(),
			Title:       "隐私政策",
			Slug:        "privacy",
			Content:     "本网站尊重并保护所有用户的个人隐私权。",
			IsPublished: true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	return b.DB.CreateInBatches(defaultPages, 10).Error
}

// initDefaultSettings 初始化默认系统配置
func (b *Blog) initDefaultSettings(ctx context.Context) error {
	var count int64
	if err := b.DB.Model(&schema.Setting{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // 已有配置，跳过
	}

	defaultSettings := []schema.Setting{
		{Key: "site_title", Value: "Blogo", Description: "网站标题"},
		{Key: "site_description", Value: "记录技术成长的点滴", Description: "网站描述"},
		{Key: "site_keywords", Value: "Go,博客,技术分享", Description: "SEO 关键词"},
		{Key: "icp_license", Value: "京ICP备12345678号", Description: "备案号"},
		{Key: "github_url", Value: "https://github.com/zhian9", Description: "GitHub 链接"},
	}

	return b.DB.CreateInBatches(defaultSettings, 10).Error
}

// RegisterV1Routers 注册博客模块的 V1 版本 API 路由。
// 分为：管理接口（需认证）和公开接口（无需认证）。
func (b *Blog) RegisterV1Routers(ctx context.Context, v1 *gin.RouterGroup) error {
	// ========== 公开接口（无需认证）==========
	public := v1.Group("")
	{
		// 文章
		public.GET("/articles/slug/:slug", b.ArticleAPI.GetBySlug)
		public.GET("/archives", b.ArticleAPI.GetArchive)
		public.POST("/articles/:id/views", b.ArticleAPI.IncViews)

		// 评论

		// 页面
		public.GET("/users/:id/profile", b.ProfileAPI.GetDashboard)
		public.GET("/pages/slug/:slug", b.PageAPI.GetBySlug)

		// 友情链接
		public.GET("/friend-links/all", b.FriendLinkAPI.GetAll)
		public.GET("/seo/articles/:slug", b.SEOAPI.ArticleMeta)

		// 标签 & 分类（公开列表）
		public.GET("/tags/all", b.TagAPI.GetAll)
		public.GET("/categories/all", b.CategoryAPI.GetAll)

		// 统计
		public.GET("/statistics/latest", b.StatisticsAPI.GetLatest)

		// 图片
		public.GET("/images/category/:category", b.ImageAPI.GetByCategory)
		public.GET("/images/:id/file", b.ImageAPI.ServeFile)

		// 订阅
		public.POST("/subscribe", b.SubscriberAPI.Subscribe)
		public.GET("/subscribe/unsubscribe", b.SubscriberAPI.UnsubscribeByEmail)
	}

	// ========== 管理接口（需认证，由中间件保护）==========
	// 文章管理
	article := v1.Group("/articles")
	{
		article.GET("", b.ArticleAPI.Query)
		article.GET("/:id", b.ArticleAPI.Get)
		article.POST("", b.ArticleAPI.Create)
		article.PUT("/:id", b.ArticleAPI.Update)
		article.DELETE("/:id", b.ArticleAPI.Delete)
		article.PATCH("/status", b.ArticleAPI.UpdateStatus)
		article.PATCH("/:id/top", b.ArticleAPI.ToggleTop)
	}

	// 评论管理
	comment := v1.Group("/comments")
	{
		comment.GET("", b.CommentAPI.Query)
		comment.GET("/stats", b.CommentAPI.Stats)
		comment.GET("/:id", b.CommentAPI.Get)
		comment.PUT("/:id", b.CommentAPI.Update)
		comment.DELETE("/:id", b.CommentAPI.Delete)
		comment.PATCH("/:id/approve", b.CommentAPI.Approve)
		comment.PATCH("/:id/reject", b.CommentAPI.Reject)
	}

	// 分类管理
	category := v1.Group("/categories")
	{
		category.GET("", b.CategoryAPI.Query)
		category.GET("/:id", b.CategoryAPI.Get)
		category.POST("", b.CategoryAPI.Create)
		category.PUT("/:id", b.CategoryAPI.Update)
		category.DELETE("/:id", b.CategoryAPI.Delete)
	}

	// 标签管理
	tag := v1.Group("/tags")
	{
		tag.GET("", b.TagAPI.Query)
		tag.GET("/:id", b.TagAPI.Get)
		tag.POST("", b.TagAPI.Create)
		tag.PUT("/:id", b.TagAPI.Update)
		tag.DELETE("/:id", b.TagAPI.Delete)
	}

	// 页面管理
	page := v1.Group("/pages")
	{
		page.GET("", b.PageAPI.Query)
		page.GET("/:id", b.PageAPI.Get)
		page.POST("", b.PageAPI.Create)
		page.PUT("/:id", b.PageAPI.Update)
		page.DELETE("/:id", b.PageAPI.Delete)
	}

	// 友情链接管理
	friendLink := v1.Group("/friend-links")
	{
		friendLink.GET("", b.FriendLinkAPI.Query)
		friendLink.GET("/:id", b.FriendLinkAPI.Get)
		friendLink.POST("", b.FriendLinkAPI.Create)
		friendLink.PUT("/:id", b.FriendLinkAPI.Update)
		friendLink.DELETE("/:id", b.FriendLinkAPI.Delete)
	}

	// 系统配置管理
	setting := v1.Group("/settings")
	{
		setting.GET("", b.SettingAPI.Query)
		setting.GET("/:key", b.SettingAPI.Get)
		setting.POST("", b.SettingAPI.Create)
		setting.PUT("/:key", b.SettingAPI.Update)
		setting.DELETE("/:key", b.SettingAPI.Delete)
		setting.GET("/all", b.SettingAPI.GetAll)
	}

	// 通知管理
	notification := v1.Group("/notifications")
	{
		notification.GET("", b.NotificationAPI.Query)
		notification.GET("/:id", b.NotificationAPI.Get)
		notification.DELETE("/:id", b.NotificationAPI.Delete)
		notification.PATCH("/:id/read", b.NotificationAPI.MarkAsRead)
		notification.PATCH("/read-all", b.NotificationAPI.MarkAllAsRead)
	}

	// 图片管理
	image := v1.Group("/images")
	{
		image.GET("", b.ImageAPI.Query)
		image.GET("/:id", b.ImageAPI.Get)
		image.POST("", b.ImageAPI.Create)
		image.POST("/upload", b.ImageAPI.Upload)
		image.PUT("/:id", b.ImageAPI.Update)
		image.DELETE("/:id", b.ImageAPI.Delete)
		image.DELETE("/batch-delete", b.ImageAPI.DeleteBatch)
	}

	// 统计数据（只读）
	stat := v1.Group("/statistics")
	{
		stat.GET("", b.StatisticsAPI.Query)
		stat.GET("/:date", b.StatisticsAPI.Get)
	}

	// 控制中心（只读）
	dashboard := v1.Group("/dashboard")
	{
		dashboard.GET("/traffic", b.StatisticsAPI.GetTraffic)
	}

	// 点赞（公开获取状态 + 需认证操作）
	like := v1.Group("/articles")
	{
		like.GET("/:id/like", b.ArticleLikeAPI.GetLikeStatus)
		like.POST("/:id/like", b.ArticleLikeAPI.Like)
		like.DELETE("/:id/like", b.ArticleLikeAPI.UnLike)
	}

	// 收藏（需认证）
	fav := v1.Group("/favorites")
	{
		fav.GET("", b.FavoriteAPI.Query)
		fav.GET("/count", b.FavoriteAPI.Count)
		fav.GET("/:article_id", b.FavoriteAPI.IsFavorite)
		fav.POST("", b.FavoriteAPI.Create)
		fav.DELETE("/:article_id", b.FavoriteAPI.Delete)
	}

	return nil
}

// Release 释放博客模块占用的资源。
func (b *Blog) Release(ctx context.Context) error {
	b.MailWorker.Stop()
	return nil
}
