// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package api

import (
	"fmt"
	"html"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/mods/blog/biz"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
)

// SEO 聚合 SEO 相关接口
type SEO struct {
	ArticleBIZ *biz.Article
}

// ArticleMeta 返回文章页的 SSR HTML（供搜索引擎爬虫使用）
func (s *SEO) ArticleMeta(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.String(404, "Not Found")
		return
	}
	article, err := s.ArticleBIZ.GetBySlug(c.Request.Context(), slug)
	if err != nil || article == nil {
		c.String(404, "Not Found")
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, buildArticleHTML(article))
}

// Sitemap 生成 sitemap.xml
func (s *SEO) Sitemap(c *gin.Context) {
	result, err := s.ArticleBIZ.Query(c.Request.Context(), schema.ArticleQueryParam{})
	if err != nil {
		c.String(500, "Internal Server Error")
		return
	}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, a := range result.Data {
		loc := fmt.Sprintf("%s/article/%s", siteURL(c), a.Slug)
		b.WriteString("  <url>\n")
		b.WriteString(fmt.Sprintf("    <loc>%s</loc>\n", html.EscapeString(loc)))
		b.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", a.UpdatedAt.Format("2006-01-02")))
		b.WriteString("    <changefreq>monthly</changefreq>\n")
		b.WriteString("  </url>\n")
	}
	b.WriteString("</urlset>")
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(200, b.String())
}

// Robots 返回 robots.txt
func (s *SEO) Robots(c *gin.Context) {
	base := siteURL(c)
	body := fmt.Sprintf(`User-agent: *
Allow: /
Disallow: /api/
Disallow: /admin/

Sitemap: %s/sitemap.xml
`, base)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(200, body)
}

// siteURL 从请求推断站点根 URL
func siteURL(c *gin.Context) string {
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	host := c.Request.Host
	if host == "" {
		host = "localhost"
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}

// buildArticleHTML 为爬虫构建包含完整 meta 标签的 HTML 页面
func buildArticleHTML(a *schema.Article) string {
	title := a.Title
	if a.SeoTitle != "" {
		title = a.SeoTitle
	}
	desc := a.Summary
	if a.SeoDesc != "" {
		desc = a.SeoDesc
	}
	keywords := a.SeoKeywords
	canonical := "/article/" + a.Slug
	ogImage := ""
	if a.CoverImage != nil {
		ogImage = a.CoverImage.URL
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s - Blogo</title>
<meta name="description" content="%s">
<meta name="keywords" content="%s">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:type" content="article">
<meta property="og:url" content="%s">
<meta property="og:image" content="%s">
<meta property="og:site_name" content="Blogo">
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="%s">
<meta name="twitter:description" content="%s">
<meta name="twitter:image" content="%s">
<link rel="canonical" href="%s">
</head>
<body>
<article>
<h1>%s</h1>
<p>%s</p>
</article>
</body>
</html>`,
		html.EscapeString(title),
		html.EscapeString(desc),
		html.EscapeString(keywords),
		html.EscapeString(title),
		html.EscapeString(desc),
		canonical,
		ogImage,
		html.EscapeString(title),
		html.EscapeString(desc),
		ogImage,
		canonical,
		html.EscapeString(a.Title),
		html.EscapeString(desc),
	)
}
