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

// Statistics 访问统计表（按天汇总）
type Statistics struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`             // 统计ID
	Date      string    `json:"date" gorm:"size:10;uniqueIndex;not null;"` // 日期（格式：2025-10-28）
	PV        int64     `json:"pv" gorm:"default:0;"`                      // 页面浏览量
	UV        int64     `json:"uv" gorm:"default:0;"`                      // 独立访客数
	IPCount   int64     `json:"ip_count" gorm:"default:0;"`                // 独立IP数
	CreatedAt time.Time `json:"created_at" gorm:"index;"`                  // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`                  // 更新时间
}

func (s *Statistics) TableName() string {
	return config.C.FormatTableName("statistics")
}

// StatisticsQueryParam 统计查询参数
type StatisticsQueryParam struct {
	util.PaginationParam
	DateGte string `form:"date_gte"` // 日期 >=
	DateLte string `form:"date_lte"` // 日期 <=
}

// StatisticsQueryOptions 查询选项
type StatisticsQueryOptions struct {
	util.QueryOptions
}

// StatisticsQueryResult 查询结果
type StatisticsQueryResult struct {
	Data       Statisticss
	PageResult *util.PaginationResult
}

// Statisticss 统计数据切片
type Statisticss []*Statistics

// ToIDs 返回统计ID列表
func (s Statisticss) ToIDs() []string {
	var ids []string
	for _, stat := range s {
		ids = append(ids, stat.ID)
	}
	return ids
}

// StatisticsForm 统计表单（通常由定时任务填充）
type StatisticsForm struct {
	Date    string `json:"date" binding:"required"`
	PV      int64  `json:"pv" binding:"min=0"`
	UV      int64  `json:"uv" binding:"min=0"`
	IPCount int64  `json:"ip_count" binding:"min=0"`
}

// Validate 验证统计表单
func (sf *StatisticsForm) Validate() error {
	// 验证日期格式 YYYY-MM-DD
	if !util.IsDate(sf.Date) {
		return errors.BadRequest("", "Date must be in YYYY-MM-DD format")
	}
	return nil
}

// FillTo 将表单数据填充到 Statistics 模型
func (sf *StatisticsForm) FillTo(stat *Statistics) error {
	stat.Date = sf.Date
	stat.PV = sf.PV
	stat.UV = sf.UV
	stat.IPCount = sf.IPCount
	return nil
}
