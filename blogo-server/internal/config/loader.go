// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var (
	once sync.Once
	C    = new(Config)
)

// MustLoad loads configuration and panics on error.
func MustLoad(dir string, names ...string) {
	once.Do(func() {
		if err := Load(dir, names...); err != nil {
			panic(fmt.Errorf("failed to load config: %w", err))
		}
	})
}

// Load loads config files using Viper.
func Load(dir string, names ...string) error {
	// 0. 首先加载 .env 文件 - 使用更可靠的方法
	if err := loadEnvFileReliable(); err != nil {
		fmt.Printf("Warning: Failed to load .env file: %v\n", err)
	} else {
		fmt.Println("Successfully loaded .env file")
	}

	// 打印关键环境变量状态（脱敏）
	printEnvStatus()

	// 1. Set default values from struct tags
	if err := defaults.Set(C); err != nil {
		return fmt.Errorf("failed to set defaults: %w", err)
	}

	v := viper.New()

	// 自动绑定环境变量
	v.AutomaticEnv()

	// 设置环境变量前缀（可选）
	v.SetEnvPrefix("blogo")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 允许空的环境变量
	v.AllowEmptyEnv(true)

	v.AddConfigPath(dir)

	// 2. Support multiple formats
	supportedExts := map[string]string{
		".json": "json",
		".toml": "toml",
		".yaml": "yaml",
		".yml":  "yaml",
	}

	// 3. Load each config name (file or dir)
	for _, name := range names {
		fullPath := filepath.Join(dir, name)
		info, err := os.Stat(fullPath)
		if err != nil {
			return fmt.Errorf("config not found: %s", fullPath)
		}

		if info.IsDir() {
			// Walk directory and merge all supported files
			err := filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				ext := strings.ToLower(filepath.Ext(path))
				if typ, ok := supportedExts[ext]; ok {
					v.SetConfigType(typ)
					v.SetConfigFile(path)
					if err := v.MergeInConfig(); err != nil {
						return fmt.Errorf("failed to merge %s: %w", path, err)
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			// Load single file
			ext := strings.ToLower(filepath.Ext(fullPath))
			if typ, ok := supportedExts[ext]; ok {
				v.SetConfigType(typ)
				v.SetConfigFile(fullPath)
				if err := v.ReadInConfig(); err != nil {
					return fmt.Errorf("failed to read %s: %w", fullPath, err)
				}
			}
		}
	}

	// 4. 在解码前展开所有环境变量占位符
	allSettings := v.AllSettings()
	fmt.Printf("Raw config settings before env expansion: %+v\n", maskSensitiveConfig(allSettings))

	expandedSettings := expandEnvVars(allSettings)
	fmt.Printf("Config settings after env expansion: %+v\n", maskSensitiveConfig(expandedSettings))

	// 5. Decode into struct using mapstructure
	decoderConfig := &mapstructure.DecoderConfig{
		TagName: "mapstructure",
		Result:  C,
		// 添加解码钩子来处理字符串替换
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}
	if err := decoder.Decode(expandedSettings); err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}

	// 6. 强制替换关键配置项的环境变量
	forceEnvVarsReplacement()

	// 7. 验证关键配置项是否已正确替换
	if err := validateConfig(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 8. Run pre-load logic
	C.PreLoad()

	return nil
}

// loadEnvFileReliable 更可靠的 .env 文件加载方法
func loadEnvFileReliable() error {
	// 尝试多个可能的 .env 文件路径
	possiblePaths := []string{
		".env",                         // 当前目录
		filepath.Join(".", ".env"),     // 当前目录
		filepath.Join("..", ".env"),    // 上级目录
		filepath.Join("../..", ".env"), // 上两级目录
	}

	for _, envPath := range possiblePaths {
		fmt.Printf("Trying to load .env from: %s\n", envPath)
		if _, err := os.Stat(envPath); err == nil {
			content, err := os.ReadFile(envPath)
			if err != nil {
				fmt.Printf("Failed to read .env file at %s: %v\n", envPath, err)
				continue
			}

			lines := strings.Split(string(content), "\n")
			envCount := 0
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					// 移除值中的引号
					value = strings.Trim(value, `"'`)

					// 总是设置环境变量，覆盖可能存在的空值
					os.Setenv(key, value)
					envCount++
					fmt.Printf("Set env: %s=***\n", key) // 不打印敏感信息
				}
			}
			fmt.Printf("Successfully loaded %d environment variables from %s\n", envCount, envPath)
			return nil
		} else {
			fmt.Printf(" .env file not found at %s: %v\n", envPath, err)
		}
	}
	return fmt.Errorf("no .env file found in any of the searched paths")
}

// printEnvStatus 打印关键环境变量状态
func printEnvStatus() {
	criticalVars := []string{"DB_DSN", "REDIS_PASSWORD", "ROOT_PASSWORD_HASH"}

	fmt.Println("=== Environment Variables Status ===")
	for _, envVar := range criticalVars {
		value := os.Getenv(envVar)
		if value == "" {
			fmt.Printf("❌ %s: NOT SET\n", envVar)
		} else {
			// 对敏感信息进行脱敏显示
			maskedValue := "***"
			if envVar == "DB_DSN" {
				maskedValue = maskDSN(value)
			} else if len(value) > 3 {
				maskedValue = value[:3] + "***"
			}
			fmt.Printf("✅ %s: %s\n", envVar, maskedValue)
		}
	}
	fmt.Println("===================================")
}

// expandEnvVars 递归展开配置中的环境变量占位符
func expandEnvVars(config interface{}) interface{} {
	switch v := config.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = expandEnvVars(value)
		}
		return result
	case []interface{}:
		var result []interface{}
		for _, item := range v {
			result = append(result, expandEnvVars(item))
		}
		return result
	case string:
		// 使用 os.ExpandEnv 替换 ${VAR} 格式的环境变量
		expanded := os.ExpandEnv(v)
		// 如果替换后仍然包含 ${，说明环境变量未设置
		if strings.Contains(expanded, "${") {
			fmt.Printf("Warning: Environment variable not set in string: %s\n", v)
		}
		return expanded
	default:
		return config
	}
}

// maskSensitiveConfig 对敏感配置信息进行脱敏
func maskSensitiveConfig(config interface{}) interface{} {
	switch v := config.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			// 对敏感字段进行脱敏
			if shouldMaskKey(key) {
				result[key] = "***"
			} else {
				result[key] = maskSensitiveConfig(value)
			}
		}
		return result
	case []interface{}:
		var result []interface{}
		for _, item := range v {
			result = append(result, maskSensitiveConfig(item))
		}
		return result
	case string:
		// 如果是DSN，进行脱敏
		if strings.Contains(v, "@tcp") {
			return maskDSN(v)
		}
		return v
	default:
		return config
	}
}

// shouldMaskKey 判断是否需要脱敏的键
func shouldMaskKey(key string) bool {
	sensitiveKeys := []string{"password", "Password", "PASSWORD", "pass", "Pass", "PASS",
		"secret", "Secret", "SECRET", "token", "Token", "TOKEN", "key", "Key", "KEY", "dsn", "DSN"}

	for _, sensitive := range sensitiveKeys {
		if strings.Contains(strings.ToLower(key), strings.ToLower(sensitive)) {
			return true
		}
	}
	return false
}

// forceEnvVarsReplacement 强制替换关键配置项的环境变量
func forceEnvVarsReplacement() {
	// 直接手动设置关键配置项，确保环境变量被正确替换
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		C.Storage.DB.DSN = dsn
		fmt.Printf("Manually set DB DSN from environment variable\n")
	}

	if redisPass := os.Getenv("REDIS_PASSWORD"); redisPass != "" {
		C.Storage.Cache.Redis.Password = redisPass
		C.Middleware.Auth.Store.Redis.Password = redisPass
		C.Util.Captcha.Redis.Password = redisPass
		fmt.Printf("Manually set Redis passwords from environment variable\n")
	}

	if rootPass := os.Getenv("ROOT_PASSWORD_HASH"); rootPass != "" {
		C.General.Root.Password = rootPass
		fmt.Printf("Manually set root password from environment variable\n")

		// R2 对對象存储环境变量（可选）
		if v := os.Getenv("R2_ACCOUNT_ID"); v != "" {
			C.Storage.R2.AccountID = v
		}
		if v := os.Getenv("R2_ACCESS_KEY_ID"); v != "" {
			C.Storage.R2.AccessKeyID = v
		}
		if v := os.Getenv("R2_SECRET_ACCESS_KEY"); v != "" {
			C.Storage.R2.SecretAccessKey = v
		}
		if v := os.Getenv("R2_BUCKET"); v != "" {
			C.Storage.R2.Bucket = v
		}
		if v := os.Getenv("R2_PUBLIC_DOMAIN"); v != "" {
			C.Storage.R2.PublicDomain = v
		}
	}

	// Email password from environment variable
	if v := os.Getenv("EMAIL_PASSWORD"); v != "" {
		C.Email.Password = v
		fmt.Printf("Manually set email password from environment variable\n")
	}
}

// validateConfig 验证配置是否正确
func validateConfig() error {
	// 检查数据库 DSN
	if C.Storage.DB.DSN == "" {
		return fmt.Errorf("database DSN is empty - environment variables may not be loaded correctly")
	}
	if strings.Contains(C.Storage.DB.DSN, "${") {
		return fmt.Errorf("database DSN contains unresolved environment variables: %s", C.Storage.DB.DSN)
	}

	// 检查 DSN 格式
	if !strings.Contains(C.Storage.DB.DSN, "/") {
		return fmt.Errorf("invalid DSN format: missing database name separator '/'. DSN: %s", maskDSN(C.Storage.DB.DSN))
	}

	fmt.Printf("✅ Config validation passed. DSN: %s\n", maskDSN(C.Storage.DB.DSN))
	return nil
}

// maskDSN 隐藏 DSN 中的密码
func maskDSN(dsn string) string {
	if dsn == "" {
		return "empty"
	}

	parts := strings.Split(dsn, "@")
	if len(parts) != 2 {
		return "invalid-dsn-format"
	}

	userPass := parts[0]
	userPassParts := strings.Split(userPass, ":")
	if len(userPassParts) != 2 {
		return "invalid-user-pass-format"
	}

	return userPassParts[0] + ":***@" + parts[1]
}
