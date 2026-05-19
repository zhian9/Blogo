// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/util"
)

// ArticlePermission 文章阅读权限服务（业务权限，不写入 Casbin RBAC）
type ArticlePermission struct {
	ArticleVisibleUserDAL *dal.ArticleVisibleUser
}

// CanUserReadArticle 判断指定用户是否有权阅读指定文章。
//
// 规则：
//   - public：任何人可读
//   - private：仅作者本人 + admin（root）可读
//   - partial_visible：作者本人 + admin + visible_users 列表中的用户可读
//
// userID 为空表示游客。
func (p *ArticlePermission) CanUserReadArticle(ctx context.Context, userID string, article *schema.Article) bool {
	if article == nil {
		return false
	}

	switch article.Visibility {
	case schema.ArticleVisibilityPublic:
		return true

	case schema.ArticleVisibilityPrivate:
		return p.isAuthorOrAdmin(ctx, userID, article)

	case schema.ArticleVisibilityPartialVisible:
		return p.canReadPartialVisible(ctx, userID, article)

	default:
		// 未知 visibility — 安全策略：仅作者/admin 可读
		return p.isAuthorOrAdmin(ctx, userID, article)
	}
}

// isAuthorOrAdmin 判断用户是否是文章作者或超级管理员
func (p *ArticlePermission) isAuthorOrAdmin(ctx context.Context, userID string, article *schema.Article) bool {
	if userID == "" {
		return false
	}
	if article.AuthorID == userID {
		return true
	}
	if util.FromIsRootUser(ctx) {
		return true
	}
	return false
}

// canReadPartialVisible 判断用户是否可以阅读部分可见文章
func (p *ArticlePermission) canReadPartialVisible(ctx context.Context, userID string, article *schema.Article) bool {
	if userID == "" {
		return false
	}
	// 作者本人或 admin 永远可见
	if p.isAuthorOrAdmin(ctx, userID, article) {
		return true
	}
	// 检查 visible_users 列表
	ok, _ := p.ArticleVisibleUserDAL.IsUserVisible(ctx, article.ID, userID)
	return ok
}

// FilterVisibleArticleIDs 从一批文章 ID 中筛选出用户有权限阅读的 ID。
// 用于列表查询时过滤 partial_visible 文章。
// 返回 allowed 集合（map 的 key 为可读的 articleID）。
func (p *ArticlePermission) FilterVisibleArticleIDs(ctx context.Context, userID string, articles schema.Articles) map[string]bool {
	allowed := make(map[string]bool, len(articles))
	if len(articles) == 0 {
		return allowed
	}

	// 收集需要查表的 partial_visible 文章 ID（非作者、非 admin）
	var needCheckIDs []string
	for _, art := range articles {
		if art == nil {
			continue
		}
		switch art.Visibility {
		case schema.ArticleVisibilityPublic:
			allowed[art.ID] = true
		case schema.ArticleVisibilityPrivate:
			if p.isAuthorOrAdmin(ctx, userID, art) {
				allowed[art.ID] = true
			}
		case schema.ArticleVisibilityPartialVisible:
			if p.isAuthorOrAdmin(ctx, userID, art) {
				allowed[art.ID] = true
			} else if userID != "" {
				needCheckIDs = append(needCheckIDs, art.ID)
			}
		}
	}

	if len(needCheckIDs) == 0 || userID == "" {
		return allowed
	}

	// 批量查询 visible_users
	visibleMap, err := p.ArticleVisibleUserDAL.BatchGetByArticleIDs(ctx, needCheckIDs)
	if err != nil {
		return allowed
	}
	for _, artID := range needCheckIDs {
		for _, vu := range visibleMap[artID] {
			if vu.UserID == userID {
				allowed[artID] = true
				break
			}
		}
	}
	return allowed
}
