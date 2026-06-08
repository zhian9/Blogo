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
)

const (
	ResourceTypeDocument = "document" // 文档
	ResourceTypeVideo    = "video"    // 视频
	ResourceTypeSlide    = "slide"    // 幻灯片
	ResourceTypeArticle  = "article"  // 博客文章
	ResourceTypeDesign   = "design"   // 设计稿
	ResourceTypeOther    = "other"    // 其他
)

// ProjectResourceTypeTypes 所有资源类型
var ProjectResourceTypeTypes = []string{
	ResourceTypeDocument,
	ResourceTypeVideo,
	ResourceTypeSlide,
	ResourceTypeArticle,
	ResourceTypeDesign,
	ResourceTypeOther,
}

// ProjectResource 项目相关资源
type ProjectResource struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`            // 唯一ID
	ProjectID string    `json:"project_id" gorm:"size:20;index;not null"` // 所属项目ID
	Title     string    `json:"title" gorm:"size:255;not null"`          // 资源标题
	URL       string    `json:"url" gorm:"size:512;not null"`            // 资源链接
	Type      string    `json:"type" gorm:"size:20;index;default:other"` // 资源类型
	SortOrder int       `json:"sort_order" gorm:"default:0;"`            // 排序
	CreatedAt time.Time `json:"created_at" gorm:"index;"`                // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`                // 更新时间
}

func (pr *ProjectResource) TableName() string {
	return config.C.FormatTableName("project_resource")
}

// ProjectResources 资源切片
type ProjectResources []*ProjectResource

// ProjectResourceForm 资源表单（创建/更新）
type ProjectResourceForm struct {
	Title     string `json:"title" binding:"required,max=255"`                                        // 标题
	URL       string `json:"url" binding:"required,max=512"`                                          // 链接
	Type      string `json:"type" binding:"omitempty,oneof=document video slide article design other"` // 类型
	SortOrder int    `json:"sort_order"`                                                              // 排序
}

// Validate 验证表单
func (rf *ProjectResourceForm) Validate() error {
	return nil
}

// FillTo 将表单数据填充到 ProjectResource 模型
func (rf *ProjectResourceForm) FillTo(res *ProjectResource) error {
	res.Title = rf.Title
	res.URL = rf.URL
	res.Type = rf.Type
	if res.Type == "" {
		res.Type = ResourceTypeOther
	}
	res.SortOrder = rf.SortOrder
	return nil
}
