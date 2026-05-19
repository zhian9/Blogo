// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package util

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Trans 事务管理器
type Trans struct {
	DB *gorm.DB
}

// TransFunc 定义事务内执行的函数模型
type TransFunc func(ctx context.Context) error

func (ts *Trans) Exec(ctx context.Context, fn TransFunc) error {
	if _, ok := FromTrans(ctx); ok {
		return fn(ctx)
	}

	// 开启新事务
	return ts.DB.Transaction(func(db *gorm.DB) error {
		// 将事务 DB 存入上下文
		return fn(NewTrans(ctx, db))
	})
}

// GetDB 根据 context 返回 DB
func GetDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	db := defDB

	// 使用事务
	if tab, ok := FromTrans(ctx); ok {
		db = tab
	}

	// 添加行锁
	if FromRowLock(ctx) {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	// 注入context
	return db.WithContext(ctx)
}

// wrapQueryOptions 查询
func wrapQueryOptions(db *gorm.DB, opts QueryOptions) *gorm.DB {
	if len(opts.SelectFields) > 0 {
		db = db.Select(opts.SelectFields)
	}
	if len(opts.OmitFields) > 0 {
		db = db.Omit(opts.OmitFields...)
	}
	if len(opts.OrderFields) > 0 {
		db = db.Order(opts.OrderFields.ToSQL())
	}
	return db
}

// WrapPageQuery 分页查询入口
func WrapPageQuery(ctx context.Context, db *gorm.DB, pp PaginationParam, opts QueryOptions, out interface{}) (*PaginationResult, error) {
	if pp.OnlyCount {
		// 仅查询总数
		var count int64
		err := db.Count(&count).Error
		if err != nil {
			return nil, err
		}
		return &PaginationResult{Total: count}, nil
	} else if !pp.Pagination {
		// 不分页（但可限制返回数量）
		pageSize := pp.PageSize
		if pageSize > 0 {
			db = db.Limit(pageSize)
		}
		db = wrapQueryOptions(db, opts)
		err := db.Find(out).Error
		return nil, err
	}

	// 标准分页
	total, err := FindPage(ctx, db, pp, opts, out)
	if err != nil {
		return nil, err
	}

	return &PaginationResult{
		Total:    total,
		Current:  pp.Current,
		PageSize: pp.PageSize,
	}, nil
}

// FindPage 标准分页查询
func FindPage(ctx context.Context, db *gorm.DB, pp PaginationParam, opts QueryOptions, out interface{}) (int64, error) {
	db = db.WithContext(ctx)

	//查询总数
	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return 0, nil
	} else if count == 0 {
		return count, nil
	}

	// 计算分页参数
	current, pageSize := pp.Current, pp.PageSize
	if current > 0 && pageSize > 0 {
		db = db.Offset((current - 1) * pageSize).Limit(pageSize)
	} else if pageSize > 0 {
		db = db.Limit(pageSize)
	}

	// 查询
	db = wrapQueryOptions(db, opts)
	err = db.Find(out).Error
	return count, err
}

// FindOne 查询单条记录。
func FindOne(ctx context.Context, db *gorm.DB, opts QueryOptions, out interface{}) (bool, error) {
	db = db.WithContext(ctx)
	db = wrapQueryOptions(db, opts)
	result := db.First(out)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // 未找到，不报错
		}
		return false, err // 其他错误
	}
	return true, nil
}

// Exists 检查记录是否存在。
// 通过 COUNT(1) 实现，避免 SELECT * 性能问题。
func Exists(ctx context.Context, db *gorm.DB) (bool, error) {
	db = db.WithContext(ctx)
	var count int64
	result := db.Count(&count)
	if err := result.Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
