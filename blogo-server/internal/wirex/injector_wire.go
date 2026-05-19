// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

//go:build wireinject
// +build wireinject

package wirex

import (
	"context"

	"github.com/google/wire"
	"github.com/zhian9/blogo-server/internal/mods"
)

// NewInjector 使用 Google Wire 生成依赖注入实现。
// 它会组装：DB、Cache、Auth 以及业务模块 Mods（包含 RBAC 与 Blog）。
func NewInjector(ctx context.Context) (*Injector, func(), error) { // wire 会生成实现到 wire_gen.go
	wire.Build(
		InitDB,
		InitCacher,
		InitAuth,
		ProvideTrans,
		ProvideCaptchaRedisClient,
		ProvideCaptchaService,
		mods.Set,                        // 包含 Blog.Set 与 RBAC.Set
		wire.Struct(new(Injector), "*"), // 组装 Injector{DB, Cache, Auth, M}
	)
	return nil, nil, nil
}
