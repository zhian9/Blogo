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

// Image 图片表，统一管理上传的图片资源
type Image struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`            // 图片唯一ID
	URL       string    `json:"url" gorm:"size:512;not null;uniqueIndex"` // 图片访问URL（如 https://cdn.example.com/xxx.png）
	Path      string    `json:"path" gorm:"size:512;not null;"`           // 存储路径（如 /uploads/2025/10/xxx.png）
	Name      string    `json:"name" gorm:"size:255;"`                    // 原始文件名（如 architecture-diagram.png）
	Size      int64     `json:"size" gorm:"not null;"`                    // 文件大小（字节）
	Type      string    `json:"type" gorm:"size:50;index;"`               // MIME 类型（如 image/png）
	Width     int       `json:"width" gorm:"default:0;"`                  // 图片宽度（像素）
	Height    int       `json:"height" gorm:"default:0;"`                 // 图片高度（像素）
	Category  string    `json:"category" gorm:"size:50;index;"`           // 分类（如 "article_cover", "ad_image", "friend_link_logo"）
	CreatedAt time.Time `json:"created_at" gorm:"index;"`                 // 上传时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`                 // 更新时间
}

func (i *Image) TableName() string {
	return config.C.FormatTableName("image")
}

// ImageQueryParam 图片查询参数
type ImageQueryParam struct {
	util.PaginationParam
	URL      string `form:"url"`      // 按URL模糊搜索
	Category string `form:"category"` // 按分类筛选
	Type     string `form:"type"`     // 按MIME类型筛选（如 image/jpeg）
}

// ImageQueryOptions 查询选项
type ImageQueryOptions struct {
	util.QueryOptions
}

// ImageQueryResult 查询结果
type ImageQueryResult struct {
	Data       Images
	PageResult *util.PaginationResult
}

// Images 图片切片
type Images []*Image

// ToIDs 返回图片ID列表
func (i Images) ToIDs() []string {
	var ids []string
	for _, img := range i {
		ids = append(ids, img.ID)
	}
	return ids
}

// ImageForm 图片表单（通常由上传接口自动填充，但保留结构一致性）
type ImageForm struct {
	URL      string `json:"url" binding:"required"`          // 图片URL
	Path     string `json:"path" binding:"required"`         // 存储路径
	Name     string `json:"name" binding:"required,max=255"` // 原始文件名
	Size     int64  `json:"size" binding:"required,min=1"`   // 文件大小
	Type     string `json:"type" binding:"required"`         // MIME类型
	Width    int    `json:"width" binding:"min=0"`           // 宽度
	Height   int    `json:"height" binding:"min=0"`          // 高度
	Category string `json:"category" binding:"required"`     // 分类
}

// Validate 验证图片表单
func (ifm *ImageForm) Validate() error {
	// 验证是否为图片类型
	if !util.IsImageMimeType(ifm.Type) {
		return errors.BadRequest("", "Unsupported image type: %s", ifm.Type)
	}

	// 验证URL是否合法
	if !util.IsURL(ifm.URL) {
		return errors.BadRequest("", "Invalid image URL")
	}

	// 可选：限制文件大小（如 <= 10MB）
	if ifm.Size > 10*1024*1024 {
		return errors.BadRequest("", "Image size exceeds 10MB limit")
	}

	return nil
}

// FillTo 将表单数据填充到 Image 模型
func (ifm *ImageForm) FillTo(image *Image) error {
	image.URL = ifm.URL
	image.Path = ifm.Path
	image.Name = ifm.Name
	image.Size = ifm.Size
	image.Type = ifm.Type
	image.Width = ifm.Width
	image.Height = ifm.Height
	image.Category = ifm.Category
	return nil
}
