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

// GetRoleMenuDB 根据上下文返回角色菜单关联表的 GORM DB 实例。
// 功能：
//   - 自动注入事务（如果存在）
//   - 自动添加行锁（如果需要）
//   - 指定模型为 schema.RoleMenu
func GetRoleMenuDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.RoleMenu))
}

// RoleMenu 是角色菜单关联实体的数据访问对象（DAO）。
// 表示角色与菜单的多对多关系（一个角色可访问多个菜单，一个菜单可被多个角色访问）。
type RoleMenu struct {
	DB *gorm.DB // 基础数据库连接
}

// Query 根据参数和选项查询角色菜单关联列表。
// 支持：
//   - 角色过滤（RoleID）
//   - 分页
//   - 字段选择/排序
func (rm *RoleMenu) Query(ctx context.Context, params schema.RoleMenuQueryParam, opts ...schema.RoleMenuQueryOptions) (*schema.RoleMenuQueryResult, error) {
	// 1. 解析查询选项
	var opt schema.RoleMenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 2. 构建基础查询
	db := GetRoleMenuDB(ctx, rm.DB)

	// 3. 应用查询条件
	if v := params.RoleID; len(v) > 0 {
		db = db.Where("role_id = ?", v) // 查询指定角色的菜单权限
	}

	// 4. 执行分页查询
	var list schema.RoleMenus
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err) // 保留错误堆栈
	}

	// 5. 返回查询结果
	return &schema.RoleMenuQueryResult{
		PageResult: pageResult,
		Data:       list,
	}, nil
}

// Get 根据 ID 获取单个角色菜单关联。
func (rm *RoleMenu) Get(ctx context.Context, id string, opts ...schema.RoleMenuQueryOptions) (*schema.RoleMenu, error) {
	var opt schema.RoleMenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.RoleMenu)
	ok, err := util.FindOne(ctx, GetRoleMenuDB(ctx, rm.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exists 检查角色菜单关联 ID 是否存在。
func (rm *RoleMenu) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetRoleMenuDB(ctx, rm.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// Create 创建新的角色菜单关联。
// 用于为角色分配菜单权限。
func (rm *RoleMenu) Create(ctx context.Context, item *schema.RoleMenu) error {
	result := GetRoleMenuDB(ctx, rm.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update 更新角色菜单关联信息。
// 更新所有字段（排除 created_at），确保数据一致性。
func (rm *RoleMenu) Update(ctx context.Context, item *schema.RoleMenu) error {
	result := GetRoleMenuDB(ctx, rm.DB).
		Where("id=?", item.ID).
		Select("*").        // 更新所有字段
		Omit("created_at"). // 但排除 created_at
		Updates(item)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除角色菜单关联。
func (rm *RoleMenu) Delete(ctx context.Context, id string) error {
	result := GetRoleMenuDB(ctx, rm.DB).Where("id=?", id).Delete(new(schema.RoleMenu))
	return errors.WithStack(result.Error)
}

// DeleteByRoleID 根据角色 ID 删除所有菜单关联。
// 用于角色删除或重置菜单权限场景。
func (rm *RoleMenu) DeleteByRoleID(ctx context.Context, roleID string) error {
	result := GetRoleMenuDB(ctx, rm.DB).Where("role_id=?", roleID).Delete(new(schema.RoleMenu))
	return errors.WithStack(result.Error)
}

// DeleteByMenuID 根据菜单 ID 删除所有角色关联。
// 用于菜单删除或重置角色权限场景。
func (rm *RoleMenu) DeleteByMenuID(ctx context.Context, menuID string) error {
	result := GetRoleMenuDB(ctx, rm.DB).Where("menu_id=?", menuID).Delete(new(schema.RoleMenu))
	return errors.WithStack(result.Error)
}
