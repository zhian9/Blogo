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

// GetMenuDB 根据上下文返回菜单表的 GORM DB 实例。
// 功能：
//   - 自动注入事务（如果存在）
//   - 自动添加行锁（如果需要）
//   - 指定模型为 schema.Menu
func GetMenuDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Menu))
}

// Menu 是菜单实体的数据访问对象（DAO）。
type Menu struct {
	DB *gorm.DB // 基础数据库连接
}

// Query 根据参数和选项查询菜单列表。
// 支持：
//   - ID 列表过滤（InIDs）
//   - 名称模糊查询（LikeName）
//   - 状态过滤（Status）
//   - 父菜单过滤（ParentID/ParentPathPrefix）
//   - 用户/角色权限过滤（通过关联表）
//   - 分页
//   - 字段选择/排序
func (m *Menu) Query(ctx context.Context, params schema.MenuQueryParam, opts ...schema.MenuQueryOptions) (*schema.MenuQueryResult, error) {
	// 1. 解析查询选项
	var opt schema.MenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 2. 构建基础查询
	db := GetMenuDB(ctx, m.DB)

	// 3. 应用查询条件
	if v := params.InIDs; len(v) > 0 {
		db = db.Where("id IN ?", v) // 精确匹配 ID 列表
	}
	if v := params.LikeName; len(v) > 0 {
		db = db.Where("name LIKE ?", "%"+v+"%") // 菜单名称模糊查询
	}
	if v := params.Status; len(v) > 0 {
		db = db.Where("status = ?", v) // 状态精确匹配（enabled/disabled）
	}
	if v := params.ParentID; len(v) > 0 {
		db = db.Where("parent_id = ?", v) // 查询直接子菜单
	}
	if v := params.ParentPathPrefix; len(v) > 0 {
		// 查询后代菜单（利用 parent_path 字段）
		db = db.Where("parent_path LIKE ?", v+"%")
	}

	// 4. RBAC 权限过滤（核心！）
	if v := params.UserID; len(v) > 0 {
		// 通过用户 → 角色 → 菜单 链路过滤
		// m. 查询用户关联的角色 ID
		userRoleQuery := GetUserRoleDB(ctx, m.DB).Where("user_id = ?", v).Select("role_id")
		// b. 查询角色关联的菜单 ID
		roleMenuQuery := GetRoleMenuDB(ctx, m.DB).Where("role_id IN (?)", userRoleQuery).Select("menu_id")
		// c. 仅返回用户有权访问的菜单
		db = db.Where("id IN (?)", roleMenuQuery)
	}
	if v := params.RoleID; len(v) > 0 {
		// 直接通过角色 ID 过滤菜单
		roleMenuQuery := GetRoleMenuDB(ctx, m.DB).Where("role_id = ?", v).Select("menu_id")
		db = db.Where("id IN (?)", roleMenuQuery)
	}

	// 5. 执行分页查询
	var list schema.Menus
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err) // 保留错误堆栈
	}

	// 6. 返回查询结果
	return &schema.MenuQueryResult{
		PageResult: pageResult,
		Data:       list,
	}, nil
}

// Get 根据 ID 获取单个菜单。
func (m *Menu) Get(ctx context.Context, id string, opts ...schema.MenuQueryOptions) (*schema.Menu, error) {
	var opt schema.MenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	menu := new(schema.Menu)
	ok, err := util.FindOne(ctx, GetMenuDB(ctx, m.DB).Where("id=?", id), opt.QueryOptions, menu)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return menu, nil
}

// GetByCodeAndParentID 根据编码和父ID获取菜单（同级唯一性校验）。
func (m *Menu) GetByCodeAndParentID(ctx context.Context, code, parentID string, opts ...schema.MenuQueryOptions) (*schema.Menu, error) {
	var opt schema.MenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	menu := new(schema.Menu)
	ok, err := util.FindOne(ctx, GetMenuDB(ctx, m.DB).Where("code=? AND parent_id=?", code, parentID), opt.QueryOptions, menu)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return menu, nil
}

// GetByNameAndParentID 根据名称和父ID获取菜单（同级唯一性校验）。
func (m *Menu) GetByNameAndParentID(ctx context.Context, name, parentID string, opts ...schema.MenuQueryOptions) (*schema.Menu, error) {
	var opt schema.MenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	menu := new(schema.Menu)
	ok, err := util.FindOne(ctx, GetMenuDB(ctx, m.DB).Where("name=? AND parent_id=?", name, parentID), opt.QueryOptions, menu)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return menu, nil
}

// Exists 检查菜单 ID 是否存在。
func (m *Menu) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuDB(ctx, m.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// ExistsCodeByParentID 检查同级菜单编码是否唯一。
// 用于创建/修改菜单时的唯一性校验。
func (m *Menu) ExistsCodeByParentID(ctx context.Context, code, parentID string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuDB(ctx, m.DB).Where("code=? AND parent_id=?", code, parentID))
	return ok, errors.WithStack(err)
}

// ExistsNameByParentID 检查同级菜单名称是否唯一。
func (m *Menu) ExistsNameByParentID(ctx context.Context, name, parentID string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuDB(ctx, m.DB).Where("name=? AND parent_id=?", name, parentID))
	return ok, errors.WithStack(err)
}

// Create 创建新菜单。
func (m *Menu) Create(ctx context.Context, menu *schema.Menu) error {
	result := GetMenuDB(ctx, m.DB).Create(menu)
	return errors.WithStack(result.Error)
}

// Update 更新菜单信息。
// 更新所有字段（排除 created_at），确保数据一致性。
func (m *Menu) Update(ctx context.Context, menu *schema.Menu) error {
	result := GetMenuDB(ctx, m.DB).
		Where("id=?", menu.ID).
		Select("*").        // 更新所有字段
		Omit("created_at"). // 但排除 created_at
		Updates(menu)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除菜单。
func (m *Menu) Delete(ctx context.Context, id string) error {
	result := GetMenuDB(ctx, m.DB).Where("id=?", id).Delete(new(schema.Menu))
	return errors.WithStack(result.Error)
}

// UpdateParentPath 更新菜单的父路径（用于快速查询祖先）。
// 格式：parent1.parent2.parent3
func (m *Menu) UpdateParentPath(ctx context.Context, id, parentPath string) error {
	result := GetMenuDB(ctx, m.DB).Where("id=?", id).Update("parent_path", parentPath)
	return errors.WithStack(result.Error)
}

// UpdateStatusByParentPath 批量更新后代菜单状态。
// 用于级联启用/禁用菜单（如禁用父菜单时禁用所有子菜单）。
func (m *Menu) UpdateStatusByParentPath(ctx context.Context, parentPath, status string) error {
	result := GetMenuDB(ctx, m.DB).Where("parent_path LIKE ?", parentPath+"%").Update("status", status)
	return errors.WithStack(result.Error)
}
