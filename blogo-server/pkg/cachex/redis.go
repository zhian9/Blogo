// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package cachex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig redis配置结构体
type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

// NewRedisCache 根据配置创建 redis 实例
func NewRedisCache(cfg RedisConfig, opts ...Option) Cacher {
	// 创建 Redis 客户端
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 使用 cli 创建缓存实例
	return newRedisCache(cli, opts...)
}

// NewRedisCacheWithClient 使用已存在的 Redis 单机客户端创建缓存实例。
func NewRedisCacheWithClient(cli *redis.Client, opts ...Option) Cacher {
	return newRedisCache(cli, opts...)
}

// NewRedisCacheWithClusterClient 使用已存在的 Redis 集群客户端创建缓存实例。
// 适用于 Redis Cluster 部署。
func NewRedisCacheWithClusterClient(cli *redis.ClusterClient, opts ...Option) Cacher {
	return newRedisCache(cli, opts...)
}

// newRedisCache 是内部构造函数，接收通用 Redis 客户端接口。
func newRedisCache(cli redisClienter, opts ...Option) Cacher {
	// 设置默认选项
	defaultOpts := &options{
		Delimiter: defaultDelimiter, // 默认分隔符 ":"
	}

	// 应用用户传入的选项（如自定义分隔符）
	for _, o := range opts {
		o(defaultOpts)
	}

	return &redisCache{
		opts: defaultOpts,
		cli:  cli,
	}
}

// redisClienter 定义 Redis 客户端必须实现的方法。
// 兼容 redis.Client 和 redis.ClusterClient。
type redisClienter interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Close() error
}

// redisCache 是 Redis 缓存的具体实现。
type redisCache struct {
	opts *options      // 缓存选项（如分隔符）
	cli  redisClienter // Redis 客户端（单机或集群）
}

// getKey 根据命名空间（ns）和键（key）生成完整 Redis 键。
func (rc *redisCache) getKey(ns, key string) string {
	return fmt.Sprintf("%s%s%s", ns, rc.opts.Delimiter, key)
}

// Set 设置缓存项。
func (rc *redisCache) Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error {
	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	cmd := rc.cli.Set(ctx, rc.getKey(ns, key), value, exp)
	return cmd.Err()
}

// Get 获取缓存项。
func (rc *redisCache) Get(ctx context.Context, ns, key string) (string, bool, error) {
	cmd := rc.cli.Get(ctx, rc.getKey(ns, key))
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			return "", false, nil // 键不存在
		}
		return "", false, err // 其他错误
	}
	return cmd.Val(), true, nil
}

// Exists 检查缓存项是否存在。
func (rc *redisCache) Exists(ctx context.Context, ns, key string) (bool, error) {
	cmd := rc.cli.Exists(ctx, rc.getKey(ns, key))
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val() > 0, nil
}

// Delete 删除缓存项。
// 如果键不存在，不报错。
func (rc *redisCache) Delete(ctx context.Context, ns, key string) error {
	b, err := rc.Exists(ctx, ns, key)
	if err != nil {
		return err
	} else if !b {
		return nil // 键不存在，直接返回
	}

	cmd := rc.cli.Del(ctx, rc.getKey(ns, key))
	if err := cmd.Err(); err != nil && err != redis.Nil {
		return err
	}
	return nil
}

// GetAndDelete 原子性获取并删除缓存项（类似 Redis 的 GETDEL）。
func (rc *redisCache) GetAndDelete(ctx context.Context, ns, key string) (string, bool, error) {
	value, ok, err := rc.Get(ctx, ns, key)
	if err != nil {
		return "", false, err
	} else if !ok {
		return "", false, nil
	}

	cmd := rc.cli.Del(ctx, rc.getKey(ns, key))
	if err := cmd.Err(); err != nil && err != redis.Nil {
		return "", false, err
	}
	return value, true, nil
}

// Iterator 遍历命名空间下的所有键值对。
// 使用 SCAN 命令避免阻塞 Redis。
func (rc *redisCache) Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error {
	var cursor uint64 = 0

LB_LOOP:
	for {
		// 扫描匹配模式的键（如 "user:*"）
		cmd := rc.cli.Scan(ctx, cursor, rc.getKey(ns, "*"), 100)
		if err := cmd.Err(); err != nil {
			return err
		}

		keys, c, err := cmd.Result()
		if err != nil {
			return err
		}

		// 获取每个键的值并回调
		for _, key := range keys {
			cmd := rc.cli.Get(ctx, key)
			if err := cmd.Err(); err != nil {
				if err == redis.Nil {
					continue // 键已被删除
				}
				return err
			}
			// 去掉命名空间前缀，只传入原始 key
			originalKey := strings.TrimPrefix(key, rc.getKey(ns, ""))
			if next := fn(ctx, originalKey, cmd.Val()); !next {
				break LB_LOOP
			}
		}

		// SCAN 完成
		if c == 0 {
			break
		}
		cursor = c
	}
	return nil
}

// Close 关闭 Redis 客户端连接。
func (rc *redisCache) Close(ctx context.Context) error {
	return rc.cli.Close()
}
