// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package bootstrap

import (
	"context"
	"fmt"

	"net/http"
	"os"
	"strings"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/utility/prom"
	"github.com/zhian9/blogo-server/internal/wirex"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/mail"
	"github.com/zhian9/blogo-server/pkg/ossx"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
)

type RunConfig struct {
	WorkDir   string
	Configs   string
	StaticDir string
}

func Run(ctx context.Context, runCfg RunConfig) error {
	defer func() {
		if err := zap.L().Sync(); err != nil {
			// Windows 上 stdout/stderr 无法 sync，忽略
			if !strings.Contains(err.Error(), "handle is invalid") &&
				!strings.Contains(err.Error(), "bad file descriptor") {
				fmt.Printf("Failed to sync zap logger: %s\n", err.Error())
			}
		}
	}()

	//加载配置
	workDir := runCfg.WorkDir
	staticDir := runCfg.StaticDir

	config.MustLoad(workDir, strings.Split(runCfg.Configs, ",")...)

	config.C.General.WorkDir = workDir
	config.C.Middleware.Static.Dir = staticDir

	// 打印最终配置
	config.C.Print()

	// 预加载
	config.C.PreLoad()

	// 初始化 R2 对象存储（若未配置则自动跳过）
	ossx.Init(ossx.Config{
		AccountID:       config.C.Storage.R2.AccountID,
		AccessKeyID:     config.C.Storage.R2.AccessKeyID,
		SecretAccessKey: config.C.Storage.R2.SecretAccessKey,
		Bucket:          config.C.Storage.R2.Bucket,
		PublicDomain:    config.C.Storage.R2.PublicDomain,
		Endpoint:        config.C.Storage.R2.Endpoint,
		UploadDir:       config.C.Storage.R2.UploadDir,
	})

	// 初始化日志
	cleanLoggerFn, err := logging.InitWithConfig(ctx, &config.C.Logger, initLoggerHook)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	//
	// 设置全局日志标签（用于分类）
	ctx = logging.NewTag(ctx, logging.TagKeyMain)

	// 记录服务启动日志（带上下文：version/pid/workdir 等）
	logging.Context(ctx).Info("starting service ...",
		zap.String("version", config.C.General.Version),
		zap.Int("pid", os.Getpid()),
		zap.String("workdir", workDir),
		zap.String("config", runCfg.Configs),
		zap.String("static", staticDir),
	)

	// 3. 启动 pprof（性能分析）
	if addr := config.C.General.PprofAddr; addr != "" {
		logging.Context(ctx).Info("pprof server is listening on " + addr)
		go func() {
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				logging.Context(ctx).Error("failed to listen pprof server", zap.Error(err))
			}
		}()
	}

	//  4. 构建依赖注入容器（Wire）
	injector, cleanInjectorFn, err := wirex.NewInjector(ctx)
	if err != nil {
		return fmt.Errorf("failed to build injector: %w", err)
	}

	// 初始化中间件（如 Casbin 策略加载）
	if err := injector.M.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize middleware: %w", err)
	}

	// 5. 初始化全局邮件发送器
	initMailSender()

	// 6. 初始化 Prometheus 指标
	prom.Init()

	// 6. 启动 HTTP 服务 + 优雅关闭

	return util.Run(ctx, func(ctx context.Context) (func(), error) {
		// 启动 HTTP 服务（Gin 路由、中间件等）
		cleanHTTPServerFn, err := startHTTPServer(ctx, injector)
		if err != nil {
			return cleanInjectorFn, fmt.Errorf("failed to start HTTP server: %w", err)
		}

		// 返回资源清理函数（按依赖顺序逆向关闭）
		return func() {
			// 释放中间件资源（如 Casbin）
			if err := injector.M.Release(ctx); err != nil {
				logging.Context(ctx).Error("failed to release injector", zap.Error(err))
			}

			// 关闭 HTTP 服务
			if cleanHTTPServerFn != nil {
				cleanHTTPServerFn()
			}

			// 清理依赖注入容器
			if cleanInjectorFn != nil {
				cleanInjectorFn()
			}

			// 关闭日志系统（flush 缓冲区）
			if cleanLoggerFn != nil {
				cleanLoggerFn()
			}
		}, nil
	})
}

func initMailSender() {
	cfg := config.C.Email
	if cfg.Host == "" || cfg.FromEmail == "" || cfg.Password == "" {
		return
	}
	mail.SetSender(&mail.SmtpSender{
		SmtpHost: cfg.Host,
		Port:     cfg.Port,
		FromName: cfg.SenderName,
		FromMail: cfg.FromEmail,
		UserName: cfg.FromEmail,
		AuthCode: cfg.Password,
	})
}
