// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package util

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/rs/xid"
)

// NewXID 生成xid 并返回string
func NewXID() string {
	return xid.New().String()
}

// RandomToken 生成加密安全的随机令牌（32 字节 → 64 位十六进制字符串）。
// 用于邮箱激活 token 等安全敏感场景：与 XID 不同，它不可预测、不可排序。
func RandomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err) // crypto/rand 读取失败属于系统级异常
	}
	return hex.EncodeToString(b)
}
