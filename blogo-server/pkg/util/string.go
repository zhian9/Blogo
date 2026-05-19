// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package util

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ContainsSpecialChars 判断字符串是否包含非字母、数字、中文、空格以外的字符（可根据需求调整）
func ContainsSpecialChars(s string) bool {
	// 允许：中文、英文、数字、空格
	matched, _ := regexp.MatchString(`[^\p{L}\p{N}\s]`, s)
	return matched
}

var slugRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func IsSlug(s string) bool {
	return slugRegexp.MatchString(s)
}

func ContainsSensitiveWords(s string) bool {
	sensitiveWords := []string{"广告", "赌博", "http://", "https://"} // 示例
	for _, word := range sensitiveWords {
		if strings.Contains(s, word) {
			return true
		}
	}
	return false
}

func IsURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func IsConfigKey(s string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, s)
	return matched
}

func IsDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// ToInt64 将字符串安全转换为 int64，失败时返回 0。
// 适用于解析 URL 查询参数、表单字段等。
func ToInt64(s string) int64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}
