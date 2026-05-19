// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// GetUserDB 根据上下文返回用户表的 GORM DB 实例
func GetUserDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.User))
}

// User 用户
type User struct {
	DB *gorm.DB
}

// Query 根据参数和选项查询用户列表
// 支持：
//   - 模糊查询
//   - 分页
//   - 排序
func (u *User) Query(ctx context.Context, params schema.UserQueryParam, opts ...schema.UserQueryOptions) (*schema.UserQueryResult, error) {
	// 解析查询选项
	var opt schema.UserQueryOptions
	if len(opts) > 0 {
		opt = opts[0] //只取第一个参数的原因是，根据单一配置原则：一个查询只需要一套配置选项，同时向后兼容
	}

	//构建基础查询
	db := GetUserDB(ctx, u.DB)

	// 条件查询
	if v := params.LikeUsername; len(v) > 0 {
		db = db.Where("username Like ?", "%"+v+"%") // 用户名模糊查询
	}
	if v := params.LikeName; len(v) > 0 {
		db = db.Where("name LIKE ?", "%"+v+"%") // 姓名模糊查询
	}
	if v := params.Status; len(v) > 0 {
		db = db.Where("status = ?", v) //状态精确匹配
	}

	// 执行分页查询
	var list schema.Users
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err) //保留错误堆栈
	}

	//返回查询结果
	return &schema.UserQueryResult{
		PageResult: pageResult,
		Data:       list,
	}, nil
}

// Get 根据 ID 获取单个用户
func (u *User) Get(ctx context.Context, id string, opts ...schema.UserQueryOptions) (*schema.User, error) {
	var opt schema.UserQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	user := new(schema.User)

	ok, err := util.FindOne(ctx, GetUserDB(ctx, u.DB).Where("id=?", id), opt.QueryOptions, user)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil //未找到
	}
	return user, err
}

// GetByUsername 根据用户名获取单个用户
func (u *User) GetByUsername(ctx context.Context, username string, opts ...schema.UserQueryOptions) (*schema.User, error) {
	var opt schema.UserQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	user := new(schema.User)
	ok, err := util.FindOne(ctx, GetUserDB(ctx, u.DB).Where("username=?", username), opt.QueryOptions, user)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil //未找到
	}
	return user, nil
}

// ExistsID 检查用户 ID 是否存在
func (u *User) ExistsID(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetUserDB(ctx, u.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

func (a *User) ExistsUsername(ctx context.Context, username string) (bool, error) {
	ok, err := util.Exists(ctx, GetUserDB(ctx, a.DB).Where("username=?", username))
	return ok, errors.WithStack(err)
}

// Create 创建新用户
func (u *User) Create(ctx context.Context, user *schema.User) error {
	result := GetUserDB(ctx, u.DB).Create(user)
	return errors.WithStack(result.Error)
}

// Update 更新用户信息
func (u *User) Update(ctx context.Context, user *schema.User, selectFields ...string) error {
	db := GetUserDB(ctx, u.DB).Where("id=?", user.ID)
	if len(selectFields) > 0 {
		db = db.Select(selectFields) //更新指定字段
	} else {
		db = db.Select("*").Omit("created_at") //更新所有字段，但排除 created_at
	}
	result := db.Updates(user)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除用户
func (u *User) Delete(ctx context.Context, id string) error {
	result := GetUserDB(ctx, u.DB).Where("id=?", id).Delete(new(schema.User))
	return errors.WithStack(result.Error)
}

// UpdatePasswordByID 仅更新用户密码。
// 用于密码修改场景，避免更新其他字段。
func (a *User) UpdatePasswordByID(ctx context.Context, id string, password string) error {
	result := GetUserDB(ctx, a.DB).Where("id=?", id).Select("password").Updates(schema.User{Password: password})
	return errors.WithStack(result.Error)
}

// GetByActivationToken 根据激活 token 查找用户。
func (u *User) GetByActivationToken(ctx context.Context, token string) (*schema.User, error) {
	user := new(schema.User)
	ok, err := util.FindOne(ctx, GetUserDB(ctx, u.DB).Where("activation_token=?", token), util.QueryOptions{}, user)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return user, nil
}

// ActivateUser 激活用户：设置状态为 activated、清除激活 token、记录激活时间。
func (u *User) ActivateUser(ctx context.Context, id string, activatedAt time.Time) error {
	result := GetUserDB(ctx, u.DB).Where("id=?", id).Select("status", "activation_token", "activated_at").Updates(schema.User{
		Status:          schema.UserStatusActivated,
		ActivationToken: "",
		ActivatedAt:     &activatedAt,
	})
	return errors.WithStack(result.Error)
}

// UpdateLastLogin 原子更新用户最后登录时间和 IP。
// 独立于其他更新操作，避免覆盖密码等敏感字段。
func (a *User) UpdateLastLogin(ctx context.Context, id string, loginAt time.Time, loginIP string) error {
	result := GetUserDB(ctx, a.DB).
		Where("id = ?", id).
		Select("last_login_at", "last_login_ip").
		Updates(schema.User{
			LastLoginAt: &loginAt,
			LastLoginIP: loginIP,
		})
	return errors.WithStack(result.Error)
}
