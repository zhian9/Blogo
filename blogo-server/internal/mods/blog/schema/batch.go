// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package schema

// BatchUpdateStatusForm 批量更新文章状态请求体
type BatchUpdateStatusForm struct {
	IDs    []string `json:"ids"`
	Status string   `json:"status"` // draft | published
}

// BatchIDsForm 通用的批量ID请求体（用于批量删除等）
type BatchIDsForm struct {
	IDs []string `json:"ids"`
}
