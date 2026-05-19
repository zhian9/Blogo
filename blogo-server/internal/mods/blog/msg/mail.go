// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package msg

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"sync"

	gomail "github.com/wneessen/go-mail"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
)

// ─── Mail types ───

type MailType string

const (
	MailTypeWelcome     MailType = "welcome"
	MailTypeNewArticle  MailType = "new_article"
	MailTypeUnsubscribe MailType = "unsubscribe"
)

// ─── Message ───

type MailMessage struct {
	Type         MailType
	ArticleTitle string
	ArticleSlug  string
	SiteURL      string
	SubEmail     string
}

// ─── Worker ───

type MailWorker struct {
	ch     chan MailMessage
	wg     sync.WaitGroup
	cancel context.CancelFunc
	DB     *dal.Subscriber
}

func NewMailWorker(subDAL *dal.Subscriber) *MailWorker {
	return &MailWorker{ch: make(chan MailMessage, 100), DB: subDAL}
}

func (w *MailWorker) Start(ctx context.Context) {
	ctx, w.cancel = context.WithCancel(ctx)
	for i := 0; i < 2; i++ {
		w.wg.Add(1)
		go w.worker(ctx, i)
	}
	log.Printf("[MailWorker] 2 workers started")
}

func (w *MailWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	close(w.ch)
	w.wg.Wait()
	log.Println("[MailWorker] stopped")
}

func (w *MailWorker) Send(msg MailMessage) {
	select {
	case w.ch <- msg:
	default:
		log.Println("[MailWorker] channel full, dropping message")
	}
}

func (w *MailWorker) worker(ctx context.Context, id int) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-w.ch:
			if !ok {
				return
			}
			w.process(ctx, msg, id)
		}
	}
}

func (w *MailWorker) process(ctx context.Context, msg MailMessage, workerID int) {
	log.Printf("[MailWorker-%d] processing message type=%s", workerID, msg.Type)
	cfg := config.C.Email
	if cfg.Host == "" || cfg.Password == "" {
		log.Printf("[MailWorker-%d] email not configured (host=%s, pwd=%t)", workerID, cfg.Host, cfg.Password != "")
		return
	}

	switch msg.Type {
	case MailTypeWelcome:
		if msg.SubEmail == "" {
			return
		}
		if err := w.sendWelcome(cfg, msg.SubEmail, msg.SiteURL); err != nil {
			log.Printf("[MailWorker-%d] welcome to %s failed: %v", workerID, msg.SubEmail, err)
		} else {
			log.Printf("[MailWorker-%d] welcome sent to %s", workerID, msg.SubEmail)
		}

	case MailTypeUnsubscribe:
		if msg.SubEmail == "" {
			return
		}
		if err := w.sendUnsubscribeGoodbye(cfg, msg.SubEmail, msg.SiteURL); err != nil {
			log.Printf("[MailWorker-%d] goodbye to %s failed: %v", workerID, msg.SubEmail, err)
		} else {
			log.Printf("[MailWorker-%d] goodbye sent to %s", workerID, msg.SubEmail)
		}

	case MailTypeNewArticle:
		subs, err := w.DB.GetAllActive(ctx)
		if err != nil {
			log.Printf("[MailWorker-%d] get subscribers failed: %v", workerID, err)
			return
		}
		log.Printf("[MailWorker-%d] sending article '%s' to %d subscribers", workerID, msg.ArticleTitle, len(subs))
		articleURL := fmt.Sprintf("%s/article/%s", msg.SiteURL, msg.ArticleSlug)
		for _, sub := range subs {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err := w.sendArticle(cfg, sub.Email, msg.ArticleTitle, articleURL, msg.SiteURL); err != nil {
				log.Printf("[MailWorker-%d] to %s failed: %v", workerID, sub.Email, err)
			}
		}
	}
}

// ─── Template data ───

type emailData struct {
	SiteURL      string
	Recipient    string
	UnsubURL     string
	ArticleTitle string
	ArticleURL   string
	ReSubURL     string
}

func newEmailData(siteURL, to string) emailData {
	return emailData{
		SiteURL:   siteURL,
		Recipient: to,
		UnsubURL:  siteURL + "/unsubscribe?email=" + url.QueryEscape(to),
		ReSubURL:  siteURL,
	}
}

// ─── Shared layout ───

const baseStyles = `margin:0;padding:0;background-color:#0B1311;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif`

const wrapperStart = `<table width="100%" cellpadding="0" cellspacing="0" style="` + baseStyles + `"><tr><td align="center" style="padding:40px 16px">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:560px;background-color:#141D1A;border-radius:16px;border:1px solid rgba(255,255,255,0.06);overflow:hidden">`

const headerBlock = `<tr><td style="padding:32px 32px 0;text-align:center">
<div style="font-family:Georgia,'Times New Roman',serif;font-style:italic;font-size:32px;font-weight:400;color:#FFFFFF;letter-spacing:-0.02em;margin-bottom:4px">Blogo</div>
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:11px;font-weight:500;color:rgba(255,255,255,0.3);letter-spacing:0.16em;text-transform:uppercase">Technical Blog</div>
</td></tr>`

const divider = `<tr><td style="padding:0 32px"><div style="height:1px;background:rgba(255,255,255,0.06);margin:24px 0"></div></td></tr>`

const wrapperEnd = `<tr><td style="padding:32px;text-align:center">
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:11px;color:rgba(255,255,255,0.22);line-height:1.6">
此邮件发送至 {{.Recipient}}<br>
由 <span style="font-family:Georgia,serif;font-style:italic">Blogo</span> 自动发送 ·
<a href="{{.UnsubURL}}" style="color:rgba(255,255,255,0.3);text-decoration:underline">退订</a>
</div>
</td></tr></table></td></tr></table>`

// ─── Welcome template ───

const welcomeBody = `{{template "header" .}}
<tr><td style="padding:28px 32px 8px;text-align:center">
<div style="font-family:Georgia,serif;font-size:20px;font-weight:400;color:#FFFFFF;letter-spacing:-0.01em">欢迎订阅 ✦</div>
</td></tr>
<tr><td style="padding:8px 32px 28px;text-align:center">
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:14px;color:rgba(255,255,255,0.55);line-height:1.8;max-width:440px;margin:0 auto">
感谢订阅 <span style="font-family:Georgia,serif;font-style:italic;color:#FFFFFF">Blogo</span> 博客。<br>
我们将为你精选<strong style="color:rgba(255,255,255,0.75)">后端开发、系统设计与开源</strong>相关的深度思考。
</div>
</td></tr>
<tr><td style="padding:0 32px 32px;text-align:center">
<div style="display:inline-block;background:rgba(255,255,255,0.04);border:1px solid rgba(255,255,255,0.08);border-radius:12px;padding:16px 24px;max-width:400px">
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:12px;color:rgba(255,255,255,0.3);letter-spacing:0.08em;text-transform:uppercase;margin-bottom:6px">每有新文章发布时</div>
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:13px;color:rgba(255,255,255,0.5);line-height:1.6">你会第一时间收到邮件提醒，直达邮箱。</div>
</div>
</td></tr>
{{template "divider" .}}
{{template "footer" .}}`

// ─── New article template ───

const articleBody = `{{template "header" .}}
<tr><td style="padding:24px 32px 4px;text-align:center">
<div style="font-family:Georgia,serif;font-size:16px;font-weight:400;color:rgba(255,255,255,0.45);letter-spacing:-0.01em">新文章发布</div>
</td></tr>
<tr><td style="padding:8px 32px 4px;text-align:center">
<div style="font-family:Georgia,serif;font-size:22px;font-weight:400;color:#FFFFFF;line-height:1.35;letter-spacing:-0.02em;max-width:440px;margin:0 auto">{{.ArticleTitle}}</div>
</td></tr>
<tr><td style="padding:24px 32px 8px;text-align:center">
<a href="{{.ArticleURL}}" style="display:inline-block;padding:14px 36px;background:linear-gradient(135deg,rgba(79,110,247,0.9),rgba(99,102,241,0.85));color:#FFFFFF;font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:15px;font-weight:600;text-decoration:none;border-radius:12px;box-shadow:0 4px 20px rgba(79,110,247,0.3);letter-spacing:0.03em" target="_blank">阅读全文</a>
</td></tr>
<tr><td style="padding:20px 32px 0;text-align:center">
<div style="display:inline-block;background:rgba(79,110,247,0.06);border:1px solid rgba(79,110,247,0.12);border-radius:10px;padding:10px 20px">
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:11px;color:rgba(255,255,255,0.3);letter-spacing:0.06em">
这是你订阅 <span style="font-family:Georgia,serif;font-style:italic;color:rgba(255,255,255,0.5)">Blogo</span> 后收到的动态通知
</div>
</div>
</td></tr>
{{template "divider" .}}
{{template "footer" .}}`

// ─── Unsubscribe template ───

const goodbyeBody = `{{template "header" .}}
<tr><td style="padding:28px 32px 8px;text-align:center">
<div style="font-family:Georgia,serif;font-size:20px;font-weight:400;color:rgba(255,255,255,0.7);letter-spacing:-0.01em">已成功退订</div>
</td></tr>
<tr><td style="padding:8px 32px 28px;text-align:center">
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:14px;color:rgba(255,255,255,0.45);line-height:1.8;max-width:420px;margin:0 auto">
很遗憾看到你离开。<br>
你将不再收到 <span style="font-family:Georgia,serif;font-style:italic;color:rgba(255,255,255,0.6)">Blogo</span> 的新文章邮件通知。
</div>
</td></tr>
<tr><td style="padding:0 32px 28px;text-align:center">
<div style="display:inline-block;background:rgba(255,255,255,0.03);border:1px solid rgba(255,255,255,0.06);border-radius:12px;padding:14px 24px">
<div style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:12px;color:rgba(255,255,255,0.3);margin-bottom:8px">如果这是误操作</div>
<a href="{{.ReSubURL}}" style="font-family:-apple-system,Helvetica,Arial,sans-serif;font-size:13px;color:rgba(79,110,247,0.8);text-decoration:underline">随时可以重新订阅</a>
</div>
</td></tr>
{{template "divider" .}}
{{template "footer" .}}`

// ─── Template registry ───

var emailTemplates = template.Must(template.New("email").Parse(`
{{define "header"}}` + headerBlock + `{{end}}
{{define "divider"}}` + divider + `{{end}}
{{define "footer"}}` + wrapperEnd + `{{end}}
{{define "welcome"}}` + welcomeBody + `{{end}}
{{define "article"}}` + articleBody + `{{end}}
{{define "goodbye"}}` + goodbyeBody + `{{end}}
`))

// ─── Render helper ───

func renderTemplate(name string, data emailData) string {
	var buf bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("[MailWorker] template render error: %v", err)
		return ""
	}
	return wrapperStart + buf.String()
}

// ─── Send methods ───

func (w *MailWorker) sendWelcome(cfg config.EmailConfig, to, siteURL string) error {
	data := newEmailData(siteURL, to)
	m := gomail.NewMsg()
	m.From(cfg.FromEmail)
	m.To(to)
	m.Subject("欢迎订阅 Blogo！")
	m.SetBodyString("text/html", renderTemplate("welcome", data))
	return w.dialAndSend(cfg, m)
}

func (w *MailWorker) sendArticle(cfg config.EmailConfig, to, title, articleURL, siteURL string) error {
	data := newEmailData(siteURL, to)
	data.ArticleTitle = title
	data.ArticleURL = articleURL
	m := gomail.NewMsg()
	m.From(cfg.FromEmail)
	m.To(to)
	m.Subject(fmt.Sprintf("新文章 · %s", title))
	m.SetBodyString("text/html", renderTemplate("article", data))
	return w.dialAndSend(cfg, m)
}

func (w *MailWorker) sendUnsubscribeGoodbye(cfg config.EmailConfig, to, siteURL string) error {
	data := newEmailData(siteURL, to)
	m := gomail.NewMsg()
	m.From(cfg.FromEmail)
	m.To(to)
	m.Subject("已退订 Blogo 邮件通知")
	m.SetBodyString("text/html", renderTemplate("goodbye", data))
	return w.dialAndSend(cfg, m)
}

func (w *MailWorker) dialAndSend(cfg config.EmailConfig, m *gomail.Msg) error {
	client, err := gomail.NewClient(cfg.Host,
		gomail.WithPort(cfg.Port), gomail.WithSSL(),
		gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(cfg.FromEmail), gomail.WithPassword(cfg.Password),
		gomail.WithTLSConfig(&tls.Config{ServerName: cfg.Host}),
	)
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}
	return client.DialAndSend(m)
}
