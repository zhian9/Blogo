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
	"github.com/zhian9/blogo-server/pkg/util"
)

const (
	TimelineTypeLaunch    = "launch"      // 🚀 首次发布
	TimelineTypeVersion   = "version"     // 🏷️ 版本发布
	TimelineTypeFeature   = "feature"     // ✨ 新功能
	TimelineTypeMilestone = "milestone"   // 🎯 里程碑
	TimelineTypeBreaking  = "breaking"    // ⚠️ 重大变更
	TimelineTypeArchived  = "archived"    // 📦 归档
)

// TimelineTypeTypes 所有时间线类型
var TimelineTypeTypes = []string{
	TimelineTypeLaunch,
	TimelineTypeVersion,
	TimelineTypeFeature,
	TimelineTypeMilestone,
	TimelineTypeBreaking,
	TimelineTypeArchived,
}

// ProjectTimeline 项目历程/里程碑
type ProjectTimeline struct {
	ID          string    `json:"id" gorm:"size:20;primarykey;"`           // 唯一ID
	ProjectID   string    `json:"project_id" gorm:"size:20;index;not null"` // 所属项目ID
	Title       string    `json:"title" gorm:"size:255;not null"`          // 里程碑标题
	Description string    `json:"description" gorm:"type:text;"`           // 里程碑描述
	Type        string    `json:"type" gorm:"size:20;index;default:update"` // 类型
	Version     string    `json:"version" gorm:"size:50;"`                 // 版本号（可选）
	ImageID     *string   `json:"image_id" gorm:"size:20;index"`           // 关联图片ID
	Image       *Image    `json:"image,omitempty" gorm:"-"`                // 关联图片（手动加载）
	Link        string    `json:"link" gorm:"size:512;"`                   // 关联链接
	EventDate   time.Time `json:"event_date" gorm:"index;not null"`        // 事件发生日期
	SortOrder   int       `json:"sort_order" gorm:"default:0;"`            // 排序
	CreatedAt   time.Time `json:"created_at" gorm:"index;"`                // 创建时间
	UpdatedAt   time.Time `json:"updated_at" gorm:"index;"`                // 更新时间
}

func (pt *ProjectTimeline) TableName() string {
	return config.C.FormatTableName("project_timeline")
}

// ProjectTimelineForm 时间线条目表单（创建/更新）
type ProjectTimelineForm struct {
	Title       string    `json:"title" binding:"required,max=255"`        // 标题
	Description string    `json:"description" binding:"max=2000"`          // 描述
	Type        string    `json:"type" binding:"omitempty,oneof=launch version feature milestone breaking archived"` // 类型
	Version     string    `json:"version" binding:"max=50"`                // 版本号
	ImageID     *string   `json:"image_id"`                                // 图片ID
	Link        string    `json:"link" binding:"max=512"`                  // 链接
	EventDate   time.Time `json:"event_date" binding:"required"`           // 事件日期
	SortOrder   int       `json:"sort_order"`                              // 排序
}

// Validate 验证表单
func (tf *ProjectTimelineForm) Validate() error {
	return nil
}

// FillTo 将表单数据填充到 ProjectTimeline 模型
func (tf *ProjectTimelineForm) FillTo(tl *ProjectTimeline) error {
	tl.Title = tf.Title
	tl.Description = tf.Description
	tl.Type = tf.Type
	if tl.Type == "" {
		tl.Type = TimelineTypeMilestone
	}
	tl.Version = tf.Version
	tl.ImageID = tf.ImageID
	tl.Link = tf.Link
	tl.EventDate = tf.EventDate
	tl.SortOrder = tf.SortOrder
	return nil
}

// ProjectTimelineQueryParam 查询参数
type ProjectTimelineQueryParam struct {
	util.PaginationParam
	ProjectID string `form:"project_id" binding:"required"` // 项目ID
	Type      string `form:"type"`                          // 类型筛选
}

// ProjectTimelineQueryResult 查询结果
type ProjectTimelineQueryResult struct {
	Data       ProjectTimelines
	PageResult *util.PaginationResult
}

// ProjectTimelines 时间线切片
type ProjectTimelines []*ProjectTimeline
