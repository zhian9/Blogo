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

	"github.com/patrickmn/go-cache"
)

type Cacher interface {
	Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error
	Get(ctx context.Context, ns, key string) (string, bool, error)
	GetAndDelete(ctx context.Context, ns, key string) (string, bool, error)
	Exists(ctx context.Context, ns, key string) (bool, error)
	Delete(ctx context.Context, ns, key string) error
	Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error
	Close(ctx context.Context) error
}

// 全局默认配置

var defaultDelimiter = ":"

type options struct {
	Delimiter string // 键分隔符
}

// Option 是函数式选项类型，用于灵活配置缓存实例
type Option func(*options)

// WithDelimiter 自定义键值对分隔符
func WithDelimiter(delimiter string) Option {
	return func(o *options) {
		o.Delimiter = delimiter
	}
}

// MemoryConfig 内存缓存配置
type MemoryConfig struct {
	CleanupInterval time.Duration //后台清理过期项的间隔
}

// NewMemoryCache 创建 memory缓存 实例
func NewMemoryCache(cfg MemoryConfig, opts ...Option) Cacher {
	// 设置默认选项
	defaultOpts := &options{
		Delimiter: defaultDelimiter,
	}

	for _, o := range opts {
		o(defaultOpts)
	}
	return &memCache{
		opts:  defaultOpts,
		cache: cache.New(0, cfg.CleanupInterval),
	}
}

// memCache 内存缓存的具体实现
type memCache struct {
	opts  *options
	cache *cache.Cache
}

// getKey 根据 ns ,key 生成完整缓存键
func (mc *memCache) getKey(ns, key string) string {
	return fmt.Sprintf("%s%s%s", ns, mc.opts.Delimiter, key)
}

// Set 设置memory
func (mc *memCache) Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error {
	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	mc.cache.Set(mc.getKey(ns, key), value, exp)
	return nil
}

func (mc *memCache) Get(ctx context.Context, ns, key string) (string, bool, error) {
	val, ok := mc.cache.Get(mc.getKey(ns, key))
	if !ok {
		return "", false, nil //键不存在
	}
	//类型断言，缓存值必须是 string
	return val.(string), ok, nil
}

func (mc *memCache) GetAndDelete(ctx context.Context, ns, key string) (string, bool, error) {
	value, ok, err := mc.Get(ctx, ns, key)
	if err != nil {
		return "", false, err
	} else if !ok { // 断言失败
		return "", false, nil
	}

	mc.cache.Delete(mc.getKey(ns, key))
	return value, true, nil
}

func (mc *memCache) Exists(ctx context.Context, ns, key string) (bool, error) {
	_, ok := mc.cache.Get(mc.getKey(ns, key))
	return ok, nil
}

func (mc *memCache) Delete(ctx context.Context, ns, key string) error {
	mc.cache.Delete(mc.getKey(ns, key))
	return nil
}

// Iterator 迭代器 遍历所有k-v
func (mc *memCache) Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error {
	nsPrefix := mc.getKey(ns, "")

	// 遍历
	for k, v := range mc.cache.Items() {
		// 检查是否以命名前缀开头
		if strings.HasPrefix(k, nsPrefix) {
			// 去除前缀,得到原始key
			originalKey := strings.TrimPrefix(k, nsPrefix)
			// 类型断言为 string
			valueStr := v.Object.(string)
			// 回调用户函数，若返回 false 则停止遍历
			if !fn(ctx, originalKey, valueStr) {
				break
			}
		}
	}
	return nil
}

// Close  释放 memory 缓存资源
func (mc *memCache) Close(ctx context.Context) error {
	mc.cache.Flush()
	return nil
}
