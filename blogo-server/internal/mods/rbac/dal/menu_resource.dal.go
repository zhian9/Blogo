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

// GetMenuResourceDB 根据上下文返回菜单资源表的 GORM DB 实例。
// 功能：
//   - 自动注入事务（如果存在）
//   - 自动添加行锁（如果需要）
//   - 指定模型为 schema.MenuResource
func GetMenuResourceDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.MenuResource))
}

// MenuResource 是菜单资源实体的数据访问对象（DAO）。
// 表示菜单关联的具体 API 资源（HTTP 方法 + 路径）。
// 示例：
//   - Method: "GET", Path: "/api/v1/users"
//   - Method: "POST", Path: "/api/v1/users"
type MenuResource struct {
	DB *gorm.DB // 基础数据库连接
}

// Query 根据参数和选项查询菜单资源列表。
// 支持：
//   - 单菜单过滤（MenuID）
//   - 多菜单批量过滤（MenuIDs）
//   - 分页
//   - 字段选择/排序
func (mr *MenuResource) Query(ctx context.Context, params schema.MenuResourceQueryParam, opts ...schema.MenuResourceQueryOptions) (*schema.MenuResourceQueryResult, error) {
	// 1. 解析查询选项
	var opt schema.MenuResourceQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 2. 构建基础查询
	db := GetMenuResourceDB(ctx, mr.DB)
	// 3. 应用查询条件
	if v := params.MenuID; len(v) > 0 {
		db = db.Where("menu_id = ?", v) // 查询单个菜单的资源
	}
	if v := params.MenuIDs; len(v) > 0 {
		db = db.Where("menu_id IN ?", v) // 批量查询多个菜单的资源
	}

	// 4. 执行分页查询
	var list schema.MenuResources
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err) // 保留错误堆栈
	}

	// 5. 返回查询结果
	return &schema.MenuResourceQueryResult{
		PageResult: pageResult,
		Data:       list,
	}, nil
}

// Get 根据 ID 获取单个菜单资源。
func (mr *MenuResource) Get(ctx context.Context, id string, opts ...schema.MenuResourceQueryOptions) (*schema.MenuResource, error) {
	var opt schema.MenuResourceQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.MenuResource)
	ok, err := util.FindOne(ctx, GetMenuResourceDB(ctx, mr.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exists 检查菜单资源 ID 是否存在。
func (mr *MenuResource) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuResourceDB(ctx, mr.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// ExistsMethodPathByMenuID 检查菜单下是否存在相同的 (Method, Path) 资源。
// 用于创建/更新资源时的唯一性校验（防止重复定义）。
func (mr *MenuResource) ExistsMethodPathByMenuID(ctx context.Context, method, path, menuID string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuResourceDB(ctx, mr.DB).Where("method=? AND path=? AND menu_id=?", method, path, menuID))
	return ok, errors.WithStack(err)
}

// Create 创建新的菜单资源。
// 用于为菜单定义可访问的 API 端点。
func (mr *MenuResource) Create(ctx context.Context, item *schema.MenuResource) error {
	result := GetMenuResourceDB(ctx, mr.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update 更新菜单资源信息。
// 更新所有字段（排除 created_at），确保数据一致性。
func (mr *MenuResource) Update(ctx context.Context, item *schema.MenuResource) error {
	result := GetMenuResourceDB(ctx, mr.DB).
		Where("id=?", item.ID).
		Select("*").        // 更新所有字段
		Omit("created_at"). // 但排除 created_at
		Updates(item)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除菜单资源。
func (mr *MenuResource) Delete(ctx context.Context, id string) error {
	result := GetMenuResourceDB(ctx, mr.DB).Where("id=?", id).Delete(new(schema.MenuResource))
	return errors.WithStack(result.Error)
}

// DeleteByMenuID 根据菜单 ID 删除所有关联资源。
// 用于菜单删除或重置资源场景。
func (mr *MenuResource) DeleteByMenuID(ctx context.Context, menuID string) error {
	result := GetMenuResourceDB(ctx, mr.DB).Where("menu_id=?", menuID).Delete(new(schema.MenuResource))
	return errors.WithStack(result.Error)
}
