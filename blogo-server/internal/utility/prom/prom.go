// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package prom

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhian9/blogo-server/internal/config"
)

var (
	once sync.Once

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

var GinMiddleware gin.HandlerFunc

func Init() {
	once.Do(func() {
		// 注册默认收集器，已注册则跳过（不 panic）
		if config.C.Util.Prometheus.DefaultCollect {
			tryRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
			tryRegister(prometheus.NewGoCollector())
		}

		tryRegister(reqCount)
		tryRegister(reqDuration)
		tryRegister(reqInFlight)

		GinMiddleware = func(c *gin.Context) {
			start := time.Now()
			reqInFlight.Inc()

			c.Next()

			status := http.StatusText(c.Writer.Status())
			reqInFlight.Dec()
			reqCount.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
			reqDuration.WithLabelValues(c.Request.Method, c.FullPath(), status).Observe(time.Since(start).Seconds())
		}
	})
}

func tryRegister(c prometheus.Collector) {
	err := prometheus.Register(c)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			panic(err)
		}
	}
}
