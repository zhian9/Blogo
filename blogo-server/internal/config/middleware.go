// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package config

// Middleware 中间件结构体
type Middleware struct {
	Recovery struct {
		Skip int `default:"3" mapstructure:"skip"`
	} `mapstructure:"recovery"`

	CORS struct {
		Enable                 bool     `mapstructure:"enable"`
		AllowAllOrigins        bool     `mapstructure:"allow_all_origins"`
		AllowOrigins           []string `mapstructure:"allow_origins"` // ✅ 必须是 []string
		AllowMethods           []string `mapstructure:"allow_methods"`
		AllowHeaders           []string `mapstructure:"allow_headers"`
		AllowCredentials       bool     `mapstructure:"allow_credentials"`
		ExposeHeaders          []string `mapstructure:"expose_headers"`
		MaxAge                 int      `mapstructure:"max_age"`
		AllowWildcard          bool     `mapstructure:"allow_wildcard"`
		AllowBrowserExtensions bool     `mapstructure:"allow_browser_extensions"`
		AllowWebSockets        bool     `mapstructure:"allow_web_sockets"`
		AllowFiles             bool     `mapstructure:"allow_files"`
	} `mapstructure:"cors"`

	Trace struct {
		SkippedPathPrefixes []string `mapstructure:"skipped_path_prefixes"`
		RequestHeaderKey    string   `default:"X-Request-Id" mapstructure:"request_header_key"`
		ResponseTraceKey    string   `default:"X-Trace-Id" mapstructure:"response_trace_key"`
	} `mapstructure:"trace"`

	Logger struct {
		SkippedPathPrefixes      []string `mapstructure:"skipped_path_prefixes"`
		MaxOutputRequestBodyLen  int      `default:"4096" mapstructure:"max_output_request_body_len"`
		MaxOutputResponseBodyLen int      `default:"1024" mapstructure:"max_output_response_body_len"`
	} `mapstructure:"logger"`

	CopyBody struct {
		SkippedPathPrefixes []string `mapstructure:"skipped_path_prefixes"`
		MaxContentLen       int64    `default:"33554432" mapstructure:"max_content_len"` // 32MB
	} `mapstructure:"copy_body"`

	Auth struct {
		Disable             bool     `mapstructure:"disable"`
		SkippedPathPrefixes []string `mapstructure:"skipped_path_prefixes"`
		SigningMethod       string   `default:"HS512" mapstructure:"signing_method"`
		SigningKey          string   `mapstructure:"signing_key"`
		OldSigningKey       string   `mapstructure:"old_signing_key"`
		Expired             int      `default:"86400" mapstructure:"expired"`
		Store               struct {
			Type      string `default:"redis" mapstructure:"type"`
			Delimiter string `default:":" mapstructure:"delimiter"`
			Memory    struct {
				CleanupInterval int `default:"60" mapstructure:"cleanup_interval"`
			} `mapstructure:"memory"`
			Badger struct {
				Path string `default:"data/auth" mapstructure:"path"`
			} `mapstructure:"badger"`
			Redis struct {
				Addr     string `mapstructure:"addr"`
				Username string `mapstructure:"username"`
				Password string `mapstructure:"password"`
				DB       int    `mapstructure:"db"`
			} `mapstructure:"redis"`
		} `mapstructure:"store"`
	} `mapstructure:"auth"`

	RateLimiter struct {
		Enable              bool     `mapstructure:"enable"`
		SkippedPathPrefixes []string `mapstructure:"skipped_path_prefixes"`
		Period              int      `mapstructure:"period"` // seconds
		MaxRequestsPerIP    int      `mapstructure:"max_requests_per_ip"`
		MaxRequestsPerUser  int      `mapstructure:"max_requests_per_user"`
		Store               struct {
			Type   string `mapstructure:"type"`
			Memory struct {
				Expiration      int `default:"3600" mapstructure:"expiration"`
				CleanupInterval int `default:"60" mapstructure:"cleanup_interval"`
			} `mapstructure:"memory"`
			Redis struct {
				Addr     string `mapstructure:"addr"`
				Username string `mapstructure:"username"`
				Password string `mapstructure:"password"`
				DB       int    `mapstructure:"db"`
			} `mapstructure:"redis"`
		} `mapstructure:"store"`
	} `mapstructure:"rate_limiter"`

	Casbin struct {
		Disable             bool     `mapstructure:"disable"`
		SkippedPathPrefixes []string `mapstructure:"skipped_path_prefixes"`
		LoadThread          int      `default:"2" mapstructure:"load_thread"`
		AutoLoadInterval    int      `default:"3" mapstructure:"auto_load_interval"` // seconds
		ModelFile           string   `default:"rbac_model.conf" mapstructure:"model_file"`
		GenPolicyFile       string   `default:"gen_rbac_policy.csv" mapstructure:"gen_policy_file"`
	} `mapstructure:"casbin"`

	Static struct {
		Dir string `mapstructure:"dir"` // Static files directory (From command arguments)
	} `mapstructure:"static"`
}
