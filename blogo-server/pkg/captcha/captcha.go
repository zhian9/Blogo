// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package captcha

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"github.com/zhian9/blogo-server/internal/config"
)

var (
	ErrRedisClientRequired = errors.New("redis client is required for captcha store")
	ErrInvalidCacheType    = errors.New("unsupported captcha cache_type")
)

type Service struct {
	driver base64Captcha.Driver
	store  base64Captcha.Store
}

// Generate 生成验证码，并同时存储答案和图片
func (s *Service) Generate(ctx context.Context) (id string, err error) {
	captcha := base64Captcha.NewCaptcha(s.driver, s.store)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		return "", err
	}

	// 额外存储图片 base64（用于 ResponseCaptcha）
	if redisStore, ok := s.store.(*RedisStore); ok {
		err = redisStore.SetImage(id, b64s)
		if err != nil {
			return "", err
		}
	}
	// 如果是 memory 模式，无法支持 ResponseCaptcha，可忽略

	return id, nil
}

// GetImage 获取图片 base64（供 ResponseCaptcha 使用）
func (s *Service) GetImage(ctx context.Context, id string) (string, error) {
	if redisStore, ok := s.store.(*RedisStore); ok {
		return redisStore.GetImage(id)
	}
	return "", errors.New("only redis store supports GetImage")
}

// NewService 创建验证码服务
// redisClient: 来自 config.Storage.Cache.Redis 的客户端
func NewService(redisClient *redis.Client) (*Service, error) {
	cfg := config.C.Util.Captcha

	// 1. 创建 Driver（数字验证码）
	driver := base64Captcha.NewDriverDigit(
		cfg.Height,
		cfg.Width,
		cfg.Length,
		0.7, // 噪点比例
		80,  // 字体大小
	)

	// 2. 创建 Store
	var store base64Captcha.Store

	switch cfg.CacheType {
	case "redis":
		if redisClient == nil {
			return nil, ErrRedisClientRequired
		}

		prefix := cfg.Redis.KeyPrefix
		if prefix == "" {
			prefix = "captcha:"
		}

		// 检查 Redis 连接是否正常，提前返回有助于定位启动或网络问题
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("captcha redis ping failed: %w", err)
		}

		store = NewRedisStore(redisClient, prefix, 5*time.Minute)

	case "memory":
		store = base64Captcha.DefaultMemStore

	default:
		return nil, ErrInvalidCacheType
	}

	return &Service{
		driver: driver,
		store:  store,
	}, nil
}

// Verify 验证用户输入
func (s *Service) Verify(ctx context.Context, captchaID, userInput string) bool {
	// 注意：base64Captcha.Store.Verify 不接收 context，但 RedisStore 内部可能需要
	// 所以需要确保 RedisStore 的 Verify 方法也支持 context（见下方 RedisStore 建议）
	return s.store.Verify(captchaID, userInput, true)
}
