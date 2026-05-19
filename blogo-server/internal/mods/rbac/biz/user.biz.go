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
	"github.com/zhian9/blogo-server/pkg/cachex"
	"github.com/zhian9/blogo-server/pkg/crypto/hash"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// User 是用户管理业务的核心对象，聚合了缓存、事务、数据访问等依赖。
type User struct {
	Cache       cachex.Cacher // 缓存客户端（用于用户角色缓存）
	Trans       *util.Trans   // 事务管理器
	UserDAL     *dal.User     // 用户数据访问层
	UserRoleDAL *dal.UserRole // 用户角色数据访问层
}

// Query 查询用户列表（带分页和角色信息）。
// 流程：
//  1. 强制启用分页
//  2. 查询用户（排除密码字段）
//  3. 批量查询用户角色（减少 N+1 查询）
//  4. 关联角色到用户
func (u *User) Query(ctx context.Context, params schema.UserQueryParam) (*schema.UserQueryResult, error) {
	params.Pagination = true

	// 1. 查询用户列表（排除密码）
	result, err := u.UserDAL.Query(ctx, params, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "created_at", Direction: util.DESC}, // 按创建时间倒序
			},
			OmitFields: []string{"password"}, // 安全：不返回密码
		},
	})
	if err != nil {
		return nil, err
	}

	// 2. 批量查询用户角色（避免 N+1 查询）
	if userIDs := result.Data.ToIDs(); len(userIDs) > 0 {
		userRoleResult, err := u.UserRoleDAL.Query(ctx, schema.UserRoleQueryParam{
			InUserIDs: userIDs, // 批量查询
		}, schema.UserRoleQueryOptions{
			JoinRole: true, // 关联角色名称
		})
		if err != nil {
			return nil, err
		}

		// 3. 构建用户ID -> 角色列表的映射
		userRolesMap := userRoleResult.Data.ToUserIDMap()
		for _, user := range result.Data {
			user.Roles = userRolesMap[user.ID]
		}
	}

	return result, nil
}

// Get 获取单个用户信息（含角色）。
func (u *User) Get(ctx context.Context, id string) (*schema.User, error) {
	// 1. 查询用户（排除密码）
	user, err := u.UserDAL.Get(ctx, id, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			OmitFields: []string{"password"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.NotFound("", "User not found")
	}

	// 2. 查询用户角色
	userRoleResult, err := u.UserRoleDAL.Query(ctx, schema.UserRoleQueryParam{
		UserID: id,
	})
	if err != nil {
		return nil, err
	}
	user.Roles = userRoleResult.Data

	return user, nil
}

// Create 创建新用户（含角色分配）。
// 流程：
//  1. 用户名唯一性校验
//  2. 设置默认密码（如果未提供）
//  3. 事务内：创建用户 + 创建用户角色
func (u *User) Create(ctx context.Context, userForm *schema.UserForm) (*schema.User, error) {
	// 1. 用户名唯一性校验
	existsUsername, err := u.UserDAL.ExistsUsername(ctx, userForm.Username)
	if err != nil {
		return nil, err
	} else if existsUsername {
		return nil, errors.BadRequest("", "Username already exists")
	}

	// 2. 初始化用户实体
	user := &schema.User{
		ID:        util.NewXID(), // 生成唯一 ID
		CreatedAt: time.Now(),
	}

	// 3. 设置默认密码（如果未提供）
	if userForm.Password == "" {
		userForm.Password = config.C.General.DefaultLoginPwd
	}

	// 4. 填充表单数据
	if err := userForm.FillTo(user); err != nil {
		return nil, err
	}

	// 5. 事务内执行（用户 + 角色）
	err = u.Trans.Exec(ctx, func(ctx context.Context) error {
		// 5.1 创建用户
		if err := u.UserDAL.Create(ctx, user); err != nil {
			return err
		}

		// 5.2 创建用户角色
		for _, userRole := range userForm.Roles {
			userRole.ID = util.NewXID()
			userRole.UserID = user.ID
			userRole.CreatedAt = time.Now()
			if err := u.UserRoleDAL.Create(ctx, userRole); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 6. 返回用户（含角色）
	user.Roles = userForm.Roles
	return user, nil
}

// Update 更新用户信息（含角色重分配）。
// 流程：
//  1. 用户存在性校验
//  2. 用户名唯一性校验（如果修改）
//  3. 事务内：更新用户 + 删除旧角色 + 创建新角色 + 清除缓存
func (u *User) Update(ctx context.Context, id string, userForm *schema.UserForm) error {
	// 1. 获取用户信息
	user, err := u.UserDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if user == nil {
		return errors.NotFound("", "User not found")
	}

	// 2. 用户名唯一性校验（仅当修改时）
	if user.Username != userForm.Username {
		existsUsername, err := u.UserDAL.ExistsUsername(ctx, userForm.Username)
		if err != nil {
			return err
		} else if existsUsername {
			return errors.BadRequest("", "Username already exists")
		}
	}

	// 3. 填充表单数据
	if err := userForm.FillTo(user); err != nil {
		return err
	}
	user.UpdatedAt = time.Now()

	// 4. 事务内执行
	return u.Trans.Exec(ctx, func(ctx context.Context) error {
		// 4.1 更新用户
		if err := u.UserDAL.Update(ctx, user); err != nil {
			return err
		}

		// 4.2 删除旧角色
		if err := u.UserRoleDAL.DeleteByUserID(ctx, id); err != nil {
			return err
		}

		// 4.3 创建新角色
		for _, userRole := range userForm.Roles {
			if userRole.ID == "" {
				userRole.ID = util.NewXID()
			}
			userRole.UserID = user.ID
			if userRole.CreatedAt.IsZero() {
				userRole.CreatedAt = time.Now()
			}
			userRole.UpdatedAt = time.Now()
			if err := u.UserRoleDAL.Create(ctx, userRole); err != nil {
				return err
			}
		}

		// 4.4 清除用户缓存（确保权限立即生效）
		return u.Cache.Delete(ctx, config.CacheNSForUser, id)
	})
}

// Delete 删除用户（级联删除角色）。
// 流程：
//  1. 用户存在性校验
//  2. 事务内：删除用户 + 删除角色 + 清除缓存
func (u *User) Delete(ctx context.Context, id string) error {
	// 1. 用户存在性校验
	exists, err := u.UserDAL.ExistsID(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "User not found")
	}

	// 2. 事务内执行
	return u.Trans.Exec(ctx, func(ctx context.Context) error {
		// 2.1 删除用户
		if err := u.UserDAL.Delete(ctx, id); err != nil {
			return err
		}
		// 2.2 删除用户角色
		if err := u.UserRoleDAL.DeleteByUserID(ctx, id); err != nil {
			return err
		}
		// 2.3 清除用户缓存
		return u.Cache.Delete(ctx, config.CacheNSForUser, id)
	})
}

// ResetPassword 重置用户密码为默认值。
func (u *User) ResetPassword(ctx context.Context, id string) error {
	// 1. 用户存在性校验
	exists, err := u.UserDAL.ExistsID(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "User not found")
	}

	// 2. 生成默认密码哈希
	hashPass, err := hash.GeneratePassword(config.C.General.DefaultLoginPwd)
	if err != nil {
		return errors.BadRequest("", "Failed to generate hash password: %s", err.Error())
	}

	// 3. 事务内更新密码
	return u.Trans.Exec(ctx, func(ctx context.Context) error {
		return u.UserDAL.UpdatePasswordByID(ctx, id, hashPass)
	})
}

// GetRoleIDs 获取用户的角色 ID 列表（用于缓存和权限校验）。
func (u *User) GetRoleIDs(ctx context.Context, id string) ([]string, error) {
	userRoleResult, err := u.UserRoleDAL.Query(ctx, schema.UserRoleQueryParam{
		UserID: id,
	}, schema.UserRoleQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"role_id"},
		},
	})
	if err != nil {
		return nil, err
	}
	return userRoleResult.Data.ToRoleIDs(), nil
}
