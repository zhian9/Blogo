// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package config

// 缓存命名空间
const (
	CacheNSForUser = "user" // 用户缓存命名空间
	CacheNSForRole = "role" // 角色缓存命名空间
)

// 缓存键
const (
	CacheKeyForSyncToCasbin = "sync:casbin" // Casbin 同步缓存键
)

// 认证相关错误码
const (
	ErrInvalidTokenID            = "com.invalid.token"                // 无效的 Token ID
	ErrInvalidCaptchaID          = "com.invalid.captcha"              // 验证码错误
	ErrInvalidUsernameOrPassword = "com.invalid.username-or-password" // 用户名或密码错误
)

// 全局通用错误码
const (
	ErrInternalError     = "com.global.internal_error"      // 服务器内部错误
	ErrNotFound          = "com.global.not_found"           // 资源不存在
	ErrForbidden         = "com.global.forbidden"           // 权限不足
	ErrBadRequest        = "com.global.bad_request"         // 请求参数错误
	ErrRateLimitExceeded = "com.global.rate_limit_exceeded" // 请求频率过高
)

// 用户模块错误码
const (
	ErrUserDuplicateUsername = "com.user.duplicate_username" // 用户名已存在
	ErrUserNotFound          = "com.user.not_found"          // 用户不存在
	ErrUserPasswordMismatch  = "com.user.password_mismatch"  // 密码错误
	ErrUserEmailInUse        = "com.user.email_in_use"       // 邮箱已被使用
	ErrUserInactive          = "com.user.inactive"           // 用户已被禁用
)

// 文章模块错误码
const (
	ErrArticleNotFound          = "com.article.not_found"           // 文章不存在
	ErrArticleSlugExists        = "com.article.slug_exists"         // 文章 Slug 已存在
	ErrArticleDraftNotPublished = "com.article.draft_not_published" // 草稿文章不可公开访问
	ErrArticleCategoryNotFound  = "com.article.category_not_found"  // 文章分类不存在
	ErrArticleTagNotFound       = "com.article.tag_not_found"       // 文章标签不存在
	ErrArticlePermissionDenied  = "com.article.permission_denied"   // 无权限访问该文章
	ErrArticleVisibleUsersEmpty = "com.article.visible_users_empty" // partial_visible 需指定可见用户
)

// 评论模块错误码
const (
	ErrCommentNotFound          = "com.comment.not_found"           // 评论不存在
	ErrCommentParentNotFound    = "com.comment.parent_not_found"    // 父评论不存在
	ErrCommentTooLong           = "com.comment.too_long"            // 评论内容过长
	ErrCommentSpamDetected      = "com.comment.spam_detected"       // 检测到垃圾评论
	ErrCommentGuestNameRequired = "com.comment.guest_name_required" // 游客评论需填写名称
	ErrCommentInvalidField      = "com.comment.invalid_field"       // 非法属性
)

// 通知模块错误码
const (
	ErrNotificationNotFound     = "com.notification.not_found"     // 通知不存在
	ErrNotificationUserRequired = "com.notification.user_required" // 用户 ID 必须提供
	ErrNotificationAlreadyRead  = "com.notification.already_read"  // 通知已是已读状态
)

// 友情链接模块错误码
const (
	ErrFriendLinkNotFound      = "com.friend_link.not_found"      // 友情链接不存在
	ErrFriendLinkURLInvalid    = "com.friend_link.url_invalid"    // 友链 URL 格式无效
	ErrFriendLinkDuplicateName = "com.friend_link.duplicate_name" // 友链名称重复
)

// 图片模块错误码
const (
	ErrImageNotFound     = "com.image.not_found"     // 图片不存在
	ErrImageUploadFailed = "com.image.upload_failed" // 图片上传失败
	ErrImageInvalidType  = "com.image.invalid_type"  // 不支持的图片类型
	ErrImageTooLarge     = "com.image.too_large"     // 图片文件过大
)

// 系统设置模块错误码
const (
	ErrSettingKeyRequired = "com.setting.key_required" // 配置项 Key 必须提供
	ErrSettingKeyInvalid  = "com.setting.key_invalid"  // 配置项 Key 格式无效（仅允许字母、数字、下划线）
	ErrSettingNotFound    = "com.setting.not_found"    // 配置项不存在
)

// 分类模块错误码
const (
	ErrCategoryNotFound   = "com.category.not_found"   // 分类不存在
	ErrCategoryNameExists = "com.category.name_exists" // 分类名称已存在
)

// 标签模块错误码
const (
	ErrTagNotFound   = "com.tag.not_found"   // 标签不存在
	ErrTagNameExists = "com.tag.name_exists" // 标签名称已存在
)

// 页面模块错误码
const (
	ErrPageNotFound   = "com.page.not_found"   // 页面不存在
	ErrPageSlugExists = "com.page.slug_exists" // 页面 Slug 已存在
)
