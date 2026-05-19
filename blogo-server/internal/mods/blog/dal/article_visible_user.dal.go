// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

func getArticleVisibleUserDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ArticleVisibleUser))
}

type ArticleVisibleUser struct {
	DB *gorm.DB
}

// GetByArticleID 获取文章的所有可见用户列表
func (a *ArticleVisibleUser) GetByArticleID(ctx context.Context, articleID string) ([]*schema.ArticleVisibleUser, error) {
	var list []*schema.ArticleVisibleUser
	err := getArticleVisibleUserDB(ctx, a.DB).Where("article_id = ?", articleID).Find(&list).Error
	return list, errors.WithStack(err)
}

// BatchGetByArticleIDs 批量获取多篇文章的可见用户列表
func (a *ArticleVisibleUser) BatchGetByArticleIDs(ctx context.Context, articleIDs []string) (map[string][]*schema.ArticleVisibleUser, error) {
	if len(articleIDs) == 0 {
		return nil, nil
	}
	var list []*schema.ArticleVisibleUser
	err := getArticleVisibleUserDB(ctx, a.DB).Where("article_id IN ?", articleIDs).Find(&list).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	m := make(map[string][]*schema.ArticleVisibleUser, len(articleIDs))
	for _, v := range list {
		m[v.ArticleID] = append(m[v.ArticleID], v)
	}
	return m, nil
}

// BatchCreate 批量创建可见用户关联
func (a *ArticleVisibleUser) BatchCreate(ctx context.Context, list []*schema.ArticleVisibleUser) error {
	if len(list) == 0 {
		return nil
	}
	return errors.WithStack(getArticleVisibleUserDB(ctx, a.DB).Create(&list).Error)
}

// DeleteByArticleID 删除文章的所有可见用户关联
func (a *ArticleVisibleUser) DeleteByArticleID(ctx context.Context, articleID string) error {
	return errors.WithStack(
		getArticleVisibleUserDB(ctx, a.DB).Where("article_id = ?", articleID).Delete(new(schema.ArticleVisibleUser)).Error,
	)
}

// IsUserVisible 检查用户是否在文章的可见用户列表中
func (a *ArticleVisibleUser) IsUserVisible(ctx context.Context, articleID, userID string) (bool, error) {
	ok, err := util.Exists(ctx, getArticleVisibleUserDB(ctx, a.DB).Where("article_id = ? AND user_id = ?", articleID, userID))
	return ok, errors.WithStack(err)
}

// ValidateUsersExist 校验 userIDs 中的用户是否全部存在，返回不存在的用户 ID 列表
func (a *ArticleVisibleUser) ValidateUsersExist(ctx context.Context, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var existIDs []string
	err := util.GetDB(ctx, a.DB).
		Table(config.C.FormatTableName("user")).
		Select("id").
		Where("id IN ?", userIDs).
		Pluck("id", &existIDs).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	existSet := make(map[string]bool, len(existIDs))
	for _, id := range existIDs {
		existSet[id] = true
	}
	var missing []string
	for _, id := range userIDs {
		if !existSet[id] {
			missing = append(missing, id)
		}
	}
	return missing, nil
}
