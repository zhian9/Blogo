// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/rbac/dal"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/cachex"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// Role 是角色管理业务的核心对象，聚合了缓存、事务、数据访问等依赖。
type Role struct {
	Cache       cachex.Cacher // 缓存客户端（用于 Casbin 策略同步）
	Trans       *util.Trans   // 事务管理器
	RoleDAL     *dal.Role     // 角色数据访问层
	RoleMenuDAL *dal.RoleMenu // 角色菜单数据访问层
	UserRoleDAL *dal.UserRole // 用户角色数据访问层
}

// Query 查询角色列表（支持分页和下拉选择模式）。
// 流程：
//  1. 根据 ResultType 决定是否分页
//  2. 设置查询字段（下拉模式仅需 id/name）
//  3. 按序号和创建时间排序
func (r *Role) Query(ctx context.Context, params schema.RoleQueryParam) (*schema.RoleQueryResult, error) {
	params.Pagination = true

	var selectFields []string
	// 下拉选择模式：不分页，仅返回 id 和 name
	if params.ResultType == schema.RoleResultTypeSelect {
		params.Pagination = false
		selectFields = []string{"id", "name"}
	}

	result, err := r.RoleDAL.Query(ctx, params, schema.RoleQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "sequence", Direction: util.DESC},   // 序号大的排前面
				{Field: "created_at", Direction: util.DESC}, // 创建时间晚的排前面
			},
			SelectFields: selectFields,
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get 获取单个角色信息（含菜单权限）。
func (r *Role) Get(ctx context.Context, id string) (*schema.Role, error) {
	// 1. 查询角色基本信息
	role, err := r.RoleDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if role == nil {
		return nil, errors.NotFound("", "Role not found")
	}

	// 2. 查询角色菜单权限
	roleMenuResult, err := r.RoleMenuDAL.Query(ctx, schema.RoleMenuQueryParam{
		RoleID: id,
	})
	if err != nil {
		return nil, err
	}
	role.Menus = roleMenuResult.Data

	return role, nil
}

// Create 创建新角色（含菜单权限分配）。
// 流程：
//  1. 角色编码唯一性校验
//  2. 事务内：创建角色 + 创建角色菜单 + 同步 Casbin
func (r *Role) Create(ctx context.Context, roleForm *schema.RoleForm) (*schema.Role, error) {
	// 1. 角色编码唯一性校验
	if exists, err := r.RoleDAL.ExistsCode(ctx, roleForm.Code); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Role code already exists")
	}

	// 2. 初始化角色实体
	role := &schema.Role{
		ID:        util.NewXID(), // 生成唯一 ID
		CreatedAt: time.Now(),
	}
	if err := roleForm.FillTo(role); err != nil {
		return nil, err
	}

	// 3. 事务内执行
	err := r.Trans.Exec(ctx, func(ctx context.Context) error {
		// 3.1 创建角色
		if err := r.RoleDAL.Create(ctx, role); err != nil {
			return err
		}

		// 3.2 创建角色菜单
		for _, roleMenu := range roleForm.Menus {
			roleMenu.ID = util.NewXID()
			roleMenu.RoleID = role.ID
			roleMenu.CreatedAt = time.Now()
			if err := r.RoleMenuDAL.Create(ctx, roleMenu); err != nil {
				return err
			}
		}

		// 3.3 触发 Casbin 策略同步
		return r.syncToCasbin(ctx)
	})
	if err != nil {
		return nil, err
	}

	// 4. 返回角色（含菜单）
	role.Menus = roleForm.Menus
	return role, nil
}

// Update 更新角色信息（含菜单权限重分配）。
// 流程：
//  1. 角色存在性校验
//  2. 角色编码唯一性校验（如果修改）
//  3. 事务内：更新角色 + 删除旧菜单 + 创建新菜单 + 同步 Casbin
func (r *Role) Update(ctx context.Context, id string, roleForm *schema.RoleForm) error {
	// 1. 获取角色信息
	role, err := r.RoleDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if role == nil {
		return errors.NotFound("", "Role not found")
	}

	// 2. 角色编码唯一性校验（仅当修改时）
	if role.Code != roleForm.Code {
		if exists, err := r.RoleDAL.ExistsCode(ctx, roleForm.Code); err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "Role code already exists")
		}
	}

	// 3. 填充表单数据
	if err := roleForm.FillTo(role); err != nil {
		return err
	}
	role.UpdatedAt = time.Now()

	// 4. 事务内执行
	return r.Trans.Exec(ctx, func(ctx context.Context) error {
		// 4.1 更新角色
		if err := r.RoleDAL.Update(ctx, role); err != nil {
			return err
		}

		// 4.2 删除旧菜单权限
		if err := r.RoleMenuDAL.DeleteByRoleID(ctx, id); err != nil {
			return err
		}

		// 4.3 创建新菜单权限
		for _, roleMenu := range roleForm.Menus {
			if roleMenu.ID == "" {
				roleMenu.ID = util.NewXID()
			}
			roleMenu.RoleID = role.ID
			if roleMenu.CreatedAt.IsZero() {
				roleMenu.CreatedAt = time.Now()
			}
			roleMenu.UpdatedAt = time.Now()
			if err := r.RoleMenuDAL.Create(ctx, roleMenu); err != nil {
				return err
			}
		}

		// 4.4 触发 Casbin 策略同步
		return r.syncToCasbin(ctx)
	})
}

// Delete 删除角色（级联删除菜单权限和用户关联）。
// 流程：
//  1. 角色存在性校验
//  2. 事务内：删除角色 + 删除菜单权限 + 删除用户关联 + 同步 Casbin
func (r *Role) Delete(ctx context.Context, id string) error {
	// 1. 角色存在性校验
	exists, err := r.RoleDAL.Exists(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Role not found")
	}

	// 2. 事务内执行
	return r.Trans.Exec(ctx, func(ctx context.Context) error {
		// 2.1 删除角色
		if err := r.RoleDAL.Delete(ctx, id); err != nil {
			return err
		}
		// 2.2 删除角色菜单权限
		if err := r.RoleMenuDAL.DeleteByRoleID(ctx, id); err != nil {
			return err
		}
		// 2.3 删除用户角色关联
		if err := r.UserRoleDAL.DeleteByRoleID(ctx, id); err != nil {
			return err
		}
		// 2.4 触发 Casbin 策略同步
		return r.syncToCasbin(ctx)
	})
}

// syncToCasbin 触发 Casbin 策略重载。
// 通过缓存设置一个时间戳信号，由 Casbinx 自动监听并重载策略。
func (r *Role) syncToCasbin(ctx context.Context) error {
	return r.Cache.Set(ctx, config.CacheNSForRole, config.CacheKeyForSyncToCasbin, fmt.Sprintf("%d", time.Now().Unix()))
}
