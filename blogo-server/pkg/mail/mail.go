// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package mail

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	gomail "github.com/wneessen/go-mail"
)

var (
	globalSender *SmtpSender
	once         sync.Once
)

func SetSender(sender *SmtpSender) {
	once.Do(func() {
		globalSender = sender
	})
}

func Send(ctx context.Context, to []string, cc []string, bcc []string, subject string, body string, file ...string) error {
	return globalSender.Send(ctx, to, cc, bcc, subject, body, file...)
}

func SendTo(ctx context.Context, to []string, subject string, body string, file ...string) error {
	return globalSender.SendTo(ctx, to, subject, body, file...)
}

func SendToOne(ctx context.Context, to string, subject string, body string) error {
	return globalSender.SendTo(ctx, []string{to}, subject, body)
}

type SmtpSender struct {
	SmtpHost string
	Port     int
	FromName string
	FromMail string
	UserName string
	AuthCode string
}

func (s *SmtpSender) Send(ctx context.Context, to []string, cc []string, bcc []string, subject string, body string, file ...string) error {
	msg := gomail.NewMsg()
	if err := msg.FromFormat(s.FromName, s.FromMail); err != nil {
		return err
	}
	msg.To(to...)
	if len(cc) > 0 {
		msg.Cc(cc...)
	}
	if len(bcc) > 0 {
		msg.Bcc(bcc...)
	}
	msg.Subject(subject)
	msg.SetBodyString("text/html", body)
	for _, fp := range file {
		msg.AttachFile(fp)
	}

	client, err := gomail.NewClient(s.SmtpHost,
		gomail.WithPort(s.Port),
		gomail.WithSSL(),
		gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(s.UserName),
		gomail.WithPassword(s.AuthCode),
		gomail.WithTLSConfig(&tls.Config{ServerName: s.SmtpHost}),
	)
	if err != nil {
		return err
	}
	return client.DialAndSend(msg)
}

func (s *SmtpSender) SendTo(ctx context.Context, to []string, subject string, body string, file ...string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = s.Send(ctx, to, nil, nil, subject, body, file...)
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}
