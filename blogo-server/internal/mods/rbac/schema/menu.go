// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package schema

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// 菜单状态定义
const (
	MenuStatusDisabled = "disabled" // 禁用
	MenuStatusEnabled  = "enabled"  // 启用
)

// MenusOrderParams 定义菜单的默认排序规则。
// 优先级：sequence（降序） > created_at（降序）
var (
	MenusOrderParams = []util.OrderByParam{
		{Field: "sequence", Direction: util.DESC},   // 按序号降序排列
		{Field: "created_at", Direction: util.DESC}, // 按创建时间降序排列
	}
)

// Menu 表示 RBAC 系统中的菜单项。
// 字段说明：
//   - ID: 唯一标识（20 字符）
//   - Code: 菜单编码（同级唯一，用于权限标识）
//   - Name: 显示名称
//   - Description: 描述信息
//   - Sequence: 排序序号（越大越靠前）
//   - Type: 菜单类型（page=页面, button=按钮）
//   - Path: 访问路径（前端路由或 API 路径）
//   - Properties: 扩展属性（JSON 格式，如图标、是否隐藏等）
//   - Status: 状态（enabled/disabled）
//   - ParentID: 父菜单 ID（根菜单为空）
//   - ParentPath: 父菜单路径（如 "root.parent"，用于快速查询祖先）
//   - Children: 子菜单列表（虚拟字段，不存入数据库）
//   - Resources: 关联的资源列表（虚拟字段，不存入数据库）
type Menu struct {
	// 唯一 ID（主键）
	ID string `json:"id" gorm:"size:20;primaryKey;"`

	// 菜单编码（同级唯一，用于权限标识）
	Code string `json:"code" gorm:"size:32;index;"`

	// 显示名称
	Name string `json:"name" gorm:"size:128;index"`

	// 描述信息
	Description string `json:"description" gorm:"size:1024"`

	// 排序序号（越大越靠前）
	Sequence int `json:"sequence" gorm:"index;"`

	// 菜单类型（page/button）
	Type string `json:"type" gorm:"size:20;index"`

	// 访问路径（前端路由或 API 路径）
	Path string `json:"path" gorm:"size:255;"`

	// 扩展属性（JSON 格式，如 {"icon": "user", "hidden": true}）
	Properties string `json:"properties" gorm:"type:text;"`

	// 状态（enabled/disabled）
	Status string `json:"status" gorm:"size:20;index"`

	// 父菜单 ID（根菜单为空）
	ParentID string `json:"parent_id" gorm:"size:20;index;"`

	// 父菜单路径（如 "root.parent"，用于快速查询祖先）
	ParentPath string `json:"parent_path" gorm:"size:255;index;"`

	// ========== 虚拟字段（不存入数据库） ==========

	// 子菜单列表（用于前端树形展示）
	Children *Menus `json:"children" gorm:"-"`

	// 创建时间
	CreatedAt time.Time `json:"created_at" gorm:"index;"`

	// 更新时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`

	// 关联的资源列表（API 路径+方法，用于 Casbin 权限）
	Resources MenuResources `json:"resources" gorm:"-"`
}

// TableName 动态返回表名（支持表前缀）。
func (a *Menu) TableName() string {
	return config.C.FormatTableName("menu")
}

// MenuQueryParam 定义菜单查询的参数结构。
type MenuQueryParam struct {
	util.PaginationParam // 分页参数

	// CodePath: 菜单编码路径（如 "system.user"）
	CodePath string `form:"code"`

	// LikeName: 菜单名称（模糊匹配）
	LikeName string `form:"name"`

	// IncludeResources: 是否包含资源列表
	IncludeResources bool `form:"includeResources"`

	// InIDs: 指定菜单 ID 列表（用于精确查询）
	InIDs []string `form:"-"`

	// Status: 菜单状态（disabled/enabled）
	Status string `form:"-"`

	// ParentID: 父菜单 ID（用于查询子菜单）
	ParentID string `form:"-"`

	// ParentPathPrefix: 父路径前缀（用于查询后代菜单）
	ParentPathPrefix string `form:"-"`

	// UserID: 用户 ID（用于查询用户可见菜单）
	UserID string `form:"-"`

	// RoleID: 角色 ID（用于查询角色关联菜单）
	RoleID string `form:"-"`
}

// MenuQueryOptions 定义菜单查询的选项（字段选择、排序等）。
type MenuQueryOptions struct {
	util.QueryOptions
}

// MenuQueryResult 定义菜单查询的返回结果。
type MenuQueryResult struct {
	// 菜单数据列表
	Data Menus

	// 分页信息
	PageResult *util.PaginationResult
}

// Menus 是 Menu 指针的切片类型。
type Menus []*Menu

// Len 返回切片长度（实现 sort.Interface）
func (m Menus) Len() int {
	return len(m)
}

// Less 定义排序规则（实现 sort.Interface）
// 优先级：Sequence（降序） > CreatedAt（降序）
func (m Menus) Less(i, j int) bool {
	if m[i].Sequence == m[j].Sequence {
		return m[i].CreatedAt.Unix() > m[j].CreatedAt.Unix()
	}
	return m[i].Sequence > m[j].Sequence
}

// Swap 交换元素（实现 sort.Interface）
func (m Menus) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// ToMap 将菜单切片转换为 ID 映射（便于快速查找）
func (m Menus) ToMap() map[string]*Menu {
	ms := make(map[string]*Menu)
	for _, menu := range m {
		ms[menu.ID] = menu
	}
	return ms
}

// SplitParentIDs 提取所有祖先菜单 ID（去重）
// 用于查询父级菜单（如构建完整菜单树）
func (m Menus) SplitParentIDs() []string {
	parentIDs := make([]string, 0, len(m))
	idMapper := make(map[string]struct{})
	for _, menu := range m {
		if _, ok := idMapper[menu.ID]; ok {
			continue
		}
		idMapper[menu.ID] = struct{}{}
		if pp := menu.ParentPath; pp != "" {
			for _, pid := range strings.Split(pp, util.TreePathDelimiter) {
				if pid == "" {
					continue
				}
				if _, ok := idMapper[pid]; ok {
					continue
				}
				parentIDs = append(parentIDs, pid)
				idMapper[pid] = struct{}{}
			}
		}
	}
	return parentIDs
}

// ToTree 将扁平菜单列表转换为树形结构
// 用于前端展示（递归构建 children）
func (m Menus) ToTree() Menus {
	var list Menus
	mm := m.ToMap()
	for _, menu := range m {
		if menu.ParentID == "" {
			// 根菜单
			list = append(list, menu)
			continue
		}
		if parent, ok := mm[menu.ParentID]; ok {
			// 添加到父菜单的 children
			if parent.Children == nil {
				children := Menus{menu}
				parent.Children = &children
				continue
			}
			*parent.Children = append(*parent.Children, menu)
		}
	}
	return list
}

// MenuForm 表示菜单创建/更新的请求结构。
type MenuForm struct {
	// 菜单编码（必填，同级唯一）
	Code string `json:"code" binding:"required,max=32"`

	// 显示名称（必填）
	Name string `json:"name" binding:"required,max=128"`

	// 描述信息
	Description string `json:"description"`

	// 排序序号
	Sequence int `json:"sequence"`

	// 菜单类型（必填，page/button）
	Type string `json:"type" binding:"required,oneof=page button"`

	// 访问路径
	Path string `json:"path"`

	// 扩展属性（JSON 格式）
	Properties string `json:"properties"`

	// 状态（必填，enabled/disabled）
	Status string `json:"status" binding:"required,oneof=disabled enabled"`

	// 父菜单 ID
	ParentID string `json:"parent_id"`

	// 关联的资源列表
	Resources MenuResources `json:"resources"`
}

// Validate 验证 MenuForm 的合法性。
// 主要验证 Properties 是否为有效 JSON。
func (mf *MenuForm) Validate() error {
	if v := mf.Properties; v != "" {
		if !json.Valid([]byte(v)) {
			return errors.BadRequest("", "invalid properties")
		}
	}
	return nil
}

// FillTo 将 MenuForm 的数据填充到 Menu 实体。
// 用于创建/更新菜单时的数据转换。
func (mf *MenuForm) FillTo(menu *Menu) error {
	menu.Code = mf.Code
	menu.Name = mf.Name
	menu.Description = mf.Description
	menu.Sequence = mf.Sequence
	menu.Type = mf.Type
	menu.Path = mf.Path
	menu.Properties = mf.Properties
	menu.Status = mf.Status
	menu.ParentID = mf.ParentID
	return nil
}
