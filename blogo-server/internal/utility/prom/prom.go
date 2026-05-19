// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package prom

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhian9/blogo-server/internal/config"
)

var (
	reqCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	reqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	reqInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed.",
		},
	)
)

// GinMiddleware 是 Gin 的 Prometheus 中间件。
// 自动收集 HTTP 请求指标（请求量、延迟、状态码）。
var GinMiddleware gin.HandlerFunc

// Init 根据配置初始化 Prometheus 监控系统。
func Init() {
	// 注册默认指标收集器（CPU、内存、Goroutine 等）
	if config.C.Util.Prometheus.DefaultCollect {
		prometheus.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		prometheus.MustRegister(prometheus.NewGoCollector())
	}

	prometheus.MustRegister(reqCount, reqDuration, reqInFlight)

	GinMiddleware = func(c *gin.Context) {
		start := time.Now()
		reqInFlight.Inc()

		c.Next()

		status := http.StatusText(c.Writer.Status())
		reqInFlight.Dec()
		reqCount.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		reqDuration.WithLabelValues(c.Request.Method, c.FullPath(), status).Observe(time.Since(start).Seconds())
	}
}
