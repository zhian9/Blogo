// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package bootstrap

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/utility/prom"
	"github.com/zhian9/blogo-server/internal/wirex"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/middleware"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
)

// startHTTPServer 启动 HTTP 服务（Gin 引擎）。
// 返回清理函数（用于优雅关闭）。
func startHTTPServer(ctx context.Context, injector *wirex.Injector) (func(), error) {
	// 1. 设置 Gin 模式（Debug/Release）
	if config.C.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 2. 创建 Gin 引擎
	e := gin.New()

	// 3. 健康检查路由（无中间件）
	e.GET("/health", func(c *gin.Context) {
		util.ResOK(c) // 返回 { "success": true }
	})

	// 4. 全局中间件（必须最先注册）
	// Recovery: 捕获 panic 避免服务崩溃
	e.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
		Skip: config.C.Middleware.Recovery.Skip,
	}))

	// 5. 全局错误处理（404/405）
	e.NoMethod(func(c *gin.Context) {
		util.ResError(c, errors.MethodNotAllowed("", "Method Not Allowed"))
	})
	e.NoRoute(func(c *gin.Context) {
		util.ResError(c, errors.NotFound("", "Not Found"))
	})

	// 6. 获取允许的路由前缀（用于中间件过滤）
	allowedPrefixes := injector.M.RouterPrefixes()

	// 7. 注册业务中间件（按执行顺序）
	if err := useHTTPMiddlewares(ctx, e, injector, allowedPrefixes); err != nil {
		return nil, err
	}

	// 8. 注册业务路由
	if err := injector.M.RegisterRouters(ctx, e); err != nil {
		return nil, err
	}

	// 9. 注册 SEO 端点（搜索引擎）
	e.GET("/sitemap.xml", injector.M.Blog.SEOAPI.Sitemap)
	e.GET("/robots.txt", injector.M.Blog.SEOAPI.Robots)

	// 10. 注册 Swagger 文档（如果启用）
	if !config.C.General.DisableSwagger {
		// 配置 Swagger UI 文档路由
		e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.URL("/docs/swagger.json"),    // 使用生成的swagger.json文件
			ginSwagger.DefaultModelsExpandDepth(-1), // 默认不展开数据模型
			ginSwagger.DocExpansion("list"),         // 默认展开所有API组
			ginSwagger.InstanceName("go-admin"),     // 实例名称
			ginSwagger.PersistAuthorization(true),   // 保持认证信息
		))
		// 提供 swagger.json 文件
		e.Static("/docs", "../docs/api")
	}

	// 10. Prometheus 指标端点（独立于路由组，不受 Auth 限制但受 Basic Auth 保护）
	if config.C.Util.Prometheus.Enable {
		e.GET("/metrics", prometheusHandler())
	}

	// 11. pprof 性能分析端点（dev 环境直接放行，生产需 super_admin）
	pprofGroup := e.Group("/debug/pprof")
	pprofGroup.Use(pprofMiddleware(injector))
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
		pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}

	// 12. 注册静态文件服务（如前端构建产物）
	e.Static("/uploads", "./storage/uploads")

	if dir := config.C.Middleware.Static.Dir; dir != "" {
		e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:                dir,
			SkippedPathPrefixes: allowedPrefixes,
		}))
	}

	// 13. 创建 HTTP 服务器
	addr := config.C.General.HTTP.Addr
	logging.Context(ctx).Info(fmt.Sprintf("HTTP server is listening on %s", addr))
	srv := &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  time.Second * time.Duration(config.C.General.HTTP.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(config.C.General.HTTP.WriteTimeout),
		IdleTimeout:  time.Second * time.Duration(config.C.General.HTTP.IdleTimeout),
	}

	// 12. 启动服务（goroutine）
	go func() {
		var err error
		// 12.1 支持 HTTPS
		if config.C.General.HTTP.CertFile != "" && config.C.General.HTTP.KeyFile != "" {
			srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			err = srv.ListenAndServeTLS(config.C.General.HTTP.CertFile, config.C.General.HTTP.KeyFile)
		} else {
			// 12.2 HTTP
			err = srv.ListenAndServe()
		}

		// 12.3 错误处理（忽略 http.ErrServerClosed）
		if err != nil && err != http.ErrServerClosed {
			logging.Context(ctx).Error("Failed to listen http server", zap.Error(err))
		}
	}()

	// 13. 返回清理函数（优雅关闭）
	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.C.General.HTTP.ShutdownTimeout))
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logging.Context(ctx).Error("Failed to shutdown http server", zap.Error(err))
		}
	}, nil
}

// useHTTPMiddlewares 按正确顺序注册所有业务中间件。
// 中间件执行顺序 = 注册顺序（从上到下）。
func useHTTPMiddlewares(_ context.Context, e *gin.Engine, injector *wirex.Injector, allowedPrefixes []string) error {
	// 1. CORS（跨域资源共享）
	// 注意：必须在最前面（处理 OPTIONS 预检请求）
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Enable:                 config.C.Middleware.CORS.Enable,
		AllowAllOrigins:        config.C.Middleware.CORS.AllowAllOrigins,
		AllowOrigins:           config.C.Middleware.CORS.AllowOrigins,
		AllowMethods:           config.C.Middleware.CORS.AllowMethods,
		AllowHeaders:           config.C.Middleware.CORS.AllowHeaders,
		AllowCredentials:       config.C.Middleware.CORS.AllowCredentials,
		ExposeHeaders:          config.C.Middleware.CORS.ExposeHeaders,
		MaxAge:                 config.C.Middleware.CORS.MaxAge,
		AllowWildcard:          config.C.Middleware.CORS.AllowWildcard,
		AllowBrowserExtensions: config.C.Middleware.CORS.AllowBrowserExtensions,
		AllowWebSockets:        config.C.Middleware.CORS.AllowWebSockets,
		AllowFiles:             config.C.Middleware.CORS.AllowFiles,
	}))

	// 2. Trace（链路追踪）
	// 生成/透传 TraceID，注入上下文
	e.Use(middleware.TraceWithConfig(middleware.TraceConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.Trace.SkippedPathPrefixes,
		RequestHeaderKey:    config.C.Middleware.Trace.RequestHeaderKey,
		ResponseTraceKey:    config.C.Middleware.Trace.ResponseTraceKey,
	}))

	// 3. Logger（请求日志）
	// 记录请求/响应详情（需 CopyBody 提供请求体）
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		AllowedPathPrefixes:      allowedPrefixes,
		SkippedPathPrefixes:      config.C.Middleware.Logger.SkippedPathPrefixes,
		MaxOutputRequestBodyLen:  config.C.Middleware.Logger.MaxOutputRequestBodyLen,
		MaxOutputResponseBodyLen: config.C.Middleware.Logger.MaxOutputResponseBodyLen,
	}))

	// 4. CopyBody（请求体复制）
	// 使请求体可多次读取（供 Logger/Auth 使用）
	e.Use(middleware.CopyBodyWithConfig(middleware.CopyBodyConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.CopyBody.SkippedPathPrefixes,
		MaxContentLen:       config.C.Middleware.CopyBody.MaxContentLen,
	}))

	// 5. OptionalAuth（可选认证：Token 存在且有效时注入用户身份，无效则静默放行）
	// 不做路径跳过——凡带 Token 的请求都应识别身份，尤其是被 Auth 中间件跳过前缀的受保护操作（如点赞）
	e.Use(middleware.OptionalAuthWithConfig(middleware.OptionalAuthConfig{
		AllowedPathPrefixes: allowedPrefixes,
		Auth:                injector.Auth,
		RootID:              config.C.General.Root.ID,
	}))

	// 6. UserStatusCheck（全局封禁即时拦截）
	// 对所有 /api/ 路径检测已登录用户的实时状态；仅豁免登录/注册/验证码等入口
	e.Use(middleware.UserStatusCheckWithConfig(middleware.UserStatusCheckConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: []string{
			"/api/v1/captcha",
			"/api/v1/login",
			"/api/v1/register",
			"/api/v1/verify-email",
		},
		CheckStatus: injector.M.RBAC.CheckStatus,
	}))

	// 6. Auth（认证）
	// 解析用户身份，注入上下文（已移至路由组级别）

	// 7. RateLimiter（限流）
	// 限制 IP/用户请求频率
	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Enable:              config.C.Middleware.RateLimiter.Enable,
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.RateLimiter.SkippedPathPrefixes,
		Period:              config.C.Middleware.RateLimiter.Period,
		MaxRequestsPerIP:    config.C.Middleware.RateLimiter.MaxRequestsPerIP,
		MaxRequestsPerUser:  config.C.Middleware.RateLimiter.MaxRequestsPerUser,
		StoreType:           config.C.Middleware.RateLimiter.Store.Type,
		MemoryStoreConfig: middleware.RateLimiterMemoryConfig{
			Expiration:      time.Second * time.Duration(config.C.Middleware.RateLimiter.Store.Memory.Expiration),
			CleanupInterval: time.Second * time.Duration(config.C.Middleware.RateLimiter.Store.Memory.CleanupInterval),
		},
		RedisStoreConfig: middleware.RateLimiterRedisConfig{
			Addr:     config.C.Middleware.RateLimiter.Store.Redis.Addr,
			Password: config.C.Middleware.RateLimiter.Store.Redis.Password,
			DB:       config.C.Middleware.RateLimiter.Store.Redis.DB,
			Username: config.C.Middleware.RateLimiter.Store.Redis.Username,
		},
	}))

	// 8. Prometheus（监控指标）
	// 自动收集 HTTP 请求指标
	if config.C.Util.Prometheus.Enable {
		e.Use(prom.GinMiddleware)
	}

	return nil
}

// prometheusHandler 返回 Prometheus 指标端点（强制 Basic Auth）
func prometheusHandler() gin.HandlerFunc {
	handler := promhttp.Handler()
	return func(c *gin.Context) {
		user, pass, ok := c.Request.BasicAuth()
		if !ok || user != config.C.Util.Prometheus.BasicUsername ||
			pass != config.C.Util.Prometheus.BasicPassword {
			c.Header("WWW-Authenticate", `Basic realm="metrics"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// pprofMiddleware 保护 pprof 端点（生产环境仅 super_admin 可访问）
func pprofMiddleware(injector *wirex.Injector) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.C.IsDebug() {
			c.Next()
			return
		}

		token := util.GetToken(c)
		if token == "" {
			util.ResError(c, errors.Forbidden("", "pprof 仅超级管理员可访问"))
			c.Abort()
			return
		}

		userID, err := injector.Auth.ParseSubject(c.Request.Context(), token)
		if err != nil || userID == "" {
			util.ResError(c, errors.Forbidden("", "pprof 仅超级管理员可访问"))
			c.Abort()
			return
		}

		if userID == config.C.General.Root.ID {
			c.Next()
			return
		}

		var count int64
		injector.M.RBAC.DB.Table("user_role AS ur").
			Joins("JOIN role r ON r.id = ur.role_id").
			Where("ur.user_id = ? AND r.code IN ?", userID, []string{"super_admin", "admin"}).
			Count(&count)
		if count > 0 {
			c.Next()
			return
		}

		util.ResError(c, errors.Forbidden("", "pprof 仅超级管理员可访问"))
		c.Abort()
	}
}
