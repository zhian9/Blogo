// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package jwtx

import (
	"context"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

var defaultDelimiter = ":"

type MemoryConfig struct {
	CleanupInterval time.Duration
}

func NewMemoryCache(cfg MemoryConfig) Cacher {
	return &memCache{
		cache: cache.New(0, cfg.CleanupInterval),
	}
}

type memCache struct {
	cache *cache.Cache
}

func (mc *memCache) getKey(ns, key string) string {
	return fmt.Sprintf("%s%s%s", ns, defaultDelimiter, key)
}

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
		return "", false, nil
	}
	return val.(string), ok, nil
}

func (mc *memCache) Exists(ctx context.Context, ns, key string) (bool, error) {
	_, ok := mc.cache.Get(mc.getKey(ns, key))
	return ok, nil
}

func (mc *memCache) Delete(ctx context.Context, ns, key string) error {
	mc.cache.Delete(mc.getKey(ns, key))
	return nil
}

func (mc *memCache) Close(ctx context.Context) error {
	mc.cache.Flush()
	return nil
}
