// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
)

var ErrCasbinDenied = errors.Forbidden("com.casbin.denied", "Permission denied")

type CasbinConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	Skipper             func(c *gin.Context) bool
	GetEnforcer         func(c *gin.Context) *casbin.Enforcer
	GetSubjects         func(c *gin.Context) []string
}

func CasbinWithConfig(config CasbinConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			(config.Skipper != nil && config.Skipper(c)) {
			c.Next()
			return
		}

		enforcer := config.GetEnforcer(c)
		if enforcer == nil {
			c.Next()
			return
		}

		subs := config.GetSubjects(c)
		for _, sub := range subs {
			if b, err := enforcer.Enforce(sub, c.Request.URL.Path, c.Request.Method); err != nil {
				util.ResError(c, err)
				return
			} else if b {
				c.Next()
				return
			}
		}
		logging.Context(c.Request.Context()).Warn("Casbin denied",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Strings("subjects", subs),
			zap.Int("skip_prefixes", len(config.SkippedPathPrefixes)),
		)
		util.ResError(c, ErrCasbinDenied)
	}
}
