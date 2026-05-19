// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// GetStatisticsDB 根据上下文返回统计表的 GORM 查询实例
func GetStatisticsDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Statistics))
}

// Statistics 统计数据访问对象
type Statistics struct {
	DB *gorm.DB
}

// Query 根据查询参数分页查询统计数据
// 支持：日期范围查询
func (s *Statistics) Query(ctx context.Context, params schema.StatisticsQueryParam, opts ...schema.StatisticsQueryOptions) (*schema.StatisticsQueryResult, error) {
	var opt schema.StatisticsQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetStatisticsDB(ctx, s.DB)

	// 条件查询
	if v := params.DateGte; len(v) > 0 {
		db = db.Where("date >= ?", v)
	}
	if v := params.DateLte; len(v) > 0 {
		db = db.Where("date <= ?", v)
	}

	// 执行分页查询
	var list schema.Statisticss
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schema.StatisticsQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

// Get 根据日期获取单条统计记录
func (s *Statistics) Get(ctx context.Context, date string, opts ...schema.SettingQueryOptions) (*schema.Statistics, error) {
	var opt schema.SettingQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	stat := new(schema.Statistics)
	ok, err := util.FindOne(ctx, GetStatisticsDB(ctx, s.DB).Where("date = ?", date), opt.QueryOptions, stat)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return stat, nil
}

// ExistsDate 检查某日期的统计记录是否存在
func (s *Statistics) ExistsDate(ctx context.Context, date string) (bool, error) {
	ok, err := util.Exists(ctx, GetStatisticsDB(ctx, s.DB).Where("date = ?", date))
	return ok, errors.WithStack(err)
}

// Create 创建新统计记录
func (s *Statistics) Create(ctx context.Context, stat *schema.Statistics) error {
	result := GetStatisticsDB(ctx, s.DB).Create(stat)
	return errors.WithStack(result.Error)
}

// Update 更新统计记录
// selectFields: 指定更新字段，为空则更新所有字段（除 created_at）
func (s *Statistics) Update(ctx context.Context, stat *schema.Statistics, selectFields ...string) error {
	db := GetStatisticsDB(ctx, s.DB).Where("date = ?", stat.Date)

	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("date", "created_at")
	}

	result := db.Updates(stat)
	return errors.WithStack(result.Error)
}

// Delete 根据日期删除统计记录
func (s *Statistics) Delete(ctx context.Context, date string) error {
	result := GetStatisticsDB(ctx, s.DB).Where("date = ?", date).Delete(new(schema.Statistics))
	return errors.WithStack(result.Error)
}

// GetLatest 获取最近 N 天的统计数据（用于趋势图）
func (s *Statistics) GetLatest(ctx context.Context, days int) ([]schema.Statistics, error) {
	var stats []schema.Statistics
	err := GetStatisticsDB(ctx, s.DB).
		Where("date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)", days).
		Order("date DESC").
		Find(&stats).Error
	return stats, errors.WithStack(err)
}
