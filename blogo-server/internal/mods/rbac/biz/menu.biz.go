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
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/rbac/dal"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/cachex"
	"github.com/zhian9/blogo-server/pkg/encoding/json"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Menu 是菜单管理业务的核心对象，聚合了缓存、事务、数据访问等依赖。
type Menu struct {
	Cache           cachex.Cacher     // 缓存客户端（用于 Casbin 策略同步）
	Trans           *util.Trans       // 事务管理器
	MenuDAL         *dal.Menu         // 菜单数据访问层
	MenuResourceDAL *dal.MenuResource // 菜单资源数据访问层
	RoleMenuDAL     *dal.RoleMenu     // 角色菜单数据访问层
}

// InitFromFile 从 JSON/YAML 文件初始化菜单数据。
// 用于系统首次启动时加载默认菜单。
func (m *Menu) InitFromFile(ctx context.Context, menuFile string) error {
	// 1. 读取文件
	f, err := os.ReadFile(menuFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// 文件不存在 → 跳过（非错误）
			logging.Context(ctx).Warn("Menu data file not found, skip init menu data from file", zap.String("file", menuFile))
			return nil
		}
		return err
	}

	// 2. 解析文件（支持 JSON/YAML）
	var menus schema.Menus
	if ext := filepath.Ext(menuFile); ext == ".json" {
		if err := json.Unmarshal(f, &menus); err != nil {
			return errors.Wrapf(err, "Unmarshal JSON file '%s' failed", menuFile)
		}
	} else if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(f, &menus); err != nil {
			return errors.Wrapf(err, "Unmarshal YAML file '%s' failed", menuFile)
		}
	} else {
		return errors.Errorf("Unsupported file type '%s'", ext)
	}

	// 3. 事务内批量创建菜单（递归处理树形结构）
	return m.Trans.Exec(ctx, func(ctx context.Context) error {
		return m.createInBatchByParent(ctx, menus, nil)
	})
}

// createInBatchByParent 递归创建菜单树（按父级分批处理）。
// parent: 父菜单（nil 表示根菜单）
func (m *Menu) createInBatchByParent(ctx context.Context, items schema.Menus, parent *schema.Menu) error {
	total := len(items)

	for i, item := range items {
		// 1. 确定父菜单 ID
		var parentID string
		if parent != nil {
			parentID = parent.ID
		}

		// 2. 尝试查找现有菜单（按 ID/Code/Name + ParentID）
		var (
			menuItem *schema.Menu
			err      error
		)
		if item.ID != "" {
			menuItem, err = m.MenuDAL.Get(ctx, item.ID)
		} else if item.Code != "" {
			menuItem, err = m.MenuDAL.GetByCodeAndParentID(ctx, item.Code, parentID)
		} else if item.Name != "" {
			menuItem, err = m.MenuDAL.GetByNameAndParentID(ctx, item.Name, parentID)
		}
		if err != nil {
			return err
		}

		// 3. 设置默认状态
		if item.Status == "" {
			item.Status = schema.MenuStatusEnabled
		}

		// 4. 更新现有菜单（仅当字段变更时）
		if menuItem != nil {
			changed := false
			if menuItem.Name != item.Name {
				menuItem.Name = item.Name
				changed = true
			}
			if menuItem.Description != item.Description {
				menuItem.Description = item.Description
				changed = true
			}
			if menuItem.Path != item.Path {
				menuItem.Path = item.Path
				changed = true
			}
			if menuItem.Type != item.Type {
				menuItem.Type = item.Type
				changed = true
			}
			if menuItem.Sequence != item.Sequence {
				menuItem.Sequence = item.Sequence
				changed = true
			}
			if menuItem.Status != item.Status {
				menuItem.Status = item.Status
				changed = true
			}
			if changed {
				menuItem.UpdatedAt = time.Now()
				if err := m.MenuDAL.Update(ctx, menuItem); err != nil {
					return err
				}
			}
		} else {
			// 5. 创建新菜单
			if item.ID == "" {
				item.ID = util.NewXID()
			}
			if item.Sequence == 0 {
				item.Sequence = total - i // 保持文件顺序
			}
			item.ParentID = parentID
			if parent != nil {
				// 构建 ParentPath（用于快速查询祖先）
				item.ParentPath = parent.ParentPath + parent.ID + util.TreePathDelimiter
			}
			menuItem = item
			if err := m.MenuDAL.Create(ctx, item); err != nil {
				return err
			}
		}

		// 6. 处理菜单资源（API 权限）
		for _, res := range item.Resources {
			// 6.1 跳过已存在的资源（按 ID）
			if res.ID != "" {
				exists, err := m.MenuResourceDAL.Exists(ctx, res.ID)
				if err != nil {
					return err
				} else if exists {
					continue
				}
			}
			// 6.2 跳过重复资源（按 Method+Path+MenuID）
			if res.Path != "" {
				exists, err := m.MenuResourceDAL.ExistsMethodPathByMenuID(ctx, res.Method, res.Path, menuItem.ID)
				if err != nil {
					return err
				} else if exists {
					continue
				}
			}
			// 6.3 创建新资源
			if res.ID == "" {
				res.ID = util.NewXID()
			}
			res.MenuID = menuItem.ID
			if err := m.MenuResourceDAL.Create(ctx, res); err != nil {
				return err
			}
		}

		// 7. 递归处理子菜单
		if item.Children != nil {
			if err := m.createInBatchByParent(ctx, *item.Children, menuItem); err != nil {
				return err
			}
		}
	}
	return nil
}

// Query 查询菜单列表（支持树形结构和资源加载）。
// 流程：
//  1. 处理 CodePath 查询（转换为 ParentPathPrefix）
//  2. 查询菜单
//  3. 补全子菜单（模糊查询时）
//  4. 加载资源（如果需要）
//  5. 构建树形结构
func (m *Menu) Query(ctx context.Context, params schema.MenuQueryParam) (*schema.MenuQueryResult, error) {
	params.Pagination = false // 菜单查询通常不分页

	// 1. 处理 CodePath 查询（如 "system.user"）
	if err := m.fillQueryParam(ctx, &params); err != nil {
		return nil, err
	}

	// 2. 查询菜单
	result, err := m.MenuDAL.Query(ctx, params, schema.MenuQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: schema.MenusOrderParams, // 按序号和创建时间排序
		},
	})
	if err != nil {
		return nil, err
	}

	// 3. 模糊查询时补全子菜单（确保树形完整）
	if params.LikeName != "" || params.CodePath != "" {
		result.Data, err = m.appendChildren(ctx, result.Data)
		if err != nil {
			return nil, err
		}
	}

	// 4. 加载菜单资源（API 权限）
	if params.IncludeResources {
		for i, item := range result.Data {
			resResult, err := m.MenuResourceDAL.Query(ctx, schema.MenuResourceQueryParam{
				MenuID: item.ID,
			})
			if err != nil {
				return nil, err
			}
			result.Data[i].Resources = resResult.Data
		}
	}

	// 5. 构建树形结构
	result.Data = result.Data.ToTree()
	return result, nil
}

// fillQueryParam 处理 CodePath 查询参数。
// 将 "system.user" 转换为 ParentPathPrefix="system.user."
func (m *Menu) fillQueryParam(ctx context.Context, params *schema.MenuQueryParam) error {
	if params.CodePath != "" {
		var (
			codes    []string
			lastMenu schema.Menu
		)
		for _, code := range strings.Split(params.CodePath, util.TreePathDelimiter) {
			if code == "" {
				continue
			}
			codes = append(codes, code)
			// 逐级查找菜单
			menu, err := m.MenuDAL.GetByCodeAndParentID(ctx, code, lastMenu.ParentID, schema.MenuQueryOptions{
				QueryOptions: util.QueryOptions{
					SelectFields: []string{"id", "parent_id", "parent_path"},
				},
			})
			if err != nil {
				return err
			} else if menu == nil {
				return errors.NotFound("", "Menu not found by code '%s'", strings.Join(codes, util.TreePathDelimiter))
			}
			lastMenu = *menu
		}
		// 设置 ParentPathPrefix 用于查询后代菜单
		params.ParentPathPrefix = lastMenu.ParentPath + lastMenu.ID + util.TreePathDelimiter
	}
	return nil
}

// appendChildren 补全菜单的子菜单和祖先菜单（用于模糊查询）。
func (m *Menu) appendChildren(ctx context.Context, data schema.Menus) (schema.Menus, error) {
	if len(data) == 0 {
		return data, nil
	}

	// 辅助函数：检查 ID 是否已在 data 中
	existsInData := func(id string) bool {
		for _, item := range data {
			if item.ID == id {
				return true
			}
		}
		return false
	}

	// 1. 补全子菜单
	for _, item := range data {
		childResult, err := m.MenuDAL.Query(ctx, schema.MenuQueryParam{
			ParentPathPrefix: item.ParentPath + item.ID + util.TreePathDelimiter,
		})
		if err != nil {
			return nil, err
		}
		for _, child := range childResult.Data {
			if existsInData(child.ID) {
				continue
			}
			data = append(data, child)
		}
	}

	// 2. 补全祖先菜单
	if parentIDs := data.SplitParentIDs(); len(parentIDs) > 0 {
		parentResult, err := m.MenuDAL.Query(ctx, schema.MenuQueryParam{
			InIDs: parentIDs,
		})
		if err != nil {
			return nil, err
		}
		for _, p := range parentResult.Data {
			if existsInData(p.ID) {
				continue
			}
			data = append(data, p)
		}
	}
	sort.Sort(data) // 重新排序

	return data, nil
}

// Get 获取单个菜单信息（含资源）。
func (m *Menu) Get(ctx context.Context, id string) (*schema.Menu, error) {
	// 1. 查询菜单
	menu, err := m.MenuDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if menu == nil {
		return nil, errors.NotFound("", "Menu not found")
	}

	// 2. 查询菜单资源
	menuResResult, err := m.MenuResourceDAL.Query(ctx, schema.MenuResourceQueryParam{
		MenuID: menu.ID,
	})
	if err != nil {
		return nil, err
	}
	menu.Resources = menuResResult.Data

	return menu, nil
}

// Create 创建新菜单（含资源）。
// 流程：
//  1. 检查菜单操作权限（DenyOperateMenu）
//  2. 验证父菜单存在性
//  3. 同级菜单编码唯一性校验
//  4. 事务内：创建菜单 + 创建资源
func (m *Menu) Create(ctx context.Context, menuForm *schema.MenuForm) (*schema.Menu, error) {
	if config.C.General.DenyOperateMenu {
		return nil, errors.BadRequest("", "Menu creation is not allowed")
	}

	// 1. 初始化菜单实体
	menu := &schema.Menu{
		ID:        util.NewXID(),
		CreatedAt: time.Now(),
	}

	// 2. 处理父菜单
	if parentID := menuForm.ParentID; parentID != "" {
		parent, err := m.MenuDAL.Get(ctx, parentID)
		if err != nil {
			return nil, err
		} else if parent == nil {
			return nil, errors.NotFound("", "Parent not found")
		}
		menu.ParentPath = parent.ParentPath + parent.ID + util.TreePathDelimiter
	}

	// 3. 同级菜单编码唯一性校验
	if exists, err := m.MenuDAL.ExistsCodeByParentID(ctx, menuForm.Code, menuForm.ParentID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Menu code already exists at the same level")
	}

	// 4. 填充表单数据
	if err := menuForm.FillTo(menu); err != nil {
		return nil, err
	}

	// 5. 事务内执行
	err := m.Trans.Exec(ctx, func(ctx context.Context) error {
		// 5.1 创建菜单
		if err := m.MenuDAL.Create(ctx, menu); err != nil {
			return err
		}

		// 5.2 创建菜单资源
		for _, res := range menuForm.Resources {
			res.ID = util.NewXID()
			res.MenuID = menu.ID
			res.CreatedAt = time.Now()
			if err := m.MenuResourceDAL.Create(ctx, res); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return menu, nil
}

// Update 更新菜单信息（含资源重分配和树形结构调整）。
// 流程：
//  1. 检查菜单操作权限
//  2. 处理父菜单变更（更新 ParentPath 和子菜单）
//  3. 同级菜单编码唯一性校验
//  4. 事务内：更新菜单 + 更新子菜单 ParentPath + 更新状态 + 重分配资源 + 同步 Casbin
func (m *Menu) Update(ctx context.Context, id string, menuForm *schema.MenuForm) error {
	if config.C.General.DenyOperateMenu {
		return errors.BadRequest("", "Menu update is not allowed")
	}

	// 1. 获取菜单信息
	menu, err := m.MenuDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if menu == nil {
		return errors.NotFound("", "Menu not found")
	}

	// 2. 处理父菜单变更
	oldParentPath := menu.ParentPath
	oldStatus := menu.Status
	var childData schema.Menus
	if menu.ParentID != menuForm.ParentID {
		if parentID := menuForm.ParentID; parentID != "" {
			parent, err := m.MenuDAL.Get(ctx, parentID)
			if err != nil {
				return err
			} else if parent == nil {
				return errors.NotFound("", "Parent not found")
			}
			menu.ParentPath = parent.ParentPath + parent.ID + util.TreePathDelimiter
		} else {
			menu.ParentPath = ""
		}

		// 获取所有子菜单（用于更新 ParentPath）
		childResult, err := m.MenuDAL.Query(ctx, schema.MenuQueryParam{
			ParentPathPrefix: oldParentPath + menu.ID + util.TreePathDelimiter,
		}, schema.MenuQueryOptions{
			QueryOptions: util.QueryOptions{
				SelectFields: []string{"id", "parent_path"},
			},
		})
		if err != nil {
			return err
		}
		childData = childResult.Data
	}

	// 3. 同级菜单编码唯一性校验
	if menu.Code != menuForm.Code {
		if exists, err := m.MenuDAL.ExistsCodeByParentID(ctx, menuForm.Code, menuForm.ParentID); err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "Menu code already exists at the same level")
		}
	}

	// 4. 填充表单数据
	if err := menuForm.FillTo(menu); err != nil {
		return err
	}

	// 5. 事务内执行
	return m.Trans.Exec(ctx, func(ctx context.Context) error {
		// 5.1 级联更新状态（如果状态变更）
		if oldStatus != menuForm.Status {
			oldPath := oldParentPath + menu.ID + util.TreePathDelimiter
			if err := m.MenuDAL.UpdateStatusByParentPath(ctx, oldPath, menuForm.Status); err != nil {
				return err
			}
		}

		// 5.2 更新子菜单 ParentPath（如果父菜单变更）
		for _, child := range childData {
			oldPath := oldParentPath + menu.ID + util.TreePathDelimiter
			newPath := menu.ParentPath + menu.ID + util.TreePathDelimiter
			err := m.MenuDAL.UpdateParentPath(ctx, child.ID, strings.Replace(child.ParentPath, oldPath, newPath, 1))
			if err != nil {
				return err
			}
		}

		// 5.3 更新菜单
		if err := m.MenuDAL.Update(ctx, menu); err != nil {
			return err
		}

		// 5.4 重分配菜单资源
		if err := m.MenuResourceDAL.DeleteByMenuID(ctx, id); err != nil {
			return err
		}
		for _, res := range menuForm.Resources {
			if res.ID == "" {
				res.ID = util.NewXID()
			}
			res.MenuID = id
			if res.CreatedAt.IsZero() {
				res.CreatedAt = time.Now()
			}
			res.UpdatedAt = time.Now()
			if err := m.MenuResourceDAL.Create(ctx, res); err != nil {
				return err
			}
		}

		// 5.5 触发 Casbin 策略同步
		return m.syncToCasbin(ctx)
	})
}

// Delete 删除菜单（级联删除子菜单、资源、角色关联）。
// 流程：
//  1. 检查菜单操作权限
//  2. 获取所有子菜单
//  3. 事务内：删除菜单 + 删除子菜单 + 同步 Casbin
func (m *Menu) Delete(ctx context.Context, id string) error {
	if config.C.General.DenyOperateMenu {
		return errors.BadRequest("", "Menu deletion is not allowed")
	}

	// 1. 获取菜单信息
	menu, err := m.MenuDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if menu == nil {
		return errors.NotFound("", "Menu not found")
	}

	// 2. 获取所有子菜单
	childResult, err := m.MenuDAL.Query(ctx, schema.MenuQueryParam{
		ParentPathPrefix: menu.ParentPath + menu.ID + util.TreePathDelimiter,
	}, schema.MenuQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"id"},
		},
	})
	if err != nil {
		return err
	}

	// 3. 事务内执行
	return m.Trans.Exec(ctx, func(ctx context.Context) error {
		// 3.1 删除当前菜单
		if err := m.delete(ctx, id); err != nil {
			return err
		}

		// 3.2 删除所有子菜单
		for _, child := range childResult.Data {
			if err := m.delete(ctx, child.ID); err != nil {
				return err
			}
		}

		// 3.3 触发 Casbin 策略同步
		return m.syncToCasbin(ctx)
	})
}

// delete 删除单个菜单（内部方法）。
func (m *Menu) delete(ctx context.Context, id string) error {
	if err := m.MenuDAL.Delete(ctx, id); err != nil {
		return err
	}
	if err := m.MenuResourceDAL.DeleteByMenuID(ctx, id); err != nil {
		return err
	}
	if err := m.RoleMenuDAL.DeleteByMenuID(ctx, id); err != nil {
		return err
	}
	return nil
}

// syncToCasbin 触发 Casbin 策略重载。
// 通过缓存设置一个时间戳信号，由 Casbinx 自动监听并重载策略。
func (m *Menu) syncToCasbin(ctx context.Context) error {
	return m.Cache.Set(ctx, config.CacheNSForRole, config.CacheKeyForSyncToCasbin, fmt.Sprintf("%d", time.Now().Unix()))
}
