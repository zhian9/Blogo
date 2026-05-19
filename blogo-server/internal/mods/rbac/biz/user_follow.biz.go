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

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/rbac/dal"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

type UserFollow struct {
	DB            *gorm.DB
	Trans         *util.Trans
	UserFollowDAL *dal.UserFollow
	UserDAL       *dal.User
}

func (uf *UserFollow) Follow(ctx context.Context, followerID string, form *schema.UserFollowForm) error {
	if followerID == form.FollowingID {
		return errors.BadRequest("", "Cannot follow yourself")
	}
	target, err := uf.UserDAL.Get(ctx, form.FollowingID)
	if err != nil {
		return err
	}
	if target == nil || target.Status != schema.UserStatusActivated {
		return errors.BadRequest(config.ErrBadRequest, "User not found or not activated")
	}
	exists, err := uf.UserFollowDAL.Exists(ctx, followerID, form.FollowingID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return uf.Trans.Exec(ctx, func(ctx context.Context) error {
		follow := &schema.UserFollow{FollowerID: followerID, FollowingID: form.FollowingID, CreatedAt: time.Now()}
		if err := uf.UserFollowDAL.Create(ctx, follow); err != nil {
			return err
		}
		if err := dal.IncFollowingCount(ctx, uf.DB, followerID); err != nil {
			return err
		}
		if err := dal.IncFollowerCount(ctx, uf.DB, form.FollowingID); err != nil {
			return err
		}
		return nil
	})
}

func (uf *UserFollow) Unfollow(ctx context.Context, followerID, followingID string) error {
	exists, err := uf.UserFollowDAL.Exists(ctx, followerID, followingID)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return uf.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := uf.UserFollowDAL.Delete(ctx, followerID, followingID); err != nil {
			return err
		}
		if err := dal.DecFollowingCount(ctx, uf.DB, followerID); err != nil {
			return err
		}
		if err := dal.DecFollowerCount(ctx, uf.DB, followingID); err != nil {
			return err
		}
		return nil
	})
}

func (uf *UserFollow) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	return uf.UserFollowDAL.Exists(ctx, followerID, followingID)
}

func (uf *UserFollow) ListFollowers(ctx context.Context, userID string, param util.PaginationParam) (schema.Users, *util.PaginationResult, error) {
	ids, page, err := uf.UserFollowDAL.ListFollowers(ctx, userID, param)
	if err != nil {
		return nil, nil, err
	}
	if len(ids) == 0 {
		return nil, page, nil
	}

	var users schema.Users
	if err := uf.DB.WithContext(ctx).Model(&schema.User{}).Where("id IN ?", ids).Omit("password").Find(&users).Error; err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return users, page, nil
}

func (uf *UserFollow) ListFollowing(ctx context.Context, userID string, param util.PaginationParam) (schema.Users, *util.PaginationResult, error) {
	ids, page, err := uf.UserFollowDAL.ListFollowing(ctx, userID, param)
	if err != nil {
		return nil, nil, err
	}
	if len(ids) == 0 {
		return nil, page, nil
	}

	var users schema.Users
	if err := uf.DB.WithContext(ctx).Model(&schema.User{}).Where("id IN ?", ids).Omit("password").Find(&users).Error; err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return users, page, nil
}
