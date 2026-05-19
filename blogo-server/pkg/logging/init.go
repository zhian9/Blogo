// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package logging

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

// 命名的独立 logger, 供中间件和审计函数直接使用
var (
	AccessLogger   *zap.Logger // HTTP 访问日志 (access.log)
	AuditLogger    *zap.Logger // 后台审计日志 (audit.log)
	SecurityLogger *zap.Logger // 安全事件日志 (security.log)
)

// ── 配置结构 ───────────────────────────────────────

type Config struct {
	Logger LoggerConfig
}

type LoggerConfig struct {
	Debug      bool
	Level      string
	CallerSkip int
	Console    bool `yaml:"console"`    // 是否输出控制台
	File       FileConfig               // 主日志
	AccessLog  FileConfig `yaml:"access_log"`  // HTTP 访问日志
	ErrorLog   FileConfig `yaml:"error_log"`   // 错误日志
	AuditLog   FileConfig `yaml:"audit_log"`   // 审计日志
	SecurityLog FileConfig `yaml:"security_log"` // 安全日志
	Hooks      []*HookConfig
}

type FileConfig struct {
	Enable     bool
	Path       string
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool
}

type HookConfig struct {
	Enable    bool
	Level     string
	Type      string
	MaxBuffer int
	MaxThread int
	Options   map[string]string
	Extra     map[string]string
}

type HookHandlerFunc func(ctx context.Context, hookCfg *HookConfig) (*Hook, error)

func LoadConfigFromYaml(filename string) (*LoggerConfig, error) {
	cfg := &Config{}
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(buf, cfg); err != nil {
		return nil, err
	}
	return &cfg.Logger, nil
}

// InitWithConfig 初始化全局 zap logger。
// 架构:
//   - 主 Logger (zap.L()): Console + app.log + error.log + hooks
//   - AccessLogger:         access.log (HTTP 请求日志)
//   - AuditLogger:          audit/admin.log (后台审计)
//   - SecurityLogger:       audit/security.log (安全事件)
func InitWithConfig(ctx context.Context, cfg *LoggerConfig, hookHandle ...HookHandlerFunc) (func(), error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	var (
		cores    []zapcore.Core
		cleanFns []func()
	)

	// ── 1. 控制台 ────────────────────────────────────
	if cfg.Console || cfg.Debug {
		consoleCfg := zap.NewDevelopmentEncoderConfig()
		consoleCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
		consoleCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		consoleCfg.EncodeCaller = zapcore.ShortCallerEncoder
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleCfg),
			zapcore.AddSync(os.Stdout),
			level,
		))
	}

	// JSON 编码器 (文件统一使用)
	jsonCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	jsonEncoder := zapcore.NewJSONEncoder(jsonCfg)

	// ── 2. 主日志文件 (app.log) ──────────────────────
	if cfg.File.Enable {
		core, clean := newFileCore(jsonEncoder, level, &cfg.File)
		if core != nil {
			cores = append(cores, core)
		}
		if clean != nil {
			cleanFns = append(cleanFns, clean)
		}
	}

	// ── 3. 错误日志 (error.log) ─────────────────────
	if cfg.ErrorLog.Enable {
		core, clean := newFileCore(jsonEncoder, zap.NewAtomicLevelAt(zapcore.ErrorLevel), &cfg.ErrorLog)
		if core != nil {
			cores = append(cores, core)
		}
		if clean != nil {
			cleanFns = append(cleanFns, clean)
		}
	}

	// ── 4. 构建主 Logger ────────────────────────────
	var mainCore zapcore.Core
	switch len(cores) {
	case 0:
		mainCore = zapcore.NewNopCore()
	case 1:
		mainCore = cores[0]
	default:
		mainCore = zapcore.NewTee(cores...)
	}

	skip := cfg.CallerSkip
	if skip <= 0 {
		skip = 2
	}

	logger := zap.New(mainCore,
		zap.WithCaller(true),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCallerSkip(skip),
	)

	// ── 5. 独立日志器 ───────────────────────────────
	if cfg.AccessLog.Enable {
		core, clean := newFileCore(jsonEncoder, level, &cfg.AccessLog)
		if core != nil {
			AccessLogger = zap.New(core, zap.WithCaller(true), zap.AddCallerSkip(skip))
		}
		if clean != nil {
			cleanFns = append(cleanFns, clean)
		}
	}

	if cfg.AuditLog.Enable {
		core, clean := newFileCore(jsonEncoder, level, &cfg.AuditLog)
		if core != nil {
			AuditLogger = zap.New(core, zap.WithCaller(true), zap.AddCallerSkip(skip))
		}
		if clean != nil {
			cleanFns = append(cleanFns, clean)
		}
	}

	if cfg.SecurityLog.Enable {
		core, clean := newFileCore(jsonEncoder, level, &cfg.SecurityLog)
		if core != nil {
			SecurityLogger = zap.New(core, zap.WithCaller(true), zap.AddCallerSkip(skip))
		}
		if clean != nil {
			cleanFns = append(cleanFns, clean)
		}
	}

	// ── 6. 日志钩子 (GORM 等) ────────────────────────
	for _, h := range cfg.Hooks {
		if !h.Enable || len(hookHandle) == 0 {
			continue
		}
		writer, err := hookHandle[0](ctx, h)
		if err != nil {
			return nil, err
		} else if writer == nil {
			continue
		}
		cleanFns = append(cleanFns, func() { writer.Flush() })

		hookLevel := zap.NewAtomicLevelAt(zap.InfoLevel)
		if lv, err := zapcore.ParseLevel(h.Level); err == nil {
			hookLevel.SetLevel(lv)
		}
		hookEncoder := zap.NewProductionEncoderConfig()
		hookEncoder.EncodeTime = zapcore.EpochMillisTimeEncoder

		hookCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(hookEncoder),
			zapcore.AddSync(writer),
			hookLevel,
		)
		logger = logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(c, hookCore)
		}))
	}

	// ── 7. 替换全局 ──────────────────────────────────
	zap.ReplaceGlobals(logger)

	return func() {
		for _, fn := range cleanFns {
			fn()
		}
	}, nil
}

// ── 内部辅助 ────────────────────────────────────────

func newFileCore(enc zapcore.Encoder, level zapcore.LevelEnabler, cfg *FileConfig) (zapcore.Core, func()) {
	_ = os.MkdirAll(filepath.Dir(cfg.Path), 0755)

	w := &lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
		LocalTime:  true,
	}

	return zapcore.NewCore(enc, zapcore.AddSync(w), level), func() { _ = w.Close() }
}
