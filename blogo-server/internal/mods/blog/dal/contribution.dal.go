// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

type Contribution struct {
	DB *gorm.DB
}

func (c *Contribution) getDB(ctx context.Context) *gorm.DB {
	return util.GetDB(ctx, c.DB).Model(new(schema.UserContribution))
}

// QueryByUser 查询用户指定日期范围内的贡献记录
func (c *Contribution) QueryByUser(ctx context.Context, userID string, dateFrom, dateTo string) ([]*schema.UserContribution, error) {
	var list []*schema.UserContribution
	err := c.getDB(ctx).
		Where("user_id = ? AND date >= ? AND date <= ?", userID, dateFrom, dateTo).
		Order("date ASC").
		Find(&list).Error
	return list, errors.WithStack(err)
}

// Upsert 创建或更新用户某天的贡献记录（增量）
func (c *Contribution) Upsert(ctx context.Context, record *schema.UserContribution) error {
	record.UpdatedAt = time.Now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	err := c.getDB(ctx).
		Where("user_id = ? AND date = ?", record.UserID, record.Date).
		Assign(map[string]interface{}{
			"publish_count": gorm.Expr("publish_count + ?", record.PublishCount),
			"edit_count":    gorm.Expr("edit_count + ?", record.EditCount),
			"login_count":   gorm.Expr("login_count + ?", record.LoginCount),
			"total_count":   gorm.Expr("total_count + ?", record.TotalCount),
			"updated_at":    record.UpdatedAt,
		}).
		FirstOrCreate(record).Error
	return errors.WithStack(err)
}

// RecordPublish 记录一次发布贡献
func (c *Contribution) RecordPublish(ctx context.Context, userID string) error {
	now := time.Now()
	date := now.Format("2006-01-02")
	return c.Upsert(ctx, &schema.UserContribution{
		ID:           util.NewXID(),
		UserID:       userID,
		Date:         date,
		PublishCount: 1,
		TotalCount:   1,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

// RecordEdit 记录一次编辑贡献
func (c *Contribution) RecordEdit(ctx context.Context, userID string) error {
	now := time.Now()
	date := now.Format("2006-01-02")
	return c.Upsert(ctx, &schema.UserContribution{
		ID:         util.NewXID(),
		UserID:     userID,
		Date:       date,
		EditCount:  1,
		TotalCount: 1,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
}

// ComputeContributions 根据文章数据计算用户贡献。
// 返回最近 365 天的每日贡献列表。
// 该函数不依赖 user_contribution 表，直接查询 article 表聚合数据。
func (c *Contribution) ComputeContributions(ctx context.Context, userID string) ([]*schema.ContributionDay, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dateFrom := today.AddDate(0, 0, -364).Format("2006-01-02")
	dateTo := today.Format("2006-01-02")

	articleTable := (&schema.Article{}).TableName()
	db := util.GetDB(ctx, c.DB)

	// 1. 查询文章发布贡献（按日期聚合）
	type dayAgg struct {
		Date string
		Cnt  int
	}
	var publishAgg []dayAgg
	if err := db.Raw(
		"SELECT DATE(published_at) as date, COUNT(*) as cnt FROM `"+articleTable+"` WHERE author_id = ? AND status = ? AND published_at >= ? AND published_at <= ? GROUP BY DATE(published_at)",
		userID, "published", dateFrom, dateTo,
	).Scan(&publishAgg).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	pubMap := make(map[string]int, len(publishAgg))
	for _, a := range publishAgg {
		pubMap[a.Date] = a.Cnt
	}

	// 2. 查询编辑贡献（按日期聚合，updated_at > published_at 表示文章被修改过）
	var editAgg []dayAgg
	if err := db.Raw(
		"SELECT DATE(updated_at) as date, COUNT(*) as cnt FROM `"+articleTable+"` WHERE author_id = ? AND updated_at >= ? AND updated_at <= ? AND updated_at > published_at GROUP BY DATE(updated_at)",
		userID, dateFrom, dateTo,
	).Scan(&editAgg).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	editMap := make(map[string]int, len(editAgg))
	for _, a := range editAgg {
		editMap[a.Date] = a.Cnt
	}

	// 3. 查询已存储的贡献记录（UserContribution 表，可选，不存在则跳过）
	existMap := make(map[string]*schema.UserContribution)
	if existing, err := c.QueryByUser(ctx, userID, dateFrom, dateTo); err == nil {
		for _, e := range existing {
			existMap[e.Date] = e
		}
	}

	// 4. 构建 365 天数据
	days := make([]*schema.ContributionDay, 0, 365)
	for i := 0; i < 365; i++ {
		d := today.AddDate(0, 0, -364+i)
		dateStr := d.Format("2006-01-02")
		day := &schema.ContributionDay{Date: dateStr}
		// 合并文章查询结果
		if pub, ok := pubMap[dateStr]; ok {
			day.PublishCount = pub
			day.Count += pub
		}
		if ed, ok := editMap[dateStr]; ok {
			day.EditCount = ed
			day.Count += ed
		}
		// 合并存储的贡献记录
		if e, ok := existMap[dateStr]; ok {
			day.PublishCount += e.PublishCount
			day.EditCount += e.EditCount
			day.LoginCount += e.LoginCount
			day.Count += e.TotalCount
		}
		days = append(days, day)
	}
	return days, nil
}

// ComputeStats 计算贡献统计
func (c *Contribution) ComputeStats(days []*schema.ContributionDay) *schema.ContributionStats {
	today := time.Now().Truncate(24 * time.Hour).Format("2006-01-02")

	var total int
	var longestStreak int
	streak := 0

	// 遍历所有天，计算总贡献和最长连续天数
	for _, day := range days {
		total += day.Count
		if day.Count > 0 {
			streak++
			if streak > longestStreak {
				longestStreak = streak
			}
		} else {
			streak = 0
		}
	}

	// 从最后一天向前计算当前连续天数（今天无贡献也算连续，因为今天还没过完）
	currentStreak := 0
	checkDate := today
	for i := len(days) - 1; i >= 0; i-- {
		d, _ := time.Parse("2006-01-02", days[i].Date)
		if d.Format("2006-01-02") > checkDate {
			continue
		}
		if days[i].Count > 0 {
			currentStreak++
			checkDate = days[i].Date
		} else if days[i].Date == today {
			// 今天无贡献但还没过完 → 不算中断
			continue
		} else {
			break
		}
	}

	// 活跃等级
	level := "Lv.1 Beginner"
	switch {
	case total >= 1000:
		level = "Lv.5 Legend"
	case total >= 500:
		level = "Lv.4 Creator"
	case total >= 200:
		level = "Lv.3 Writer"
	case total >= 50:
		level = "Lv.2 Contributor"
	}

	return &schema.ContributionStats{
		TotalContributions: total,
		CurrentStreak:      currentStreak,
		LongestStreak:      longestStreak,
		ActiveLevel:        level,
	}
}
