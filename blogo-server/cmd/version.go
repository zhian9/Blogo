// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build-time variables injected via ldflags:
//
//	go build -ldflags "-X 'github.com/zhian9/blogo-server/cmd.gitHash=...' \
//	                    -X 'github.com/zhian9/blogo-server/cmd.buildTime=...'"
var (
	gitHash   = "unknown"
	buildTime = "unknown"
)

// VersionCmd 返回显示版本信息的 cobra 命令。
func VersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示 Blogo 版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Blogo %s\n", version)
			if gitHash != "unknown" {
				fmt.Printf("  commit:  %s\n", gitHash)
			}
			if buildTime != "unknown" {
				fmt.Printf("  built:   %s\n", buildTime)
			}
		},
	}
}
