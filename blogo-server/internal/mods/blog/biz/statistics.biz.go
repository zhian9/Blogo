// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	rschema "github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

type Statistics struct {
	DB            *gorm.DB        // 数据库连接（用于跨表聚合查询）
	Trans         *util.Trans     // 事务管理器
	StatisticsDAL *dal.Statistics // 统计数据访问层
}

// Query 查询统计数据
func (s *Statistics) Query(ctx context.Context, params schema.StatisticsQueryParam) (*schema.StatisticsQueryResult, error) {
	params.Pagination = true
	return s.StatisticsDAL.Query(ctx, params, schema.StatisticsQueryOptions{})
}

// Get 获取某日统计数据
func (s *Statistics) Get(ctx context.Context, date string) (*schema.Statistics, error) {
	stat, err := s.StatisticsDAL.Get(ctx, date)
	if err != nil {
		return nil, err
	} else if stat == nil {
		return nil, errors.NotFound("", "Statistics not found")
	}
	return stat, nil
}

// Create 创建统计数据（通常由定时任务调用）
func (s *Statistics) Create(ctx context.Context, form *schema.StatisticsForm) (*schema.Statistics, error) {
	// 1. 校验日期格式
	if err := form.Validate(); err != nil {
		return nil, err
	}

	// 2. 检查是否已存在
	exists, err := s.StatisticsDAL.ExistsDate(ctx, form.Date)
	if err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Statistics for this date already exists")
	}

	// 3. 事务内创建
	err = s.Trans.Exec(ctx, func(ctx context.Context) error {
		stat := &schema.Statistics{
			ID:        util.NewXID(),
			CreatedAt: time.Now(),
		}
		form.FillTo(stat)
		return s.StatisticsDAL.Create(ctx, stat)
	})
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, form.Date)
}

// Update 更新统计数据（谨慎使用）
func (s *Statistics) Update(ctx context.Context, date string, form *schema.StatisticsForm) error {
	exists, err := s.StatisticsDAL.ExistsDate(ctx, date)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Statistics not found")
	}

	if err := form.Validate(); err != nil {
		return err
	}

	return s.Trans.Exec(ctx, func(ctx context.Context) error {
		stat := &schema.Statistics{Date: date}
		form.FillTo(stat)
		return s.StatisticsDAL.Update(ctx, stat)
	})
}

// GetLatest 获取最近 N 天统计数据（用于趋势图）
func (s *Statistics) GetLatest(ctx context.Context, days int) ([]schema.Statistics, error) {
	return s.StatisticsDAL.GetLatest(ctx, days)
}

// GetPublicStats 获取首页公开聚合统计（文章数、分类数、用户数）
func (s *Statistics) GetPublicStats(ctx context.Context) (*schema.PublicStats, error) {
	var stats schema.PublicStats

	if err := s.DB.Model(&schema.Article{}).Where("status = ?", schema.ArticleStatusPublished).Count(&stats.ArticleCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	if err := s.DB.Model(&schema.Category{}).Count(&stats.CategoryCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	if err := s.DB.Model(&rschema.User{}).Where("status = ?", rschema.UserStatusActivated).Count(&stats.UserCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &stats, nil
}
