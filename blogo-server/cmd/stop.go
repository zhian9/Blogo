// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running server",
	Long:  `Stop the server by reading the PID from .lock file and sending a termination signal`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return stop(cmd, args)
	},
}

func stop(cmd *cobra.Command, args []string) error {
	appName := cmd.Root()
	lockFile := fmt.Sprintf("%s.local", appName)

	//读取 PID 文件
	pidBytes, err := os.ReadFile(lockFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("lock file %s not found .Is the service running", lockFile)
		}
		return fmt.Errorf("Failed to read the lock file %s : %w", lockFile, err)
	}

	//解析 PID （去除空白符）
	pidStr := strings.TrimSpace(string(pidBytes))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid PID in lock file %s: %s", lockFile, pidStr)
	}

	//发送 SIGTERM 信号
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find the process with PID %d: %w", pid, err)
	}

	err = proc.Signal(os.Interrupt)
	if err != nil {
		if !strings.Contains(err.Error(), "process already finished") &&
			!strings.Contains(err.Error(), "no such process") {
			return fmt.Errorf("failed to signal process %d: %w", pid, err)
		}
	}

	// 删除 lock 文件
	if err := os.Remove(lockFile); err != nil {
		return fmt.Errorf("failed to remove lock file %s: %w", lockFile, err)
	}

	fmt.Printf("Service %s stopped successfully (PID: %d)\n", appName, pid)
	return nil
}

// StopCmd 返回停止服务的command
func StopCmd() *cobra.Command {
	return stopCmd
}
