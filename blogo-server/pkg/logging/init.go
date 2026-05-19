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

// Config 配置结构体
type Config struct {
	Logger LoggerConfig
}

// LoggerConfig 对应配置文件中的Logger
type LoggerConfig struct {
	// 是否开启调式模式
	Debug      bool
	Level      string // 日志级别
	CallerSkip int    //调用栈跳过层数
	File       struct {
		Enable     bool   //是否启用日志文件
		Path       string // 日志文件路径
		MaxSize    int    //单个文件最大大小
		MaxBackups int    //保留的历史文件数量
	}
	Hooks []*HookConfig //日志钩子列表
}

// HookConfig 表示一个日志钩子的配置
type HookConfig struct {
	Enable    bool              //是否启用
	Level     string            //日志级别
	Type      string            //钩子类型。例如:gorm
	MaxBuffer int               //缓冲区最大日志条数
	MaxThread int               //最大并发处理协程数
	Options   map[string]string //钩子专属配置
	Extra     map[string]string //额外扩展字段
}

// HookHandlerFunc 钩子处理器的函数类型
type HookHandlerFunc func(ctx context.Context, hookCfg *HookConfig) (*Hook, error)

// LoadConfigFromYaml 从指定的 yaml 文件加载日志配置
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

// InitWithConfig 根据配置初始化全局 zap 日志器，并支持钩子扩展
func InitWithConfig(ctx context.Context, cfg *LoggerConfig, hookHandle ...HookHandlerFunc) (func(), error) {
	//1.初始化 zap 基础配置
	var zconfig zap.Config
	if cfg.Debug {
		//调试模式
		cfg.Level = "debug"
		zconfig = zap.NewDevelopmentConfig()
	} else {
		// 生成模式
		zconfig = zap.NewProductionConfig()
	}

	//解析日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	// 设置对应的日志级别
	zconfig.Level.SetLevel(level)

	//2.创建基础的 logger
	var (
		logger   *zap.Logger
		cleanFns []func() //用于收集清理函数
	)

	if cfg.File.Enable {
		// 启用日志文件： 使用 lumberjack 实现日志轮转（切割）
		filename := cfg.File.Path
		_ = os.MkdirAll(filepath.Dir(filename), 0777) // 确保日子还存在

		fileWriter := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    cfg.File.MaxSize,    // mb
			MaxBackups: cfg.File.MaxBackups, // 文件数量
			Compress:   false,               // 不压缩
			LocalTime:  true,                // 使用本地时间
		}

		//注册清理函数，关闭文件句柄
		cleanFns = append(cleanFns, func() {
			_ = fileWriter.Close()
		})

		// 创建只输出到文件的 zap core
		zc := zapcore.NewCore(
			zapcore.NewJSONEncoder(zconfig.EncoderConfig),
			zapcore.AddSync(fileWriter),
			zconfig.Level,
		)
		logger = zap.New(zc)
	} else {
		// 只输出到控制台
		iLogger, err := zconfig.Build()
		if err != nil {
			return nil, err
		}
		logger = iLogger
	}

	//3.设置日志选项：调用者，堆栈，跳过层数
	skip := cfg.CallerSkip
	if skip <= 0 {
		skip = 2 //默认跳过 2 层 （本函数+调用者）
	}

	logger = logger.WithOptions(
		zap.WithCaller(true),              //显示调用者
		zap.AddStacktrace(zap.ErrorLevel), // error 以上日志级别自动记录堆栈
		zap.AddCallerSkip(skip),           // 跳过指定层数
	)

	//4. 初始化日志钩子（eg:gorm数据库日志）
	for _, h := range cfg.Hooks {
		if !h.Enable || len(hookHandle) == 0 {
			continue
		}

		//调用外部钩子处理器
		writer, err := hookHandle[0](ctx, h)
		if err != nil {
			return nil, err
		} else if writer == nil {
			continue
		}

		//注册钩子清理函数
		cleanFns = append(cleanFns, func() {
			writer.Flush()
		})

		//解析钩子日志级别
		hookLevel := zap.NewAtomicLevel()
		if level, err := zapcore.ParseLevel(h.Level); err == nil {
			hookLevel.SetLevel(level)
		} else {
			hookLevel.SetLevel(zap.InfoLevel)
		}

		// 为钩子创建独立的 encoder
		hookEncoder := zap.NewProductionEncoderConfig()
		hookEncoder.EncodeTime = zapcore.EpochMillisTimeEncoder
		hookEncoder.EncodeDuration = zapcore.MillisDurationEncoder

		// 创建钩子专用的 zap core
		hookCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(hookEncoder),
			zapcore.AddSync(writer),
			hookLevel,
		)

		logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, hookCore)
		}))
	}

	//5. 替换全局日治器
	zap.ReplaceGlobals(logger)

	return func() {
		for _, fn := range cleanFns {
			fn()
		}
	}, nil
}
