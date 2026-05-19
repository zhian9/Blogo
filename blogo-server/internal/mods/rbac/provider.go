// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package rbac

import (
	"github.com/zhian9/blogo-server/internal/mods/rbac/dal"
	"github.com/zhian9/blogo-server/pkg/cachex"
)

// NewCasbinx 构造 Casbinx，仅注入需要的导出依赖字段。
// 非导出字段（如 enforcer、ticker）由运行期逻辑自行初始化。
func NewCasbinx(cache cachex.Cacher, menu *dal.Menu, mr *dal.MenuResource, role *dal.Role) *Casbinx {
	return &Casbinx{
		Cache:           cache,
		MenuDAL:         menu,
		MenuResourceDAL: mr,
		RoleDAL:         role,
	}
}
