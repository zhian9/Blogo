// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhian9/blogo-server/cmd"
)

// Build-time variables injected via ldflags:
//
//	go build -ldflags "-X main.VERSION=$(git describe --tags --always) \
//	                    -X main.GIT_HASH=$(git rev-parse --short HEAD) \
//	                    -X main.BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
var (
	VERSION    = "v1.0.0"
	GIT_HASH   = "unknown"
	BUILD_TIME = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "Blogo",
	Short:   "Blogo — a lightweight blog & RBAC admin scaffold powered by Gin + GORM + Casbin + Wire",
	Long:    `Blogo 是一个轻量级博客与后台管理系统，基于 Gin / GORM 2.0 / Casbin 2.0 / Google Wire 构建，提供 RBAC 权限管理、文章 CMS、评论审核、操作审计等能力。`,
	Version: VERSION,
}

func init() {
	rootCmd.AddCommand(cmd.StartCmd())
	rootCmd.AddCommand(cmd.StopCmd())
	rootCmd.AddCommand(cmd.VersionCmd(VERSION))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// @title           Blogo API
// @version         v1.0.0
// @description     Modern SaaS Blog Platform — REST API for content management, user system, RBAC, analytics, and more.
// @termsOfService  https://github.com/zhian9/Blogo
//
// @contact.name   李星云 (lxy911)
// @contact.email  lxyaa911@gmail.com
// @contact.url    https://github.com/zhian9
//
// @license.name   MIT
// @license.url    https://github.com/zhian9/Blogo/blob/master/LICENSE
//
// @host           localhost:8040
// @BasePath       /
// @schemes        http https
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter "Bearer <your JWT token>". Login via /api/v1/login to obtain a token.
func main() {
	Execute()
}
