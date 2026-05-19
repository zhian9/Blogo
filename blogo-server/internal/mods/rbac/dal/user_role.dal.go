// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"
	"fmt"

	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// GetUserRoleDB 根据上下文返回用户角色关联表的 GORM DB 实例。
// 功能：
//   - 自动注入事务（如果存在）
//   - 自动添加行锁（如果需要）
//   - 指定模型为 schema.UserRole
func GetUserRoleDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.UserRole))
}

// UserRole 是用户角色关联实体的数据访问对象（DAO）。
// 表示用户与角色的多对多关系。
type UserRole struct {
	DB *gorm.DB // 基础数据库连接
}

// Query 根据参数和选项查询用户角色关联列表。
// 支持：
//   - 用户 ID 列表过滤（InUserIDs）
//   - 单用户过滤（UserID）
//   - 角色过滤（RoleID）
//   - 关联角色信息（JoinRole）
//   - 分页
//   - 字段选择/排序
func (ur *UserRole) Query(ctx context.Context, params schema.UserRoleQueryParam, opts ...schema.UserRoleQueryOptions) (*schema.UserRoleQueryResult, error) {
	// 1. 解析查询选项
	var opt schema.UserRoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 2. 构建基础查询（使用别名 ur 避免字段冲突）
	db := ur.DB.Table(fmt.Sprintf("%s AS ur", new(schema.UserRole).TableName()))

	// 3. 关联角色表（可选）
	//    - 用于返回角色名称（role_name）
	if opt.JoinRole {
		db = db.Joins(fmt.Sprintf("LEFT JOIN %s b ON ur.role_id = b.id", new(schema.Role).TableName()))
		db = db.Select("ur.*, b.name AS role_name, b.code AS role_code")
	}

	// 4. 应用查询条件
	if v := params.InUserIDs; len(v) > 0 {
		db = db.Where("ur.user_id IN (?)", v) // 批量查询用户角色
	}
	if v := params.UserID; len(v) > 0 {
		db = db.Where("ur.user_id = ?", v) // 查询指定用户的角色
	}
	if v := params.RoleID; len(v) > 0 {
		db = db.Where("ur.role_id = ?", v) // 查询指定角色的用户
	}

	// 5. 执行分页查询
	var list schema.UserRoles
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err) // 保留错误堆栈
	}

	// 6. 返回查询结果
	return &schema.UserRoleQueryResult{
		PageResult: pageResult,
		Data:       list,
	}, nil
}

// Get 根据 ID 获取单个用户角色关联。
func (ur *UserRole) Get(ctx context.Context, id string, opts ...schema.UserRoleQueryOptions) (*schema.UserRole, error) {
	var opt schema.UserRoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.UserRole)
	ok, err := util.FindOne(ctx, GetUserRoleDB(ctx, ur.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exists 检查用户角色关联 ID 是否存在。
func (ur *UserRole) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetUserRoleDB(ctx, ur.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// Create 创建新的用户角色关联。
func (ur *UserRole) Create(ctx context.Context, item *schema.UserRole) error {
	result := GetUserRoleDB(ctx, ur.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update 更新用户角色关联信息。
// 更新所有字段（排除 created_at），确保数据一致性。
func (ur *UserRole) Update(ctx context.Context, item *schema.UserRole) error {
	result := GetUserRoleDB(ctx, ur.DB).
		Where("id=?", item.ID).
		Select("*").        // 更新所有字段
		Omit("created_at"). // 但排除 created_at
		Updates(item)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除用户角色关联。
func (ur *UserRole) Delete(ctx context.Context, id string) error {
	result := GetUserRoleDB(ctx, ur.DB).Where("id=?", id).Delete(new(schema.UserRole))
	return errors.WithStack(result.Error)
}

// DeleteByUserID 根据用户 ID 删除所有角色关联。
// 用于用户删除或重置角色场景。
func (ur *UserRole) DeleteByUserID(ctx context.Context, userID string) error {
	result := GetUserRoleDB(ctx, ur.DB).Where("user_id=?", userID).Delete(new(schema.UserRole))
	return errors.WithStack(result.Error)
}

// DeleteByRoleID 根据角色 ID 删除所有用户关联。
// 用于角色删除或重置用户场景。
func (ur *UserRole) DeleteByRoleID(ctx context.Context, roleID string) error {
	result := GetUserRoleDB(ctx, ur.DB).Where("role_id=?", roleID).Delete(new(schema.UserRole))
	return errors.WithStack(result.Error)
}
