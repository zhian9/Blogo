// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package mail

import (
	"context"
	"sync"
	"time"

	"gopkg.in/gomail.v2"
)

var (
	// globalSender 是全局唯一的 SMTP 发送器实例
	globalSender *SmtpSender
	// once 确保 SetSender 只能初始化一次（线程安全）
	once sync.Once
)

// SetSender 设置全局 SMTP 发送器（仅首次调用生效）
// 用于在程序启动时配置邮件服务参数
func SetSender(sender *SmtpSender) {
	once.Do(func() {
		globalSender = sender
	})
}

// Send 使用全局 SMTP 发送器发送邮件，支持 To、Cc、Bcc
func Send(ctx context.Context, to []string, cc []string, bcc []string, subject string, body string, file ...string) error {
	return globalSender.Send(ctx, to, cc, bcc, subject, body, file...)
}

// SendTo 是 Send 的简化版，仅指定收件人（To），自动重试 3 次
func SendTo(ctx context.Context, to []string, subject string, body string, file ...string) error {
	return globalSender.SendTo(ctx, to, subject, body, file...)
}

// SendToOne 发送邮件给单个收件人（最常用场景），自动重试
func SendToOne(ctx context.Context, to string, subject string, body string) error {
	return globalSender.SendTo(ctx, []string{to}, subject, body)
}

// SmtpSender 表示一个 SMTP 邮件客户端配置
type SmtpSender struct {
	SmtpHost string // SMTP 服务器地址，例如 "smtp.qq.com"
	Port     int    // SMTP 端口，例如 587 或 465
	FromName string // 发件人显示名称，例如 "系统通知"
	FromMail string // 发件人邮箱地址，例如 "noreply@example.com"
	UserName string // 登录用户名（通常是邮箱地址）
	AuthCode string // 授权码（不是邮箱密码！）
}

// Send 执行实际的邮件发送逻辑
func (s *SmtpSender) Send(ctx context.Context, to []string, cc []string, bcc []string, subject string, body string, file ...string) error {
	// 创建新邮件消息，使用 Base64 编码避免中文乱码
	msg := gomail.NewMessage(gomail.SetEncoding(gomail.Base64))

	// 设置发件人（格式： "显示名 <邮箱>"）
	msg.SetHeader("From", msg.FormatAddress(s.FromMail, s.FromName))

	// 设置收件人（To）、抄送（Cc）、密送（Bcc）
	msg.SetHeader("To", to...)
	if len(cc) > 0 {
		msg.SetHeader("Cc", cc...)
	}
	if len(bcc) > 0 {
		msg.SetHeader("Bcc", bcc...)
	}

	msg.SetHeader("Subject", subject)
	// 设置 HTML 正文，指定 UTF-8 编码以支持中文
	msg.SetBody("text/html;charset=utf-8", body)

	// 添加附件（可选）
	for _, filePath := range file {
		msg.Attach(filePath)
	}

	// 创建 SMTP 拨号器（含认证信息）
	d := gomail.NewDialer(s.SmtpHost, s.Port, s.UserName, s.AuthCode)

	// 实际连接并发送邮件
	return d.DialAndSend(msg)
}

// SendTo 是带重试机制的发送方法（最多重试 3 次，每次间隔 500ms）
// 适用于网络不稳定场景
func (s *SmtpSender) SendTo(ctx context.Context, to []string, subject string, body string, file ...string) error {
	var err error
	for i := 0; i < 3; i++ {
		// 调用完整版 Send（Cc/Bcc 为空）
		err = s.Send(ctx, to, nil, nil, subject, body, file...)
		if err == nil {
			// 发送成功，直接返回
			return nil
		}
		// 发送失败，等待 500ms 后重试
		time.Sleep(500 * time.Millisecond)
	}
	// 3 次都失败，返回最后一次错误
	return err
}
