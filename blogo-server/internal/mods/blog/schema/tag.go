// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package schema

import (
	"time"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// Tag 标签表，用于文章打标签（如“Go”、“Docker”）
type Tag struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`             // 唯一ID
	Name      string    `json:"name" gorm:"size:100;not null;uniqueIndex"` // 标签名称（唯一）
	CreatedAt time.Time `json:"created_at" gorm:"index;"`                  // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`                  // 更新时间
}

func (t *Tag) TableName() string {
	return config.C.FormatTableName("tag")
}

// TagQueryParam 标签查询参数
type TagQueryParam struct {
	util.PaginationParam
	Name string `form:"name"` // 标签名称模糊搜索
}

// TagQueryOptions 查询选项
type TagQueryOptions struct {
	util.QueryOptions
}

// TagQueryResult 查询结果
type TagQueryResult struct {
	Data       Tags
	PageResult *util.PaginationResult
}

// Tags 标签切片
type Tags []*Tag

// ToIDs 返回标签ID列表
func (t Tags) ToIDs() []string {
	var ids []string
	for _, tag := range t {
		ids = append(ids, tag.ID)
	}
	return ids
}

// TagForm 标签表单（用于创建/更新）
type TagForm struct {
	Name string `json:"name" binding:"required,max=100"` // 标签名称
}

// Validate 验证标签表单
func (tf *TagForm) Validate() error {
	if tf.Name != "" && util.ContainsSpecialChars(tf.Name) {
		return errors.BadRequest("", "Tag name cannot contain special characters")
	}
	return nil
}

// FillTo 将表单数据填充到 Tag 模型
func (tf *TagForm) FillTo(tag *Tag) error {
	tag.Name = tf.Name
	return nil
}
