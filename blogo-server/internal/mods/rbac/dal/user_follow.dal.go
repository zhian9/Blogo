// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
// Project repository: https://github.com/zhian9/go-admin

package dal

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

func GetUserFollowDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.UserFollow))
}

type UserFollow struct {
	DB *gorm.DB
}

// Create 创建关注记录
func (uf *UserFollow) Create(ctx context.Context, follow *schema.UserFollow) error {
	result := GetUserFollowDB(ctx, uf.DB).Create(follow)
	return errors.WithStack(result.Error)
}

// Delete 删除关注记录（取消关注）
func (uf *UserFollow) Delete(ctx context.Context, followerID, followingID string) error {
	result := GetUserFollowDB(ctx, uf.DB).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(new(schema.UserFollow))
	return errors.WithStack(result.Error)
}

// Exists 检查是否已关注
func (uf *UserFollow) Exists(ctx context.Context, followerID, followingID string) (bool, error) {
	ok, err := util.Exists(ctx, GetUserFollowDB(ctx, uf.DB).
		Where("follower_id = ? AND following_id = ?", followerID, followingID))
	return ok, errors.WithStack(err)
}

// CountFollowers 统计粉丝数
func (uf *UserFollow) CountFollowers(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := GetUserFollowDB(ctx, uf.DB).Where("following_id = ?", userID).Count(&count).Error
	return count, errors.WithStack(err)
}

// CountFollowing 统计关注数
func (uf *UserFollow) CountFollowing(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := GetUserFollowDB(ctx, uf.DB).Where("follower_id = ?", userID).Count(&count).Error
	return count, errors.WithStack(err)
}

// ListFollowers 粉丝列表（分页）
func (uf *UserFollow) ListFollowers(ctx context.Context, userID string, param util.PaginationParam) ([]string, *util.PaginationResult, error) {
	var ids []string
	db := GetUserFollowDB(ctx, uf.DB).Where("following_id = ?", userID).Order("created_at DESC")
	pageResult, err := util.WrapPageQuery(ctx, db.Select("follower_id"), param, util.QueryOptions{}, &ids)
	return ids, pageResult, errors.WithStack(err)
}

// ListFollowing 关注列表（分页）
func (uf *UserFollow) ListFollowing(ctx context.Context, userID string, param util.PaginationParam) ([]string, *util.PaginationResult, error) {
	var ids []string
	db := GetUserFollowDB(ctx, uf.DB).Where("follower_id = ?", userID).Order("created_at DESC")
	pageResult, err := util.WrapPageQuery(ctx, db.Select("following_id"), param, util.QueryOptions{}, &ids)
	return ids, pageResult, errors.WithStack(err)
}

// IncFollowerCount 原子增加粉丝数
func IncFollowerCount(ctx context.Context, db *gorm.DB, userID string) error {
	return errors.WithStack(GetUserDB(ctx, db).Where("id = ?", userID).
		Update("follower_count", gorm.Expr("follower_count + 1")).Error)
}

// DecFollowerCount 原子减少粉丝数
func DecFollowerCount(ctx context.Context, db *gorm.DB, userID string) error {
	return errors.WithStack(GetUserDB(ctx, db).Where("id = ?", userID).
		Update("follower_count", gorm.Expr("GREATEST(follower_count - 1, 0)")).Error)
}

// IncFollowingCount 原子增加关注数
func IncFollowingCount(ctx context.Context, db *gorm.DB, userID string) error {
	return errors.WithStack(GetUserDB(ctx, db).Where("id = ?", userID).
		Update("following_count", gorm.Expr("following_count + 1")).Error)
}

// DecFollowingCount 原子减少关注数
func DecFollowingCount(ctx context.Context, db *gorm.DB, userID string) error {
	return errors.WithStack(GetUserDB(ctx, db).Where("id = ?", userID).
		Update("following_count", gorm.Expr("GREATEST(following_count - 1, 0)")).Error)
}
