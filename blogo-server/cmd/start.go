// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhian9/blogo-server/internal/bootstrap"
	"github.com/zhian9/blogo-server/internal/config"
)

var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start server",
	Long:    `Start the Blogo HTTP server with optional daemon mode`,
	Aliases: []string{"s"}, //命令别名(如 ./app s)
	RunE: func(cmd *cobra.Command, args []string) error {
		return start(cmd, args)
	},
}

// start 项目启动命令，真正的程序入口
func start(cmd *cobra.Command, args []string) error {
	// 从标志中读取参数
	workDir, _ := cmd.Flags().GetString("workdir")
	staticDir, _ := cmd.Flags().GetString("static")
	configs, _ := cmd.Flags().GetString("config")
	daemon, _ := cmd.Flags().GetBool("daemon")
	config.MustLoad(workDir, strings.Split(configs, ",")...)

	//守护进程模式（daemon = true）
	if daemon {
		//获取当前可执行文件的绝对路径
		bin, err := filepath.Abs(os.Args[0])
		if err != nil {
			return fmt.Errorf("failed to get absolute path for command: %w", err)
		}

		// 构建子进程命令参数
		var args []string
		args = append(args, "start")
		if workDir != "" {
			args = append(args, "--workdir", workDir)
		}
		if configs != "" {
			args = append(args, "--config", configs)
		}
		if staticDir != "" {
			args = append(args, "--static", staticDir)
		}
		// 注意：不再传递 --daemon，避免无限 fork

		// 打印即将执行的命令（调试用）
		fmt.Printf("execute command: %s %s\n", bin, strings.Join(args, " "))
		//创建子进程
		command := exec.Command(bin, args...)

		// 重定向 子进程的日志到日志文件
		stdLogFile := fmt.Sprintf("%s.log", cmd.Root().Name())
		file, err := os.OpenFile(stdLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Printf("Falied to open the log file %s : %w", stdLogFile, err)
			}
		}(file)

		command.Stdout = file
		command.Stderr = file

		// 启动子进程
		err = command.Start()
		if err != nil {
			return fmt.Errorf("failed to start daemon process: %w", err)
		}

		// 保存子进程 PID 到 .lock 文件
		pid := command.Process.Pid
		pidFile := fmt.Sprintf("%s.lock", cmd.Root().Name())
		if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0666); err != nil {
			fmt.Printf("warning: failed to write pid file %s: %v\n", pidFile, err)
		}

		fmt.Printf("Service %s daemon started successfully with PID %d\n", config.C.General.AppName, pid)
		os.Exit(0) // 守护进程启动后，父进程退出
	}
	//
	err := bootstrap.Run(context.Background(), bootstrap.RunConfig{
		WorkDir:   workDir,
		Configs:   configs,
		StaticDir: staticDir,
	})
	if err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}
	return nil
}

// init registers flags for the start command.
func init() {
	// 定义标志（flags）
	startCmd.Flags().StringP("workdir", "d", "configs", "Working directory")
	startCmd.Flags().StringP("config", "c", "dev", "Runtime configuration files or directory (relative to workdir, multiple separated by commas)")
	startCmd.Flags().StringP("static", "s", "", "Static files directory")
	startCmd.Flags().Bool("daemon", false, "Run as a daemon")
}

// StartCmd returns the start cobra command.
// 供 root command 添加使用
func StartCmd() *cobra.Command {
	return startCmd
}
