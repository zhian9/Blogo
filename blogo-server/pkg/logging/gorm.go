// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package logging

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

type Logger struct {
	// 唯一 ID，使用 xid 生成（20 字符，比 UUID 短）
	ID string `gorm:"size:20;primaryKey;" json:"id"`
	// 日志级别（debug/info/warn/error 等）
	Level string `gorm:"size:20;index;" json:"level"`
	// 分布式链路追踪 ID，用于关联同一请求的多个日志
	TraceID string `gorm:"size:64;index;" json:"trace_id"`
	// 操作用户 ID（可为空）
	UserID string `gorm:"size:20;index;" json:"user_id"`
	// 日志分类标签（如 "login", "system"）
	Tag string `gorm:"size:32;index;" json:"tag"`
	// 日志主消息内容（来自 zap 的 msg 字段）
	Message string `gorm:"size:1024;" json:"message"`
	// 错误堆栈信息（仅 error 及以上级别有）
	Stack string `gorm:"type:text;" json:"stack"`
	// 其他结构化字段的 JSON 序列化结果（如 request_id, ip, duration 等）
	Data string `gorm:"type:text;" json:"data"`
	// 日志创建时间（自动索引）
	CreatedAt time.Time `gorm:"index;" json:"created_at"`
}

// NewGormHook 创建一个新的Gorm 日志钩子实例
func NewGormHook(db *gorm.DB) *GormHook {
	//自动迁移
	err := db.AutoMigrate(new(Logger))
	if err != nil {
		panic(err)
	}

	return &GormHook{
		db: db,
	}
}

// GormHook 是日志写入器
type GormHook struct {
	db *gorm.DB //GORM 数据库链接
}

// Exec 将 zap 输出的 JSON 日志解析并写入数据库。
func (h *GormHook) Exec(extra map[string]string, b []byte) error {
	//1.创建日志实体，生成唯一ID
	msg := &Logger{
		ID: xid.New().String(),
	}
	// 2.使用高性能  jsoniter 解析日志 JSON
	data := make(map[string]interface{})
	err := jsoniter.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	// 3.提取zap标准字段到 Logger结构体对应字段
	if v, ok := data["ts"]; ok {
		if tsFloat, ok := v.(float64); ok {
			msg.CreatedAt = time.UnixMilli(int64(tsFloat))
		}
		delete(data, "ts")
	}

	if v, ok := data["msg"]; ok {
		if msgStr, ok := v.(string); ok {
			msg.Message = msgStr
		}
		delete(data, "msg")
	}
	if v, ok := data["tag"]; ok {
		if tagStr, ok := v.(string); ok {
			msg.Tag = tagStr
		}
		delete(data, "tag")
	}

	if v, ok := data["trace_id"]; ok {
		if tidStr, ok := v.(string); ok {
			msg.TraceID = tidStr
		}
		delete(data, "trace_id")
	}

	if v, ok := data["user_id"]; ok {
		if uidStr, ok := v.(string); ok {
			msg.UserID = uidStr
		}
		delete(data, "user_id")
	}

	if v, ok := data["level"]; ok {
		if levelStr, ok := v.(string); ok {
			msg.Level = levelStr
		}
		delete(data, "level")
	}

	if v, ok := data["stack"]; ok {
		if stackStr, ok := v.(string); ok {
			msg.Stack = stackStr
		}
		delete(data, "stack")
	}

	// 移除调用者信息（通常不需要存入数据库）
	delete(data, "caller")

	// 4. 合并额外参数（如 Hook 配置中的 extra）
	for k, v := range extra {
		data[k] = v
	}

	// 5. 将剩余字段序列化为 JSON 存入 Data 字段
	if len(data) > 0 {
		buf, err := jsoniter.Marshal(data)
		if err != nil {
			return err
		}
		msg.Data = string(buf)
	}

	// 6. 写入数据库
	return h.db.Create(msg).Error
}

// Close 关闭底层数据库连接（通常在程序退出时调用）。
func (h *GormHook) Close() error {
	// 获取底层 *sql.DB
	db, err := h.db.DB()
	if err != nil {
		return err
	}
	// 关闭连接池
	return db.Close()
}
