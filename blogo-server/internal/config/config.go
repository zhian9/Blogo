// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

// Package config defines the application configuration structure.
package config

import (
	"encoding/json"
	"fmt"

	"github.com/zhian9/blogo-server/pkg/logging"
)

// Config 配置结构体
type Config struct {
	Logger     logging.LoggerConfig `mapstructure:"logger"`
	General    General              `mapstructure:"general"`
	Storage    Storage              `mapstructure:"storage"`
	Middleware Middleware           `mapstructure:"middleware"`
	Util       Util                 `mapstructure:"util"`
	Dictionary Dictionary           `mapstructure:"dictionary"`
	Email      EmailConfig          `mapstructure:"email"`
}

// General 通用配置结构体
type General struct {
	AppName            string `default:"goAdmin" mapstructure:"app_name"`
	Version            string `default:"v1.0.0" mapstructure:"version"`
	Debug              bool   `mapstructure:"debug"`
	PprofAddr          string `mapstructure:"pprof_addr"`
	DisableSwagger     bool   `mapstructure:"disable_swagger"`
	DisablePrintConfig bool   `mapstructure:"disable_print_config"`
	DefaultLoginPwd    string `default:"" mapstructure:"default_login_pwd"` // 新用户默认密码，建议通过 DEFAULT_LOGIN_PWD 环境变量设置
	WorkDir            string `mapstructure:"work_dir"`
	MenuFile           string `mapstructure:"menu_file"` // 菜单配置文件路径
	DenyOperateMenu    bool   `mapstructure:"deny_operate_menu"`
	AllowRegister      bool   `mapstructure:"allow_register"`                           // 是否允许公开注册
	CommentModeration  bool   `default:"true" mapstructure:"comment_moderation"`        // 评论审核开关（先审后发）
	SiteURL            string `default:"http://localhost:8040" mapstructure:"site_url"` // 站点 URL（邮件激活链接等，部署时通过 SITE_URL 环境变量设置）
	HTTP               struct {
		Addr            string `default:":8088" mapstructure:"addr"`
		ShutdownTimeout int    `default:"10" mapstructure:"shutdown_timeout"`
		ReadTimeout     int    `default:"60" mapstructure:"read_timeout"`
		WriteTimeout    int    `default:"60" mapstructure:"write_timeout"`
		IdleTimeout     int    `default:"10" mapstructure:"idle_timeout"`
		CertFile        string `mapstructure:"cert_file"`
		KeyFile         string `mapstructure:"key_file"`
	} `mapstructure:"http"`
	Root struct {
		ID       string `default:"root" mapstructure:"id"`
		Username string `default:"admin" mapstructure:"username"`
		Password string `mapstructure:"password"`
		Name     string `default:"Admin" mapstructure:"name"`
	} `mapstructure:"root"`
}

// Storage 存储配置
type Storage struct {
	R2 struct {
		AccountID       string `mapstructure:"account_id"`
		AccessKeyID     string `mapstructure:"access_key_id"`
		SecretAccessKey string `mapstructure:"secret_access_key"`
		Bucket          string `mapstructure:"bucket"`
		PublicDomain    string `mapstructure:"public_domain"`
		Endpoint        string `mapstructure:"endpoint"`
		UploadDir       string `default:"uploads" mapstructure:"upload_dir"`
	} `mapstructure:"r2"`
	Cache struct {
		Type      string `default:"memory" mapstructure:"type"`
		Delimiter string `default:":" mapstructure:"delimiter"`
		Redis     struct {
			Addr     string `default:"127.0.0.1:6379" mapstructure:"addr"`
			Username string `default:"root" mapstructure:"username"`
			Password string
			DB       int `default:"1" mapstructure:"db"`
		} `mapstructure:"redis"`
		Memory struct {
			CleanupInterval int `default:"60" mapstructure:"cleanup_interval"` // 秒
		} `mapstructure:"memory"`
		Badger struct {
			Path string `default:"data/cache" mapstructure:"path"`
		} `mapstructure:"badger"`
	} `mapstructure:"cache"`
	DB struct {
		Debug       bool   `default:"true" mapstructure:"debug"`
		Type        string `default:"mysql" mapstructure:"type"`
		DSN         string
		MaxLifeTime int    `default:"86400" mapstructure:"max_lifetime"`
		MaxIdleTime int    `default:"3600" mapstructure:"max_idle_time"`
		MaxOpenConn int    `default:"100" mapstructure:"max_open_conn"`
		MaxIdleConn int    `default:"50" mapstructure:"max_idle_conn"`
		TablePrefix string `mapstructure:"table_prefix"`
		AutoMigrate bool   `default:"true" mapstructure:"auto_migrate"`
		PrepareStmt bool   `mapstructure:"prepare_stmt"`
		Resolver    []struct {
			DBType   string   `mapstructure:"db_type"`
			Sources  []string `mapstructure:"sources"`
			Replicas []string `mapstructure:"replicas"`
			Tables   []string `mapstructure:"tables"`
		} `mapstructure:"resolver"`
	} `mapstructure:"db"`
}

// Util 工具配置结构体
type Util struct {
	Captcha struct {
		Length    int    `default:"4" mapstructure:"length"`
		Width     int    `default:"400" mapstructure:"width"`
		Height    int    `default:"160" mapstructure:"height"`
		CacheType string `default:"redis" mapstructure:"cache_type"`
		Redis     struct {
			Addr      string `mapstructure:"addr"`
			Username  string `mapstructure:"username"`
			Password  string `mapstructure:"password"`
			DB        int    `default:"1" mapstructure:"db"`
			KeyPrefix string `default:"captcha:" mapstructure:"key_prefix"`
		} `mapstructure:"redis"`
	} `mapstructure:"captcha"`
	Prometheus struct {
		Enable         bool     `mapstructure:"enable"`
		Port           int      `default:"9100" mapstructure:"port"`
		BasicUsername  string   `default:"admin" mapstructure:"basic_username"`
		BasicPassword  string   `default:"admin" mapstructure:"basic_password"`
		LogApis        []string `mapstructure:"log_apis"`
		LogMethods     []string `mapstructure:"log_methods"`
		DefaultCollect bool     `mapstructure:"default_collect"`
	} `mapstructure:"prometheus"`
}

// Dictionary 常量配置
type Dictionary struct {
	UserCacheExp int `default:"4" mapstructure:"user_cache_exp"` // 小时
}

// IsDebug 是否启用调试模式
func (c *Config) IsDebug() bool {
	return c.General.Debug
}

// String 将配置结构体格式化为带缩进的 JSON 字符串
func (c *Config) String() string {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		panic("Failed to marshal config: " + err.Error())
	}
	return string(b)
}

// PreLoad 自动复用 Redis 配置
func (c *Config) PreLoad() {
	if addr := c.Storage.Cache.Redis.Addr; addr != "" {
		username := c.Storage.Cache.Redis.Username
		password := c.Storage.Cache.Redis.Password

		if c.Util.Captcha.CacheType == "redis" && c.Util.Captcha.Redis.Addr == "" {
			c.Util.Captcha.Redis.Addr = addr
			c.Util.Captcha.Redis.Username = username
			c.Util.Captcha.Redis.Password = password
		}

		if c.Middleware.RateLimiter.Store.Type == "redis" &&
			c.Middleware.RateLimiter.Store.Redis.Addr == "" {
			c.Middleware.RateLimiter.Store.Redis.Addr = addr
			c.Middleware.RateLimiter.Store.Redis.Username = username
			c.Middleware.RateLimiter.Store.Redis.Password = password
		}

		if c.Middleware.Auth.Store.Type == "redis" &&
			c.Middleware.Auth.Store.Redis.Addr == "" {
			c.Middleware.Auth.Store.Redis.Addr = addr
			c.Middleware.Auth.Store.Redis.Username = username
			c.Middleware.Auth.Store.Redis.Password = password
		}
	}
}

// Print 打印配置（除非被禁用）
func (c *Config) Print() {
	if c.General.DisablePrintConfig {
		return
	}
	fmt.Println("// ----------------------- Load configurations start ------------------------")
	fmt.Println(c.String())
	fmt.Println("// ----------------------- Load configurations end --------------------------")
}

// FormatTableName 为表名添加前缀
// EmailConfig 邮件服务配置
type EmailConfig struct {
	Host       string `mapstructure:"host"`        // SMTP 主机
	Port       int    `mapstructure:"port"`        // SMTP 端口
	SenderName string `mapstructure:"sender_name"` // 发件人显示名
	FromEmail  string `mapstructure:"from_email"`  // 发件人邮箱
	Password   string // 从环境变量 EMAIL_PASSWORD 读取
}

func (c *Config) FormatTableName(name string) string {
	return c.Storage.DB.TablePrefix + name
}
