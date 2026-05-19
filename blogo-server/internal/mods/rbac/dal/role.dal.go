// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// GetRoleDB 根据上下文返回角色表的 GORM DB 实例。
// 功能：
//   - 自动注入事务
//   - 自动添加行锁
//   - 指定模型为 schema.Role
func GetRoleDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Role))
}

// Role 是角色实体的数据访问对象（DAO）。
type Role struct {
	DB *gorm.DB // 基础数据库连接
}

// Query 根据参数和选项查询角色列表。
// 支持：
//   - ID 列表过滤（InIDs）
//   - 名称模糊查询（LikeName）
//   - 状态过滤（Status）
//   - 更新时间范围（GtUpdatedAt）
//   - 分页
//   - 字段选择/排序
func (r *Role) Query(ctx context.Context, params schema.RoleQueryParam, opts ...schema.RoleQueryOptions) (*schema.RoleQueryResult, error) {
	// 1. 解析查询选项
	var opt schema.RoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 2. 构建基础查询
	db := GetRoleDB(ctx, r.DB)

	// 3. 应用查询条件
	if v := params.InIDs; len(v) > 0 {
		db = db.Where("id IN (?)", v) // 精确匹配 ID 列表
	}
	if v := params.LikeName; len(v) > 0 {
		db = db.Where("name LIKE ?", "%"+v+"%") // 角色名称模糊查询
	}
	if v := params.Status; len(v) > 0 {
		db = db.Where("status = ?", v) // 状态精确匹配（enabled/disabled）
	}
	if v := params.GtUpdatedAt; v != nil {
		// 更新时间范围查询（用于增量同步）
		db = db.Where("updated_at > ?", v)
	}

	// 4. 执行分页查询
	var list schema.Roles
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err) // 保留错误堆栈
	}

	// 5. 返回查询结果
	return &schema.RoleQueryResult{
		PageResult: pageResult,
		Data:       list,
	}, nil
}

// Get 根据 ID 获取单个角色。
// 支持字段选择（通过 opts）。
func (r *Role) Get(ctx context.Context, id string, opts ...schema.RoleQueryOptions) (*schema.Role, error) {
	var opt schema.RoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	role := new(schema.Role)
	ok, err := util.FindOne(ctx, GetRoleDB(ctx, r.DB).Where("id=?", id), opt.QueryOptions, role)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil // 未找到返回 nil, nil
	}
	return role, nil
}

// Exists 检查角色 ID 是否存在。
// 用于删除/更新前的校验。
func (r *Role) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetRoleDB(ctx, r.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// ExistsCode 检查角色编码（Code）是否已存在。
// 用于创建/修改角色时的唯一性校验（Code 通常全局唯一）。
func (r *Role) ExistsCode(ctx context.Context, code string) (bool, error) {
	ok, err := util.Exists(ctx, GetRoleDB(ctx, r.DB).Where("code=?", code))
	return ok, errors.WithStack(err)
}

// Create 创建新角色。
func (r *Role) Create(ctx context.Context, role *schema.Role) error {
	result := GetRoleDB(ctx, r.DB).Create(role)
	return errors.WithStack(result.Error)
}

// Update 更新角色信息。
// 更新所有字段（排除 created_at），确保数据一致性。
func (r *Role) Update(ctx context.Context, role *schema.Role) error {
	result := GetRoleDB(ctx, r.DB).
		Where("id=?", role.ID).
		Select("*").        // 更新所有字段
		Omit("created_at"). // 但排除 created_at（避免意外修改）
		Updates(role)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除角色。
// 注意：实际项目中需先解除角色与用户/菜单的关联！
func (r *Role) Delete(ctx context.Context, id string) error {
	result := GetRoleDB(ctx, r.DB).Where("id=?", id).Delete(new(schema.Role))
	return errors.WithStack(result.Error)
}
