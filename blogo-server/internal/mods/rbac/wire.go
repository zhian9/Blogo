// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package rbac

import (
	"github.com/google/wire"
	"github.com/zhian9/blogo-server/internal/mods/rbac/api"
	"github.com/zhian9/blogo-server/internal/mods/rbac/biz"
	"github.com/zhian9/blogo-server/internal/mods/rbac/dal"
)

var Set = wire.NewSet(
	wire.Struct(new(RBAC), "*"),
	NewCasbinx,
	wire.Struct(new(dal.Menu), "*"),
	wire.Struct(new(biz.Menu), "*"),
	wire.Struct(new(api.Menu), "*"),
	wire.Struct(new(dal.MenuResource), "*"),
	wire.Struct(new(dal.Role), "*"),
	wire.Struct(new(biz.Role), "*"),
	wire.Struct(new(api.Role), "*"),
	wire.Struct(new(dal.RoleMenu), "*"),
	wire.Struct(new(dal.User), "*"),
	wire.Struct(new(biz.User), "*"),
	wire.Struct(new(api.User), "*"),
	wire.Struct(new(dal.UserRole), "*"),
	wire.Struct(new(biz.Login), "*"),
	wire.Struct(new(api.Login), "*"),
	wire.Struct(new(api.Logger), "*"),
	wire.Struct(new(biz.Logger), "*"),
	wire.Struct(new(dal.Logger), "*"),
	wire.Struct(new(api.OperationLog), "*"),
	wire.Struct(new(biz.OperationLog), "*"),
	wire.Struct(new(dal.OperationLog), "*"),
	wire.Struct(new(dal.UserFollow), "*"),
	wire.Struct(new(biz.UserFollow), "*"),
	wire.Struct(new(api.UserFollow), "*"),
)
