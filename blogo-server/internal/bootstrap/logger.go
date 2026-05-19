// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package bootstrap

import (
	"context"

	"github.com/spf13/cast"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/pkg/gormx"
	"github.com/zhian9/blogo-server/pkg/logging"
)

func initLoggerHook(_ context.Context, cfg *logging.HookConfig) (*logging.Hook, error) {
	extra := cfg.Extra
	if extra == nil {
		extra = make(map[string]string)
	}
	extra["appname"] = config.C.General.AppName

	// 兼容大小写/不同风格的配置键名
	optionString := func(m map[string]string, keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k]; ok && v != "" {
				return v
			}
		}
		return ""
	}
	optionInt := func(m map[string]string, keys ...string) int {
		return cast.ToInt(optionString(m, keys...))
	}
	optionBool := func(m map[string]string, keys ...string) bool {
		return cast.ToBool(optionString(m, keys...))
	}

	switch cfg.Type {
	case "gorm":
		db, err := gormx.New(gormx.Config{
			Debug:       optionBool(cfg.Options, "Debug", "debug"),
			DBType:      optionString(cfg.Options, "DBType", "db_type", "type"),
			DSN:         optionString(cfg.Options, "DSN", "dsn"),
			MaxLifeTime: optionInt(cfg.Options, "MaxLifeTime", "max_life_time"),
			MaxIdleTime: optionInt(cfg.Options, "MaxIdleTime", "max_idle_time"),
			MaxOpenConn: optionInt(cfg.Options, "MaxOpenConn", "max_open_conn"),
			MaxIdleConn: optionInt(cfg.Options, "MaxIdleConn", "max_idle_conn"),
			TablePrefix: config.C.Storage.DB.TablePrefix,
		})
		if err != nil {
			return nil, err
		}

		// 当未在配置中指定 max_buffer/max_thread 时，使用 Hook 的内置默认值
		opts := []logging.HookOptions{logging.SetHookExtra(cfg.Extra)}
		if cfg.MaxBuffer > 0 {
			opts = append(opts, logging.SetHookMaxJobs(cfg.MaxBuffer))
		}
		if cfg.MaxThread > 0 {
			opts = append(opts, logging.SetHookMaxWorkers(cfg.MaxThread))
		}
		hook := logging.NewHook(logging.NewGormHook(db), opts...)
		return hook, nil
	default:
		return nil, nil
	}
}
