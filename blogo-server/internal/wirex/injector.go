// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package wirex

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/redis/go-redis/v9"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods"
	"github.com/zhian9/blogo-server/pkg/cachex"
	"github.com/zhian9/blogo-server/pkg/captcha"
	"github.com/zhian9/blogo-server/pkg/gormx"
	"github.com/zhian9/blogo-server/pkg/jwtx"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// Injector 是 Wire 生成的依赖注入容器。
// 包含应用所需的核心依赖：
//   - DB: 数据库连接
//   - Cache: 缓存客户端
//   - Auth: JWT 认证器
//   - M: 业务模块集合（如 UserService, MenuService）
type Injector struct {
	DB    *gorm.DB      // GORM 数据库实例
	Cache cachex.Cacher // 缓存客户端（支持 memory/badger/redis）
	Auth  jwtx.Auther   // JWT 认证器（含 token 存储）
	M     *mods.Mods    // 业务模块（由 Wire 自动注入）
}

// InitDB 初始化 GORM 数据库连接。
// 返回：
//   - *gorm.DB: 数据库实例
//   - func(): 清理函数（关闭底层 sql.DB 连接池）
//   - error: 初始化错误
func InitDB(ctx context.Context) (*gorm.DB, func(), error) {
	cfg := config.C.Storage.DB

	// 转换读写分离配置（从 config 结构体到 gormx 结构体）
	resolver := make([]gormx.ResolverConfig, len(cfg.Resolver))
	for i, v := range cfg.Resolver {
		resolver[i] = gormx.ResolverConfig{
			DBType:   v.DBType,
			Sources:  v.Sources,
			Replicas: v.Replicas,
			Tables:   v.Tables,
		}
	}

	// 使用 gormx 创建数据库实例
	db, err := gormx.New(gormx.Config{
		Debug:       cfg.Debug,
		PrepareStmt: cfg.PrepareStmt,
		DBType:      cfg.Type,
		DSN:         cfg.DSN,
		MaxLifeTime: cfg.MaxLifeTime,
		MaxIdleTime: cfg.MaxIdleTime,
		MaxOpenConn: cfg.MaxOpenConn,
		MaxIdleConn: cfg.MaxIdleConn,
		TablePrefix: cfg.TablePrefix,
		Resolver:    resolver,
	})
	if err != nil {
		return nil, nil, err
	}

	// 返回清理函数：关闭数据库连接池
	return db, func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close() // 忽略关闭错误
		}
	}, nil
}

// InitCacher 初始化缓存客户端。
// 支持 memory / badger / redis 三种后端。
func InitCacher(ctx context.Context) (cachex.Cacher, func(), error) {
	cfg := config.C.Storage.Cache

	var cache cachex.Cacher
	switch cfg.Type {
	case "redis":
		// 创建 Redis 缓存客户端
		cache = cachex.NewRedisCache(cachex.RedisConfig{
			Addr:     cfg.Redis.Addr,
			DB:       cfg.Redis.DB,
			Username: cfg.Redis.Username,
			Password: cfg.Redis.Password,
		}, cachex.WithDelimiter(cfg.Delimiter))
		// TODO badger

	default:
		// 默认使用内存缓存
		cache = cachex.NewMemoryCache(cachex.MemoryConfig{
			CleanupInterval: time.Second * time.Duration(cfg.Memory.CleanupInterval),
		}, cachex.WithDelimiter(cfg.Delimiter))
	}

	// 返回清理函数：关闭缓存后端（如 Redis 连接、Badger DB）
	return cache, func() {
		_ = cache.Close(ctx)
	}, nil
}

// InitAuth 初始化 JWT 认证器。
// 包含签名密钥、过期时间、存储后端（用于 token 黑名单）。
func InitAuth(ctx context.Context) (jwtx.Auther, func(), error) {
	cfg := config.C.Middleware.Auth

	// 1. 配置 JWT 选项
	var opts []jwtx.Option
	opts = append(opts, jwtx.SetExpired(cfg.Expired))
	opts = append(opts, jwtx.SetSigningKey(cfg.SigningKey, cfg.OldSigningKey))

	// 设置签名算法（默认 HS512）
	var method jwt.SigningMethod
	switch cfg.SigningMethod {
	case "HS256":
		method = jwt.SigningMethodHS256
	case "HS384":
		method = jwt.SigningMethodHS384
	default:
		method = jwt.SigningMethodHS512
	}
	opts = append(opts, jwtx.SetSigningMethod(method))

	// 2. 初始化 token 存储后端（用于会话管理、黑名单）
	var cache cachex.Cacher
	switch cfg.Store.Type {
	case "redis":
		cache = cachex.NewRedisCache(cachex.RedisConfig{
			Addr:     cfg.Store.Redis.Addr,
			DB:       cfg.Store.Redis.DB,
			Username: cfg.Store.Redis.Username,
			Password: cfg.Store.Redis.Password,
		}, cachex.WithDelimiter(cfg.Store.Delimiter))
	default:
		cache = cachex.NewMemoryCache(cachex.MemoryConfig{
			CleanupInterval: time.Second * time.Duration(cfg.Store.Memory.CleanupInterval),
		}, cachex.WithDelimiter(cfg.Store.Delimiter))
	}

	// 3. 创建认证器
	auth := jwtx.New(jwtx.NewStoreWithCache(cache), opts...)

	// 返回清理函数：释放认证器资源（如关闭存储后端）
	return auth, func() {
		_ = auth.Release(ctx)
	}, nil
}

// ProvideTrans 提供事务管理器
func ProvideTrans(db *gorm.DB) *util.Trans {
	return &util.Trans{DB: db}
}

// ProvideCaptchaRedisClient 提供用于验证码的 Redis 客户端
func ProvideCaptchaRedisClient() *redis.Client {
	cfg := config.C.Util.Captcha.Redis
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

// ProvideCaptchaService 提供验证码服务
func ProvideCaptchaService(rdb *redis.Client) (*captcha.Service, error) {
	return captcha.NewService(rdb)
}
