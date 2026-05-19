// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zhian9/blogo-server/pkg/logging"
	"go.uber.org/zap"
)

// RunConfig 配置 Run 行为（可扩展）
type RunConfig struct {
	ShutdownTimeout time.Duration // 优雅关闭超时（默认 30 秒）
	Reloadable      bool          // 是否支持 SIGHUP 重载
}

// DefaultRunConfig 默认配置
var DefaultRunConfig = RunConfig{
	ShutdownTimeout: 30 * time.Second,
	Reloadable:      false,
}

// Run 启动应用并监听系统信号，支持优雅关闭。
func Run(ctx context.Context, handler func(ctx context.Context) (func(), error), opts ...RunConfig) error {
	// 应用配置
	cfg := DefaultRunConfig
	if len(opts) > 0 {
		cfg = opts[0]
	}

	// 创建可取消的上下文（用于主动退出）
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 信号通道
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// 启动服务
	cleanFn, err := handler(ctx)
	if err != nil {
		return err
	}

	// 等待信号或上下文取消
	select {
	case sig := <-sc:
		logging.Context(ctx).Info("Received signal", zap.String("signal", sig.String()))
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			// 正常退出
		case syscall.SIGHUP:
			if cfg.Reloadable {
				logging.Context(ctx).Info("SIGHUP received, reloading config...")
				// TODO: 实现配置重载
				// config.Reload()
				// return Run(ctx, handler, cfg) // 递归重启
			} else {
				logging.Context(ctx).Info("SIGHUP received but reload is disabled")
			}
		default:
			logging.Context(ctx).Warn("Received unhandled signal", zap.String("signal", sig.String()))
		}
	case <-ctx.Done():
		logging.Context(ctx).Info("Context canceled, shutting down...")
	}

	// 优雅关闭
	logging.Context(ctx).Info("Shutting down server...")

	// 创建带超时的关闭上下文
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	// 在 goroutine 中执行清理（避免阻塞主 goroutine）
	done := make(chan struct{})
	go func() {
		defer close(done)
		if cleanFn != nil {
			// 清理函数错误应记录但不中断退出
			if err := func() error {
				cleanFn()
				return nil
			}(); err != nil {
				logging.Context(ctx).Error("Failed to clean up resources", zap.Error(err))
			}
		}
	}()

	// 等待清理完成或超时
	select {
	case <-done:
		logging.Context(ctx).Info("Server shutdown complete")
	case <-shutdownCtx.Done():
		logging.Context(ctx).Error("Server shutdown timeout", zap.Duration("timeout", cfg.ShutdownTimeout))
	}

	// 确保日志刷入磁盘（替代 time.Sleep）
	_ = zap.L().Sync()

	// 正常退出
	os.Exit(0)
	return nil // 实际不会执行（os.Exit 已终止）
}
