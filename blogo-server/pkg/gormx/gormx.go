// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package gormx

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdmysql "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

// Config 定义数据库连接配置
type Config struct {
	Debug       bool             // 是否启用 GORM 调试模式（打印 SQL）
	PrepareStmt bool             // 是否启用预编译语句（提升性能）
	DBType      string           // 数据库类型：mysql/postgres/sqlite3
	DSN         string           // 数据源地址（Data Source Name）
	MaxLifeTime int              // 连接最大生命周期（秒）
	MaxIdleTime int              // 空闲连接最大存活时间（秒）
	MaxOpenConn int              // 最大打开连接数
	MaxIdleConn int              // 最大空闲连接数
	TablePrefix string           // 表名前缀
	Resolver    []ResolverConfig // 读写分离规则列表
}

// ResolverConfig 定义单个读写分离规则。
// 适用于分库分表或主从复制场景。
type ResolverConfig struct {
	DBType   string   // 数据库类型：mysql/postgres/sqlite3
	Sources  []string // 主库 DSN 列表（写操作）
	Replicas []string // 从库 DSN 列表（读操作）
	Tables   []string // 此规则适用的表名列表（空表示所有表）
}

// New 根据配置创建 GORM 数据库实例
func New(cfg Config) (*gorm.DB, error) {
	// 根据数据库类型选择对应的 GORM 驱动
	var dialector gorm.Dialector

	switch strings.ToLower(cfg.DBType) {
	case "mysql":
		// 自动创建数据库(若不存在)
		if err := createDatabaseWithMySQL(cfg.DSN); err != nil {
			return nil, fmt.Errorf("failed to create  MySQL database : %w", err)
		}
		dialector = mysql.Open(cfg.DSN)
		//TODO sqlite/postgres/mongodb
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}

	// 2.配置 GORM 行为
	ormCfg := &gorm.Config{
		// 表名策略：添加前缀 + 单数表名（User → user）
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   cfg.TablePrefix,
			SingularTable: true, //单数表名
		},
		// 默认关闭日志
		Logger: logger.Discard,
		// 启用预编译语句
		PrepareStmt: cfg.PrepareStmt,
	}

	// Debug模式 : 启用 SQL 日志
	if cfg.Debug {
		ormCfg.Logger = logger.Default
	}

	// 3. 打开数据库连接
	db, err := gorm.Open(dialector, ormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	//4. 配置读写分离
	if len(cfg.Resolver) > 0 {
		resolver := &dbresolver.DBResolver{}

		// 遍历每个读写分离规则
		for _, r := range cfg.Resolver {
			resolverCfg := dbresolver.Config{}
			var open func(dsn string) gorm.Dialector
			dbType := strings.ToLower(r.DBType)

			//根据数据库类型选择驱动
			switch dbType {
			case "mysql":
				open = mysql.Open
				//TODO sqlite/postgres
			default:
				zap.L().Warn("unsupported database type in resolver", zap.String("type", r.DBType))
				continue
			}

			// sqlite3 添加从库
			for _, replica := range r.Replicas {
				if dbType == "sqlite3" {
					_ = os.MkdirAll(filepath.Dir(replica), os.ModePerm)
				}
				resolverCfg.Replicas = append(resolverCfg.Replicas, open(replica))
			}

			// 添加主库（写）
			for _, source := range r.Sources {
				if dbType == "sqlite3" {
					_ = os.MkdirAll(filepath.Dir(source), os.ModePerm)
				}
				resolverCfg.Sources = append(resolverCfg.Sources, open(source))
			}

			// 注册规则到指定表
			tables := stringSliceToInterfaceSlice(r.Tables)
			resolver.Register(resolverCfg, tables...)

			zap.L().Info("Use resolver",
				zap.Strings("tables", r.Tables),
				zap.Strings("replicas", r.Replicas),
				zap.Strings("sources", r.Sources))
		}
		// 设置连接池参数（应用于所有 resolver 连接）
		resolver.SetMaxIdleConns(cfg.MaxIdleConn).
			SetMaxOpenConns(cfg.MaxOpenConn).
			SetConnMaxLifetime(time.Duration(cfg.MaxLifeTime) * time.Second).
			SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)

		// 启用 dbresolver 插件
		if err := db.Use(resolver); err != nil {
			return nil, fmt.Errorf("failed to use dbresolver: %w", err)
		}
	}
	// 5. 全局启用调试模式（如果需要）
	if cfg.Debug {
		db = db.Debug()
	}

	// 6. 配置底层 sql.DB 连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifeTime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)

	return db, nil
}

// stringSliceToInterfaceSlice 将 []string 转换为 []interface{}
// 用于适配 dbresolver.Register 的变参接口。
func stringSliceToInterfaceSlice(s []string) []interface{} {
	r := make([]interface{}, len(s))
	for i, v := range s {
		r[i] = v
	}
	return r
}

// createDatabaseWithMySQL 自动创建 MySQL 数据库（如果不存在）。
// 注意：仅适用于 root 用户或有 CREATE 权限的账号。
func createDatabaseWithMySQL(dsn string) error {
	// 解析 DSN 获取数据库名、用户名、密码等
	cfg, err := sdmysql.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse MySQL DSN: %w", err)
	}

	// 连接到 MySQL（不指定数据库）
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/", cfg.User, cfg.Passwd, cfg.Addr))
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer db.Close()

	// 创建数据库（如果不存在）
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET = `utf8mb4`;", cfg.DBName)
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create database `%s`: %w", cfg.DBName, err)
	}
	return nil
}
