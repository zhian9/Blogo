// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package util

import (
	"github.com/google/uuid"
	"github.com/rs/xid"
)

// NewXID 生成xid 并返回string
func NewXID() string {
	return xid.New().String()
}

// MustNewUUID 生成uuid 并返回string
func MustNewUUID() string {
	v, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return v.String()
}
