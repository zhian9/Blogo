// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package jwtx

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

// Auther 定义JWT 认证器接口
type Auther interface {
	GenerateToken(ctx context.Context, subject string) (TokenInfo, error)
	GenerateTokenWithRole(ctx context.Context, subject, roleCode string) (TokenInfo, error)
	DestroyToken(ctx context.Context, accessToken string) error
	ParseCustomClaims(accessToken string) (*CustomClaims, error)
	ParseSubject(ctx context.Context, accessToken string) (string, error)
	Release(ctx context.Context) error
}

// ErrInvalidToken Token 无效错误
var ErrInvalidToken = errors.New("Invalid token")

// options 定义 JWT 认证器的内部配置。
type options struct {
	signingMethod jwt.SigningMethod                       // 签名算法（如 HS512）
	signingKey    []byte                                  // 当前签名密钥
	signingKey2   []byte                                  // 旧签名密钥（用于密钥轮换）
	keyFuncs      []func(*jwt.Token) (interface{}, error) // 密钥解析函数列表
	expired       int                                     // Token 有效期（秒）
	tokenType     string                                  // Token 类型（如 "Bearer"）
}

// Option 是配置选项的函数类型（函数式选项模式）。
type Option func(*options)

// SetSigningMethod 设置签名算法。
func SetSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}

// SetSigningKey 设置签名密钥（支持新旧密钥轮换）。
// key: 当前密钥
// oldKey: 旧密钥（用于验证使用旧密钥签名的 Token）
func SetSigningKey(key, oldKey string) Option {
	return func(o *options) {
		o.signingKey = []byte(key)
		if oldKey != "" && key != oldKey {
			o.signingKey2 = []byte(oldKey)
		}
	}
}

// SetExpired 设置 Token 有效期（秒）。
func SetExpired(expired int) Option {
	return func(o *options) {
		o.expired = expired
	}
}

// New 创建新的 JWT 认证器实例。
// store: Token 存储后端（用于黑名单）
// opts: 配置选项（签名密钥、算法、有效期等）。
// 如果未提供签名密钥，将触发 panic —— 生产环境务必设置强随机密钥。
func New(store Storer, opts ...Option) Auther {
	o := options{
		tokenType:     "Bearer",
		expired:       7200, // 2 小时
		signingMethod: jwt.SigningMethodHS512,
	}

	// 应用用户配置
	for _, opt := range opts {
		opt(&o)
	}

	if len(o.signingKey) == 0 {
		panic("jwtx: signing key is required, set JWT_SIGNING_KEY environment variable")
	}

	// 构建密钥解析函数列表
	// 1. 当前密钥
	o.keyFuncs = append(o.keyFuncs, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return o.signingKey, nil
	})

	// 2. 旧密钥（如果存在）
	if o.signingKey2 != nil {
		o.keyFuncs = append(o.keyFuncs, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return o.signingKey2, nil
		})
	}

	return &JWTAuth{
		opts:  &o,
		store: store,
	}
}

// JWTAuth 是 JWT 认证器的具体实现。
type JWTAuth struct {
	opts  *options // 配置
	store Storer   // Token 存储后端（用于黑名单）
}

// GenerateToken 生成 JWT Token。
func (a *JWTAuth) GenerateToken(ctx context.Context, subject string) (TokenInfo, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.opts.expired) * time.Second).Unix()

	// 创建 JWT Token（使用 StandardClaims）
	token := jwt.NewWithClaims(a.opts.signingMethod, &jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt,
		NotBefore: now.Unix(),
		Subject:   subject, // 用户唯一标识
	})

	// 使用当前密钥签名
	tokenStr, err := token.SignedString(a.opts.signingKey)
	if err != nil {
		return nil, err
	}

	return &tokenInfo{
		ExpiresAt:   expiresAt,
		TokenType:   a.opts.tokenType,
		AccessToken: tokenStr,
	}, nil
}

// parseToken 解析并验证 JWT Token。
// 支持多密钥轮换（依次尝试所有 keyFuncs）。
func (a *JWTAuth) parseToken(tokenStr string) (*jwt.StandardClaims, error) {
	var (
		token *jwt.Token
		err   error
	)

	// 依次尝试所有密钥解析函数
	for _, keyFunc := range a.opts.keyFuncs {
		token, err = jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, keyFunc)
		if err != nil || token == nil || !token.Valid {
			continue // 尝试下一个密钥
		}
		break // 解析成功
	}

	// 所有密钥都失败
	if err != nil || token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return token.Claims.(*jwt.StandardClaims), nil
}

// callStore 安全调用存储后端（避免空指针）。
func (a *JWTAuth) callStore(fn func(Storer) error) error {
	if store := a.store; store != nil {
		return fn(store)
	}
	return nil
}

// DestroyToken 使 Token 失效（加入黑名单）。
// 实现方式：将 Token 存入存储后端，有效期 = 原 Token 剩余时间。
func (a *JWTAuth) DestroyToken(ctx context.Context, tokenStr string) error {
	claims, err := a.parseToken(tokenStr)
	if err != nil {
		return err
	}

	return a.callStore(func(store Storer) error {
		// 计算剩余有效期（用于设置黑名单过期时间）
		expired := time.Until(time.Unix(claims.ExpiresAt, 0))
		return store.Set(ctx, tokenStr, expired)
	})
}

// ParseSubject 从 Token 中解析用户标识。
// 步骤：
//  1. 解析 Token
//  2. 检查黑名单（如果存储后端存在）
//  3. 返回用户标识
func (a *JWTAuth) ParseSubject(ctx context.Context, tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", ErrInvalidToken
	}

	// 1. 解析 Token
	claims, err := a.parseToken(tokenStr)
	if err != nil {
		return "", err
	}

	// 2. 检查黑名单
	err = a.callStore(func(store Storer) error {
		if exists, err := store.Check(ctx, tokenStr); err != nil {
			return err
		} else if exists {
			return ErrInvalidToken // Token 在黑名单中
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// 3. 返回用户标识
	return claims.Subject, nil
}

// Release 释放资源（关闭存储后端）。
func (a *JWTAuth) Release(ctx context.Context) error {
	return a.callStore(func(store Storer) error {
		return store.Close(ctx)
	})
}

// CustomClaims 扩展 StandardClaims，支持自定义字段
type CustomClaims struct {
	RoleCode string `json:"role_code,omitempty"`
	jwt.StandardClaims
}

// GenerateTokenWithRole 生成包含角色编码的 JWT Token
func (a *JWTAuth) GenerateTokenWithRole(ctx context.Context, subject, roleCode string) (TokenInfo, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.opts.expired) * time.Second).Unix()

	// 使用 CustomClaims
	claims := &CustomClaims{
		RoleCode: roleCode,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: expiresAt,
			NotBefore: now.Unix(),
			Subject:   subject,
		},
	}

	token := jwt.NewWithClaims(a.opts.signingMethod, claims)
	tokenStr, err := token.SignedString(a.opts.signingKey)
	if err != nil {
		return nil, err
	}

	return &tokenInfo{
		ExpiresAt:   expiresAt,
		TokenType:   a.opts.tokenType,
		AccessToken: tokenStr,
	}, nil
}

// ParseCustomClaims 从 Token 中解析自定义 Claims（含 role_code）
func (a *JWTAuth) ParseCustomClaims(tokenStr string) (*CustomClaims, error) {
	if tokenStr == "" {
		return nil, ErrInvalidToken
	}

	var (
		token *jwt.Token
		err   error
	)

	// 依次尝试所有密钥解析函数
	for _, keyFunc := range a.opts.keyFuncs {
		token, err = jwt.ParseWithClaims(tokenStr, &CustomClaims{}, keyFunc)
		if err != nil || token == nil || !token.Valid {
			continue
		}
		break
	}

	if err != nil || token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return token.Claims.(*CustomClaims), nil
}
