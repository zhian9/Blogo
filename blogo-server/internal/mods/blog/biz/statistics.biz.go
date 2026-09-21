// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	rschema "github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/cachex"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

type Statistics struct {
	DB            *gorm.DB        // 数据库连接（用于跨表聚合查询）
	Trans         *util.Trans     // 事务管理器
	StatisticsDAL *dal.Statistics // 统计数据访问层
	Cache         cachex.Cacher   // 缓存客户端（访问量计数：PV/UV/IP）
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
	if err := form.Validate(); err != nil {
		return nil, err
	}

	exists, err := s.StatisticsDAL.ExistsDate(ctx, form.Date)
	if err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Statistics for this date already exists")
	}

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

// GetLatest 获取最近 N 天统计数据（用于趋势图，按日期升序，缺失日期补 0）
func (s *Statistics) GetLatest(ctx context.Context, days int) ([]schema.Statistics, error) {
	if days <= 0 {
		days = 7
	}

	rows, err := s.StatisticsDAL.GetLatest(ctx, days)
	if err != nil {
		return nil, err
	}

	byDate := make(map[string]schema.Statistics, len(rows))
	for _, row := range rows {
		byDate[row.Date] = row
	}

	// 折线图需要连续的时间轴：没有数据的日期补 0，避免前后端各自补洞
	result := make([]schema.Statistics, 0, days)
	today := time.Now()
	for i := days - 1; i >= 0; i-- {
		date := today.AddDate(0, 0, -i).Format(statisticsDateLayout)
		if row, ok := byDate[date]; ok {
			result = append(result, row)
			continue
		}
		result = append(result, schema.Statistics{Date: date})
	}
	return result, nil
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

// statisticsDateLayout 统计日期格式：Redis 计数键与 statistics.date 都用它
const statisticsDateLayout = "2006-01-02"

// visitCounterTTL 访问计数在 Redis 里的保留时间。
// 跨天后仍要能读到昨天的累计值用于落库，所以给足两天。
const visitCounterTTL = 48 * time.Hour

// RecordVisit 记录一次页面访问：PV +1、记录访客与 IP 去重标记，然后把当天累计值写入 statistics 表。
//
// 计数放在 Redis 而不是每次直接写库，是为了避免「每个页面访问一次数据库写」：
//   - pv:<date>            当日浏览量（读改写）
//   - uv:<date>:<visitor>  访客去重标记（Set 幂等，重复访问不会重复计数）
//   - ip:<date>:<ip>       独立 IP 去重标记
//
// 落库走 DAL.Upsert（按 date 覆盖）写入绝对值，因此接口被重复调用也不会重复累加。
func (s *Statistics) RecordVisit(ctx context.Context, visitorID, clientIP string) error {
	if s.Cache == nil {
		return fmt.Errorf("statistics: cache client is not initialized")
	}

	date := time.Now().Format(statisticsDateLayout)
	ns := config.CacheNSForStats

	// 1. PV +1（键带 TTL，过期自动清理）
	pv := int64(1)
	if raw, ok, err := s.Cache.Get(ctx, ns, "pv:"+date); err != nil {
		logging.Context(ctx).Error("record visit: read pv failed", zap.Error(err))
	} else if ok {
		if v, convErr := strconv.ParseInt(raw, 10, 64); convErr == nil {
			pv = v + 1
		}
	}
	if err := s.Cache.Set(ctx, ns, "pv:"+date, strconv.FormatInt(pv, 10), visitCounterTTL); err != nil {
		logging.Context(ctx).Error("record visit: save pv failed", zap.Error(err))
	}

	// 2. 访客 / IP 去重标记（同一个键重复写入不会增加计数）
	if visitor := sanitizeVisitKey(visitorID); visitor != "" {
		if err := s.Cache.Set(ctx, ns, "uv:"+date+":"+visitor, "1", visitCounterTTL); err != nil {
			logging.Context(ctx).Error("record visit: save uv failed", zap.Error(err))
		}
	}
	if ip := sanitizeVisitKey(clientIP); ip != "" {
		if err := s.Cache.Set(ctx, ns, "ip:"+date+":"+ip, "1", visitCounterTTL); err != nil {
			logging.Context(ctx).Error("record visit: save ip failed", zap.Error(err))
		}
	}

	// 3. 统计当天独立访客数与独立 IP 数
	uv, ipCount := s.countVisitDistinct(ctx, ns, date)

	now := time.Now()
	return s.StatisticsDAL.Upsert(ctx, &schema.Statistics{
		ID:        util.NewXID(),
		Date:      date,
		PV:        pv,
		UV:        uv,
		IPCount:   ipCount,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// countVisitDistinct 遍历当天写入的去重标记，返回（独立访客数、独立 IP 数）。
func (s *Statistics) countVisitDistinct(ctx context.Context, ns, date string) (uv, ipCount int64) {
	uvPrefix := "uv:" + date + ":"
	ipPrefix := "ip:" + date + ":"

	err := s.Cache.Iterator(ctx, ns, func(_ context.Context, key, _ string) bool {
		switch {
		case strings.HasPrefix(key, uvPrefix):
			uv++
		case strings.HasPrefix(key, ipPrefix):
			ipCount++
		}
		return true
	})
	if err != nil {
		logging.Context(ctx).Error("record visit: count distinct failed", zap.Error(err))
	}
	return uv, ipCount
}

// sanitizeVisitKey 去掉缓存键分隔符，避免访客标识破坏键结构。
func sanitizeVisitKey(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), ":", "_")
}
