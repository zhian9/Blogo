// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package rbac

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// seedMenuItem 种子菜单节点
type seedMenuItem struct {
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Path      string         `json:"path"`
	Sequence  int            `json:"sequence"`
	Icon      string         `json:"icon,omitempty"`
	Children  []seedMenuItem `json:"children,omitempty"`
	Resources []seedMenuRes  `json:"resources,omitempty"`
}

type seedMenuRes struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// systemMenuTree 系统核心菜单树（硬编码）
func systemMenuTree() []seedMenuItem {
	return []seedMenuItem{
		{
			Code: "dashboard", Name: "控制中心", Type: "directory", Path: "/dashboard", Sequence: 1,
			Icon: "DashboardOutlined",
		},
		{
			Code: "content", Name: "内容管理", Type: "directory", Path: "", Sequence: 2,
			Icon: "AppstoreOutlined",
			Children: []seedMenuItem{
				{Code: "articles", Name: "文章管理", Type: "menu", Path: "/articles", Sequence: 1, Icon: "FileTextOutlined"},
				{Code: "categories", Name: "分类管理", Type: "menu", Path: "/categories", Sequence: 2, Icon: "AppstoreOutlined"},
				{Code: "tags", Name: "标签管理", Type: "menu", Path: "/tags", Sequence: 3, Icon: "TagOutlined"},
				{Code: "comments", Name: "评论管理", Type: "menu", Path: "/comments", Sequence: 4, Icon: "MessageOutlined"},
			},
		},
		{
			Code: "system", Name: "系统管理", Type: "directory", Path: "", Sequence: 3,
			Icon: "SettingOutlined",
			Children: []seedMenuItem{
				{Code: "users", Name: "用户管理", Type: "menu", Path: "/users", Sequence: 1, Icon: "UserOutlined"},
				{Code: "roles", Name: "角色管理", Type: "menu", Path: "/roles", Sequence: 2, Icon: "TeamOutlined"},
				{Code: "menus", Name: "菜单管理", Type: "menu", Path: "/menus", Sequence: 3, Icon: "MenuOutlined"},
				{Code: "settings", Name: "系统设置", Type: "menu", Path: "/settings", Sequence: 4, Icon: "ToolOutlined"},
			},
		},
		{
			Code: "audit", Name: "安全审计", Type: "directory", Path: "", Sequence: 4,
			Icon: "SafetyCertificateOutlined",
			Children: []seedMenuItem{
				{Code: "operation-logs", Name: "操作日志", Type: "menu", Path: "/logs/audit", Sequence: 1, Icon: "HistoryOutlined"},
			},
		},
	}
}

// InitMenus 硬编码初始化系统菜单树（每次启动全量覆盖，确保与代码一致）
func (r *RBAC) InitMenus(ctx context.Context) error {
	// 清除旧菜单数据（menu.json 时代遗留）
	r.DB.Exec("DELETE FROM menu_resource")
	r.DB.Exec("DELETE FROM role_menu")
	r.DB.Exec("DELETE FROM menu")

	tree := systemMenuTree()
	return r.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range tree {
			if err := createMenuItem(tx, nil, item); err != nil {
				return err
			}
		}
		return nil
	})
}

func createMenuItem(tx *gorm.DB, parentID *string, item seedMenuItem) error {
	// 查找或创建菜单节点
	var existing schema.Menu
	err := tx.Where("code = ? AND parent_id <=> ?", item.Code, parentID).First(&existing).Error
	if err == nil {
		// 已存在 → 更新关键字段
		updates := map[string]interface{}{
			"name": item.Name, "type": item.Type, "path": item.Path,
			"sequence": item.Sequence, "status": schema.MenuStatusEnabled,
		}
		if item.Icon != "" {
			props, _ := json.Marshal(map[string]string{"icon": item.Icon})
			updates["properties"] = string(props)
		}
		return tx.Model(&existing).Updates(updates).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// 新建
	menuID := util.NewXID()
	props := ""
	if item.Icon != "" {
		b, _ := json.Marshal(map[string]string{"icon": item.Icon})
		props = string(b)
	}

	var parentPath string
	if parentID != nil {
		var parent schema.Menu
		if tx.Where("id = ?", *parentID).First(&parent).Error == nil {
			if parent.ParentPath != "" {
				parentPath = parent.ParentPath + util.TreePathDelimiter + *parentID
			} else {
				parentPath = *parentID
			}
		}
	}

	menu := schema.Menu{
		ID:         menuID,
		Code:       item.Code,
		Name:       item.Name,
		Type:       item.Type,
		Path:       item.Path,
		Sequence:   item.Sequence,
		Properties: props,
		Status:     schema.MenuStatusEnabled,
		ParentPath: parentPath,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if parentID != nil {
		menu.ParentID = *parentID
	}
	if err := tx.Create(&menu).Error; err != nil {
		return err
	}

	// 递归创建子节点
	for _, child := range item.Children {
		if err := createMenuItem(tx, &menuID, child); err != nil {
			return err
		}
	}

	return nil
}

// InitAdminMenuAccess 确保 super_admin 角色拥有全部菜单
func (r *RBAC) InitAdminMenuAccess(ctx context.Context) error {
	var superAdminRole schema.Role
	if err := r.DB.Where("code = ?", "super_admin").First(&superAdminRole).Error; err != nil {
		var oldAdmin schema.Role
		if err := r.DB.Where("code = ?", "admin").First(&oldAdmin).Error; err != nil {
			return nil
		}
		superAdminRole = oldAdmin
	}

	// 获取所有菜单
	var allMenus []schema.Menu
	if err := r.DB.Find(&allMenus).Error; err != nil {
		return err
	}

	// 确保 super_admin 拥有每个菜单
	for _, menu := range allMenus {
		var count int64
		r.DB.Model(&schema.RoleMenu{}).
			Where("role_id = ? AND menu_id = ?", superAdminRole.ID, menu.ID).
			Count(&count)
		if count == 0 {
			rm := schema.RoleMenu{
				ID:        util.NewXID(),
				RoleID:    superAdminRole.ID,
				MenuID:    menu.ID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := r.DB.Create(&rm).Error; err != nil {
				logging.Context(ctx).Error("failed to grant menu to super_admin", zap.Error(err))
			}
		}
	}

	return nil
}
