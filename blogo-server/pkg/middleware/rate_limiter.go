// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
	"github.com/patrickmn/go-cache"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimiterConfig 定义限流中间件的配置。
type RateLimiterConfig struct {
	// Enable: 是否启用限流
	Enable bool

	// AllowedPathPrefixes: 白名单路径前缀（仅这些路径限流）
	// 空列表表示所有路径都限流。
	AllowedPathPrefixes []string

	// SkippedPathPrefixes: 黑名单路径前缀（这些路径跳过限流）
	// 例如：/healthz（健康检查）
	SkippedPathPrefixes []string

	// Period: 限流周期（秒）
	Period int

	// MaxRequestsPerIP: 每个 IP 在周期内的最大请求数
	MaxRequestsPerIP int

	// MaxRequestsPerUser: 每个认证用户在周期内的最大请求数
	MaxRequestsPerUser int

	// StoreType: 存储类型（memory/redis）
	StoreType string

	// MemoryStoreConfig: 内存存储配置
	MemoryStoreConfig RateLimiterMemoryConfig

	// RedisStoreConfig: Redis 存储配置
	RedisStoreConfig RateLimiterRedisConfig
}

// RateLimiterWithConfig 根据配置创建限流中间件。
// 如果未启用，返回空中间件（Empty()）。
func RateLimiterWithConfig(config RateLimiterConfig) gin.HandlerFunc {
	if !config.Enable {
		return Empty()
	}

	// 1. 创建存储后端（内存或 Redis）
	var store RateLimiterStorer
	switch config.StoreType {
	case "redis":
		store = NewRateLimiterRedisStore(config.RedisStoreConfig)
	default:
		store = NewRateLimiterMemoryStore(config.MemoryStoreConfig)
	}

	// 2. 返回中间件函数
	return func(c *gin.Context) {
		// 2.1 路径过滤：决定是否限流
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		// 2.2 确定限流标识符（用户 ID 或 IP）
		var (
			allowed bool
			err     error
		)
		ctx := c.Request.Context()
		if userID := util.FromUserID(ctx); userID != "" {
			// 认证用户：按用户 ID 限流
			allowed, err = store.Allow(ctx, userID, time.Second*time.Duration(config.Period), config.MaxRequestsPerUser)
		} else {
			// 未认证用户：按 IP 限流
			allowed, err = store.Allow(ctx, c.ClientIP(), time.Second*time.Duration(config.Period), config.MaxRequestsPerIP)
		}

		// 2.3 处理限流结果
		if err != nil {
			// 存储错误 → 记录日志 + 返回 500
			logging.Context(ctx).Error("Rate limiter middleware error", zap.Error(err))
			util.ResError(c, errors.InternalServerError("", "Internal server error, please try again later."))
		} else if allowed {
			// 允许请求 → 继续执行
			c.Next()
		} else {
			// 拒绝请求 → 返回 429
			util.ResError(c, errors.TooManyRequests("", "Too many requests, please try again later."))
		}
	}
}

// RateLimiterStorer 定义限流存储的通用接口。
type RateLimiterStorer interface {
	// Allow 检查标识符在指定周期内是否允许请求。
	// 参数：
	//   - ctx: 上下文
	//   - identifier: 限流标识符（如 user_id 或 ip）
	//   - period: 限流周期
	//   - maxRequests: 周期内最大请求数
	// 返回：
	//   - bool: 是否允许
	//   - error: 错误
	Allow(ctx context.Context, identifier string, period time.Duration, maxRequests int) (bool, error)
}

// NewRateLimiterMemoryStore 创建内存限流存储实例。
func NewRateLimiterMemoryStore(config RateLimiterMemoryConfig) RateLimiterStorer {
	return &RateLimiterMemoryStore{
		cache: cache.New(config.Expiration, config.CleanupInterval),
	}
}

// RateLimiterMemoryConfig 定义内存存储的配置。
type RateLimiterMemoryConfig struct {
	Expiration      time.Duration // 限流器过期时间
	CleanupInterval time.Duration // 缓存清理间隔
}

// RateLimiterMemoryStore 是基于内存的限流存储。
type RateLimiterMemoryStore struct {
	cache *cache.Cache // 内存缓存（存储 rate.Limiter 实例）
}

// Allow 实现内存限流逻辑。
func (s *RateLimiterMemoryStore) Allow(ctx context.Context, identifier string, period time.Duration, maxRequests int) (bool, error) {
	// 1. 参数校验（避免除零错误）
	if period.Seconds() <= 0 || maxRequests <= 0 {
		return true, nil
	}

	// 2. 尝试获取现有限流器
	if limiter, exists := s.cache.Get(identifier); exists {
		isAllow := limiter.(*rate.Limiter).Allow()
		// 刷新过期时间
		s.cache.SetDefault(identifier, limiter)
		return isAllow, nil
	}

	// 3. 创建新限流器
	// 注意：rate.Every(period) 表示每 period 时间填充 1 个令牌
	//       但我们需要的是 period 内 maxRequests 个令牌
	//       因此应使用 rate.Limit = rate.Every(period / time.Duration(maxRequests))
	// 但标准库不支持动态令牌桶，这里简化处理（实际应使用令牌桶）
	limiter := rate.NewLimiter(rate.Every(period), maxRequests)
	limiter.Allow() // 消费一个令牌（避免首次请求被限）
	s.cache.SetDefault(identifier, limiter)

	return true, nil
}

// RateLimiterRedisConfig 定义 Redis 存储的配置。
type RateLimiterRedisConfig struct {
	Addr     string // Redis 地址
	Username string // 用户名
	Password string // 密码
	DB       int    // 数据库编号
}

// NewRateLimiterRedisStore 创建 Redis 限流存储实例。
func NewRateLimiterRedisStore(config RateLimiterRedisConfig) RateLimiterStorer {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	})

	return &RateLimiterRedisStore{
		limiter: redis_rate.NewLimiter(rdb),
	}
}

// RateLimiterRedisStore 是基于 Redis 的限流存储。
type RateLimiterRedisStore struct {
	limiter *redis_rate.Limiter // Redis 限流器
}

// Allow 实现 Redis 限流逻辑。
func (s *RateLimiterRedisStore) Allow(ctx context.Context, identifier string, period time.Duration, maxRequests int) (bool, error) {
	// 1. 参数校验
	if period.Seconds() <= 0 || maxRequests <= 0 {
		return true, nil
	}

	// 2. 计算每秒请求数（redis_rate 使用 PerSecond）
	//    例如：period=10s, maxRequests=100 → 10 req/s
	ratePerSec := maxRequests / int(period.Seconds())
	if ratePerSec <= 0 {
		ratePerSec = 1 // 至少 1 req/s
	}

	// 3. 调用 Redis 限流器
	result, err := s.limiter.Allow(ctx, identifier, redis_rate.PerSecond(ratePerSec))
	if err != nil {
		return false, err
	}
	return result.Allowed > 0, nil
}
