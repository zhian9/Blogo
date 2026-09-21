// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package util

import (
	"strings"
	"time"
)

// dateLayout 纯日期格式（前端归档、文章筛选用的就是这种）
const dateLayout = "2006-01-02"

// ParseFlexibleTime 解析时间字符串，兼容以下几种写法：
//
//	2026-07-01                     纯日期（前端归档筛选传的格式）
//	2026-07-01 15:04:05            日期 + 时间
//	2026-07-01T15:04:05Z            RFC3339（历史客户端格式，带时区）
//	2026-07-01T15:04:05+08:00       RFC3339 带偏移
//
// endOfDay 为 true 且输入是纯日期时，返回当天 23:59:59，
// 用于区间上界（否则 "2026-07-31" 会变成当天 00:00:00，把当天的文章排除掉）。
func ParseFlexibleTime(value string, endOfDay bool) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		dateLayout,
	}

	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, value, time.Local)
		if err != nil {
			continue
		}
		if endOfDay && layout == dateLayout {
			t = t.Add(24*time.Hour - time.Second)
		}
		return t, true
	}

	return time.Time{}, false
}
