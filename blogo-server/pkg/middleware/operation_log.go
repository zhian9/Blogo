// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"context"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
)

// UserIDParser 从请求中解析用户ID的函数类型
type UserIDParser func(c *gin.Context) (string, error)

// OperationLogConfig 操作日志配置
type OperationLogConfig struct {
	Enabled             bool               // 是否启用
	SkippedPaths        []string           // 跳过的路径
	AsyncWrite          bool               // 是否异步写入（推荐 true）
	OperationLogBIZFunc func() interface{} // 返回 OperationLog BIZ 的函数
	ParseUserID         UserIDParser       // 从 JWT token 解析用户ID（Auth 跳过时作为 fallback）
}

// DefaultOperationLogConfig 默认配置
var DefaultOperationLogConfig = OperationLogConfig{
	Enabled:    true,
	AsyncWrite: true,
	SkippedPaths: []string{
		"/api/v1/health",
		"/api/v1/login",
		"/api/v1/register",
		"/api/v1/captcha",
		"/api/v1/current/refresh-token",
	},
}

// OperationLog 返回使用默认配置的操作日志中间件
func OperationLog() gin.HandlerFunc {
	return OperationLogWithConfig(DefaultOperationLogConfig)
}

// OperationLogWithConfig 根据自定义配置创建操作日志中间件
func OperationLogWithConfig(cfg OperationLogConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	if cfg.OperationLogBIZFunc == nil {
		logging.Context(nil).Warn("OperationLogBIZFunc is not set, operation log middleware is disabled")
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// 检查是否应该跳过此路径
		if shouldSkipPath(c.Request.URL.Path, cfg.SkippedPaths) {
			c.Next()
			return
		}

		// 只记录 POST/PUT/DELETE 等修改类操作
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
			c.Next()
			return
		}

		// 获取用户ID（由Auth中间件设置到context中）
		userID := util.FromUserID(c.Request.Context())
		if userID == "" {
			// Auth 可能跳过了此路径（如 /api/v1/articles），尝试直接解析 token
			if cfg.ParseUserID != nil {
				if id, err := cfg.ParseUserID(c); err == nil {
					userID = id
				}
			}
		}
		if userID == "" {
			// 未登录，跳过记录
			c.Next()
			return
		}

		// 获取客户端IP（处理代理情况）
		clientIP := c.ClientIP()

		// 解析操作信息
		module, actionType, description := parseOperationInfo(c.Request.URL.Path, method, c)

		// 在响应后处理日志
		c.Next()

		// 判断操作是否成功（2xx 标记为成功）
		statusCode := c.Writer.Status()
		isSuccess := statusCode >= 200 && statusCode < 300

		// 获取错误信息
		var errorMsg string
		if !isSuccess {
			errorMsg = c.GetString("error_message")
			if errorMsg == "" {
				errorMsg = ""
			}
		}

		// 构建操作日志参数
		logParam := &schema.OperationLogCreateParam{
			OperatorID:    userID,
			Operator:      userID, // 暂时使用ID，后续可通过JOIN获取username
			OperatorIP:    clientIP,
			Module:        module,
			ActionType:    actionType,
			Description:   description,
			ResourceID:    extractResourceID(c.Request.URL.Path),
			ResourceName:  extractResourceName(c, method),
			RequestPath:   c.Request.URL.Path,
			RequestMethod: method,
			UserAgent:     c.Request.UserAgent(),
			Status:        isSuccess,
			StatusCode:    statusCode,
			ErrorMsg:      errorMsg,
		}

		// 异步或同步写入数据库
		if cfg.AsyncWrite {
			// 异步写入，避免阻塞响应
			go func() {
				biz := cfg.OperationLogBIZFunc()
				olBizInterface := biz.(interface {
					Create(ctx context.Context, item *schema.OperationLog) error
				})
				operLog := logParam.ToOperationLog()
				if err := olBizInterface.Create(context.Background(), operLog); err != nil {
					logging.Context(context.Background()).Error(
						"failed to save operation log",
						zap.Error(err),
						zap.String("operator", logParam.Operator),
						zap.String("module", logParam.Module),
					)
				}
			}()
		} else {
			// 同步写入
			biz := cfg.OperationLogBIZFunc()
			olBizInterface := biz.(interface {
				Create(ctx context.Context, item *schema.OperationLog) error
			})
			operLog := logParam.ToOperationLog()
			if err := olBizInterface.Create(context.Background(), operLog); err != nil {
				logging.Context(context.Background()).Error(
					"failed to save operation log",
					zap.Error(err),
					zap.String("operator", logParam.Operator),
					zap.String("module", logParam.Module),
				)
			}
		}
	}
}

// shouldSkipPath 检查路径是否应该跳过
func shouldSkipPath(path string, skippedPaths []string) bool {
	for _, skipped := range skippedPaths {
		if strings.Contains(path, skipped) {
			return true
		}
	}
	return false
}

// parseOperationInfo 根据路由路径和HTTP方法解析操作信息
// 返回：(模块, 操作类型, 操作描述)
func parseOperationInfo(path, method string, c *gin.Context) (module, actionType, description string) {
	// 移除 /api/v1 前缀
	path = strings.TrimPrefix(path, "/api/v1/")
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")

	if len(parts) == 0 {
		return "unknown", "unknown", ""
	}

	// 第一个部分通常是资源名称
	resource := parts[0]

	// 根据HTTP方法判断操作类型
	switch method {
	case "POST":
		actionType = "新增"
	case "PUT":
		actionType = "编辑"
	case "PATCH":
		actionType = "更新"
	case "DELETE":
		actionType = "删除"
	default:
		actionType = "操作"
	}

	// 映射模块名称
	module = mapModuleName(resource)

	// 生成描述
	description = generateDescription(c, method, resource, actionType)

	return module, actionType, description
}

// mapModuleName 将资源URL映射到模块名称
func mapModuleName(resource string) string {
	// 处理常见复数形式：categories → category, tags → tag, articles → article
	singular := resource
	switch {
	case strings.HasSuffix(resource, "ies"):
		singular = strings.TrimSuffix(resource, "ies") + "y"
	case strings.HasSuffix(resource, "s"):
		singular = strings.TrimSuffix(resource, "s")
	}

	moduleMap := map[string]string{
		"user":          "用户管理",
		"role":          "角色管理",
		"menu":          "菜单管理",
		"article":       "文章管理",
		"category":      "分类管理",
		"tag":           "标签管理",
		"comment":       "评论管理",
		"page":          "页面管理",
		"image":         "媒体管理",
		"setting":       "系统设置",
		"operation-log": "操作日志",
		"friend-link":   "友链管理",
	}

	if name, ok := moduleMap[singular]; ok {
		return name
	}
	return resource
}

// generateDescription 根据请求内容生成操作描述
func generateDescription(c *gin.Context, method, resource, actionType string) string {
	moduleName := mapModuleName(resource)

	// 尝试从请求体中提取资源名称
	var resourceName string
	switch method {
	case "POST", "PUT", "PATCH":
		// 尝试从表单中提取 name/title/username 等字段
		if title, exist := c.GetPostForm("title"); exist && title != "" {
			resourceName = title
		} else if name, exist := c.GetPostForm("name"); exist && name != "" {
			resourceName = name
		} else if username, exist := c.GetPostForm("username"); exist && username != "" {
			resourceName = username
		}
	case "DELETE":
		// 从URL路径中提取ID
		path := c.Request.URL.Path
		if parts := strings.Split(path, "/"); len(parts) > 1 {
			resourceName = parts[len(parts)-1]
		}
	}

	// 构建描述字符串
	if resourceName != "" {
		return actionType + "了" + moduleName + "「" + resourceName + "」"
	}
	return actionType + moduleName
}

// extractResourceID 从URL路径中提取资源ID
func extractResourceID(path string) string {
	// 移除 /api/v1 前缀
	path = strings.TrimPrefix(path, "/api/v1/")
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")

	if len(parts) >= 2 {
		// 检查第二部分是否看起来像ID（UUID或数字）
		id := parts[1]
		if isValidID(id) {
			return id
		}
	}

	return ""
}

// extractResourceName 从请求中提取资源名称
func extractResourceName(c *gin.Context, method string) string {
	switch method {
	case "POST", "PUT", "PATCH":
		// 尝试从表单数据中获取
		if title, exist := c.GetPostForm("title"); exist && title != "" {
			return title
		}
		if name, exist := c.GetPostForm("name"); exist && name != "" {
			return name
		}
		if username, exist := c.GetPostForm("username"); exist && username != "" {
			return username
		}
	}
	return ""
}

// isValidID 检查是否为有效的ID格式
func isValidID(id string) bool {
	// 检查是否为UUID或XID格式（通常20个字符）
	// 或者数字ID
	if len(id) > 3 && len(id) < 50 {
		// 简单检查：不包含特殊字符（除了 - 和 _）
		match, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, id)
		return match
	}
	return false
}
