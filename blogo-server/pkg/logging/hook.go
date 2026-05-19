// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package logging

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// HookExecute 日志写入器的执行接口
type HookExecute interface {
	Exec(extra map[string]string, b []byte) error //执行日志写入操作
	Close() error                                 //关闭底层资源
}

// hookOptions Hook 内部配置
type hookOptions struct {
	maxJobs    int               //日志队列最大容量
	maxWorkers int               //工作者协程数量
	extra      map[string]string //传给 Execute的 额外参数
}

// HookOptions 是Hook的配置选项函数类型
type HookOptions func(*hookOptions)

func SetHookMaxJobs(maxJobs int) HookOptions {
	return func(o *hookOptions) {
		o.maxJobs = maxJobs
	}
}

func SetHookMaxWorkers(maxWorkers int) HookOptions {
	return func(o *hookOptions) {
		o.maxWorkers = maxWorkers
	}
}

func SetHookExtra(extra map[string]string) HookOptions {
	return func(options *hookOptions) {
		options.extra = extra
	}
}

// Hook 异步日志写入器
type Hook struct {
	opts   *hookOptions    //配置选项
	q      chan []byte     //日志队列
	wg     *sync.WaitGroup //等待所有的 worker 退出
	e      HookExecute     // 具体的日志执行器
	closed int32           // 原子标志位，标记是否已关闭 （1=已关闭）
}

func NewHook(exec HookExecute, opt ...HookOptions) *Hook {
	//设置默认配置
	opts := &hookOptions{
		maxJobs:    1024, // 默认缓冲 1024条日志
		maxWorkers: 2,    //默认 2 个写入协程
	}

	//应用用户传入的选项
	for _, o := range opt {
		o(opts)
	}

	// 初始化 WaitGroup
	wg := new(sync.WaitGroup)
	wg.Add(opts.maxWorkers)

	// 创建 Hook 实例
	h := &Hook{
		opts: opts,
		q:    make(chan []byte, opts.maxJobs), //带缓冲的 channel
		wg:   wg,
		e:    exec,
	}

	// 启动工作
	h.dispatch()
	return h
}

// dispatch 启动多个工作协程，从队列中消费日志并写入。
func (h *Hook) dispatch() {
	for i := 0; i < h.opts.maxWorkers; i++ {
		go func() {
			// 捕获 panic，防止 worker 崩溃导致日志丢失
			defer func() {
				h.wg.Done()
				if r := recover(); r != nil {
					fmt.Println("Recovered from panic in logger hook:", r)
				}
			}()

			// 持续从队列读取日志
			for data := range h.q {
				// 调用具体执行器写入日志
				err := h.e.Exec(h.opts.extra, data)
				if err != nil {
					//注意：此处仅打印错误，生产环境建议接入监控告警
					fmt.Println("Failed to write log entry:", err.Error())
				}
			}
		}()
	}
}

// Write 实现 io.Writer 接口，供 zap 调用。
// zap 会将格式化后的 JSON 日志通过此方法写入。
func (h *Hook) Write(p []byte) (int, error) {
	// 如果已关闭，直接返回（不再处理新日志）
	if atomic.LoadInt32(&h.closed) == 1 {
		return len(p), nil
	}

	// 检查队列是否已满（非阻塞）
	if len(h.q) == h.opts.maxJobs {
		// 队列满时静默丢弃日志！
		// 生产环境可考虑：记录丢弃计数、触发告警、或阻塞等待
		fmt.Println("Too many jobs, waiting for queue to be empty, discard")
		return len(p), nil
	}

	// 深拷贝日志数据（防止 p 被复用导致数据错乱）
	data := make([]byte, len(p))
	copy(data, p)

	// 发送到队列（非阻塞，因为有缓冲）
	h.q <- data

	return len(p), nil
}

// Flush 等待所有日志处理完毕，并关闭 Hook。
// 通常在程序退出时调用，确保日志不丢失。
func (h *Hook) Flush() {
	// 原子检查是否已关闭
	if atomic.LoadInt32(&h.closed) == 1 {
		return
	}

	// 标记为已关闭
	atomic.StoreInt32(&h.closed, 1)

	// 关闭队列 channel，通知所有 worker 退出
	close(h.q)

	// 等待所有 worker 处理完剩余日志
	h.wg.Wait()

	// 关闭底层资源（如数据库连接）
	err := h.e.Close()
	if err != nil {
		fmt.Println("Failed to close logger hook:", err.Error())
	}
}
