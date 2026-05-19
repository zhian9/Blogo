// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package captcha

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mojocn/base64Captcha"
)

// RedisStore 实现 base64Captcha.Store 接口
type RedisStore struct {
	client *redis.Client
	prefix string
	expire time.Duration
}

func NewRedisStore(client *redis.Client, prefix string, expire time.Duration) base64Captcha.Store {
	return &RedisStore{
		client: client,
		prefix: prefix,
		expire: expire,
	}
}

func (r *RedisStore) SetImage(id string, b64s string) error {
	key := r.prefix + "img:" + id // 用不同 key 存图片
	return r.client.Set(context.Background(), key, b64s, r.expire).Err()
}

// GetImage 获取图片 base64
func (r *RedisStore) GetImage(id string) (string, error) {
	key := r.prefix + "img:" + id
	return r.client.Get(context.Background(), key).Result()
}

// Set 必须返回 error（接口要求）
func (r *RedisStore) Set(id string, value string) error {
	key := r.prefix + id
	return r.client.Set(context.Background(), key, value, r.expire).Err()
}

// Get 返回 string（接口要求，不返回 error）
func (r *RedisStore) Get(id string, clear bool) string {
	key := r.prefix + id
	val, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return ""
	}
	if clear {
		r.client.Del(context.Background(), key)
	}
	return val
}

// Verify 验证答案（忽略大小写）
func (r *RedisStore) Verify(id string, answer string, clear bool) bool {
	key := r.prefix + id
	val, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return false
	}
	match := strings.EqualFold(val, answer)
	if match && clear {
		r.client.Del(context.Background(), key)
	}
	return match
}
