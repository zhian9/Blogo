// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package rbac

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/rbac/dal"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/cachex"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
)

// Casbinx 是 Casbin 权限系统的封装，负责策略加载与重载。
type Casbinx struct {
	// enforcer: 原子存储 Casbin Enforcer 实例（线程安全读取）
	enforcer *atomic.Value `wirex:"-"` // 不由 Wire 注入

	// ticker: 定时器，用于自动检查策略更新
	ticker *time.Ticker `wirex:"-"`

	// Cache: 缓存客户端（用于接收策略更新信号）
	Cache cachex.Cacher

	// 数据访问层（由 Wire 注入）
	MenuDAL         *dal.Menu
	MenuResourceDAL *dal.MenuResource
	RoleDAL         *dal.Role
}

// GetEnforcer 安全获取当前 Casbin Enforcer 实例。
// 返回 nil 表示尚未加载或已禁用。
func (a *Casbinx) GetEnforcer() *casbin.Enforcer {
	if v := a.enforcer.Load(); v != nil {
		return v.(*casbin.Enforcer)
	}
	return nil
}

// policyQueueItem 表示一个角色的权限策略项。
type policyQueueItem struct {
	RoleID    string               // 角色 ID
	Resources schema.MenuResources // 该角色可访问的资源列表
}

// Load 初始化 Casbin 系统。
// 如果配置禁用，则直接返回。
func (a *Casbinx) Load(ctx context.Context) error {
	if config.C.Middleware.Casbin.Disable {
		logging.Context(ctx).Info("Casbin is disabled by config, skipping policy load")
		return nil
	}

	// 初始化原子存储
	a.enforcer = new(atomic.Value)

	// 首次加载策略
	if err := a.load(ctx); err != nil {
		return err
	}

	// 启动后台自动重载协程
	go a.autoLoad(ctx)
	return nil
}

// load 从数据库加载所有角色的权限策略，生成 Casbin 策略文件。
func (a *Casbinx) load(ctx context.Context) error {
	start := time.Now()

	// 1. 查询所有启用的角色（仅需 ID）
	roleResult, err := a.RoleDAL.Query(ctx, schema.RoleQueryParam{
		Status: schema.RoleStatusEnabled,
	}, schema.RoleQueryOptions{
		QueryOptions: util.QueryOptions{SelectFields: []string{"id"}},
	})
	if err != nil {
		return err
	} else if len(roleResult.Data) == 0 {
		return nil // 无角色，无需加载
	}

	// 2. 并发准备：通道 + WaitGroup + 互斥锁
	var resCount int32                                         // 资源总数（原子计数）
	queue := make(chan *policyQueueItem, len(roleResult.Data)) // 角色策略队列
	threadNum := config.C.Middleware.Casbin.LoadThread         // 并发线程数
	lock := new(sync.Mutex)                                    // 写缓冲区互斥锁
	buf := new(bytes.Buffer)                                   // 最终策略缓冲区

	// 3. 启动工作协程（并发生成策略字符串）
	wg := new(sync.WaitGroup)
	wg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go func() {
			defer wg.Done()
			ibuf := new(bytes.Buffer) // 每个协程独立缓冲区，避免锁竞争
			for item := range queue {
				// 生成 Casbin 策略行：p, 角色ID, 路径, 方法
				for _, res := range item.Resources {
					_, _ = ibuf.WriteString(fmt.Sprintf("p, %s, %s, %s\n", item.RoleID, res.Path, res.Method))
				}
			}
			// 合并到主缓冲区（加锁）
			lock.Lock()
			_, _ = buf.Write(ibuf.Bytes())
			lock.Unlock()
		}()
	}

	// 4. 查询每个角色的资源权限，并发送到队列
	for _, item := range roleResult.Data {
		resources, err := a.queryRoleResources(ctx, item.ID)
		if err != nil {
			logging.Context(ctx).Error("Failed to query role resources", zap.Error(err))
			continue
		}
		atomic.AddInt32(&resCount, int32(len(resources)))
		queue <- &policyQueueItem{
			RoleID:    item.ID,
			Resources: resources,
		}
	}
	close(queue) // 关闭队列，通知工作协程退出
	wg.Wait()    // 等待所有协程完成

	// 4.5 注入 admin 角色通配符规则：p, adminRoleID, *, *（超级管理员无上特权）
	var adminRole schema.Role
	if err := a.RoleDAL.DB.Where("code IN ? AND status = ?", []string{"super_admin", "admin"}, schema.RoleStatusEnabled).First(&adminRole).Error; err == nil {
		lock.Lock()
		_, _ = buf.WriteString(fmt.Sprintf("p, %s, *, *\n", adminRole.ID))
		lock.Unlock()
	}

	// 5. 写入策略文件并初始化 Enforcer
	if buf.Len() > 0 {
		policyFile := filepath.Join(config.C.General.WorkDir, config.C.Middleware.Casbin.GenPolicyFile)

		// 移除旧备份（Windows 下 Rename 目标存在会失败）
		_ = os.Remove(policyFile + ".bak")
		// 备份旧文件
		_ = os.Rename(policyFile, policyFile+".bak")

		// 确保目录存在
		_ = os.MkdirAll(filepath.Dir(policyFile), 0755)

		// 移除只读属性（上次运行可能设置了 0444）
		_ = os.Chmod(policyFile, 0666)

		// 写入新策略文件
		if err := os.WriteFile(policyFile, buf.Bytes(), 0666); err != nil {
			logging.Context(ctx).Error("Failed to write policy file", zap.Error(err))
			return err
		}
		// 设置只读（防止意外修改）
		_ = os.Chmod(policyFile, 0444)

		// 6. 创建 Casbin Enforcer
		modelFile := filepath.Join(config.C.General.WorkDir, config.C.Middleware.Casbin.ModelFile)
		e, err := casbin.NewEnforcer(modelFile, policyFile)
		if err != nil {
			logging.Context(ctx).Error("Failed to create casbin enforcer", zap.Error(err))
			return err
		}
		e.EnableLog(config.C.IsDebug()) // 调试模式下打印 Casbin 日志
		a.enforcer.Store(e)             // 原子更新 Enforcer
	}

	// 7. 记录加载日志
	logging.Context(ctx).Info("Casbin load policy",
		zap.Duration("cost", time.Since(start)),
		zap.Int("roles", len(roleResult.Data)),
		zap.Int32("resources", resCount),
		zap.Int("bytes", buf.Len()),
	)
	return nil
}

// queryRoleResources 查询角色可访问的所有资源（含菜单树继承）。
// 规则：角色拥有菜单 A，则自动拥有 A 及其所有父菜单的资源权限。
func (a *Casbinx) queryRoleResources(ctx context.Context, roleID string) (schema.MenuResources, error) {
	// 1. 查询角色关联的菜单（含 parent_path）
	menuResult, err := a.MenuDAL.Query(ctx, schema.MenuQueryParam{
		RoleID: roleID,
		Status: schema.MenuStatusEnabled,
	}, schema.MenuQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"id", "parent_id", "parent_path"},
		},
	})
	if err != nil {
		return nil, err
	} else if len(menuResult.Data) == 0 {
		return nil, nil
	}

	// 2. 收集所有菜单 ID（去重 + 包含父级）
	menuIDs := make([]string, 0, len(menuResult.Data))
	menuIDMapper := make(map[string]struct{}) // 去重
	for _, item := range menuResult.Data {
		if _, ok := menuIDMapper[item.ID]; ok {
			continue
		}
		menuIDs = append(menuIDs, item.ID)
		menuIDMapper[item.ID] = struct{}{}

		// 解析 parent_path（如 "root.parent"）
		if pp := item.ParentPath; pp != "" {
			for _, pid := range strings.Split(pp, util.TreePathDelimiter) {
				if pid == "" {
					continue
				}
				if _, ok := menuIDMapper[pid]; ok {
					continue
				}
				menuIDs = append(menuIDs, pid)
				menuIDMapper[pid] = struct{}{}
			}
		}
	}

	// 3. 查询这些菜单关联的所有资源（API 路径+方法）
	menuResourceResult, err := a.MenuResourceDAL.Query(ctx, schema.MenuResourceQueryParam{
		MenuIDs: menuIDs,
	})
	if err != nil {
		return nil, err
	}

	return menuResourceResult.Data, nil
}

// autoLoad 定期检查缓存中的更新信号，触发策略重载。
// 信号机制：
//   - 当菜单/角色/资源变更时，业务逻辑会更新缓存键 config.CacheKeyForSyncToCasbin
//   - 值为时间戳（Unix 纳秒），表示最后更新时间
func (a *Casbinx) autoLoad(ctx context.Context) {
	var lastUpdated int64 // 上次加载的时间戳

	// 创建定时器（间隔由配置决定）
	a.ticker = time.NewTicker(time.Duration(config.C.Middleware.Casbin.AutoLoadInterval) * time.Second)

	for range a.ticker.C {
		// 1. 从缓存读取更新信号
		val, ok, err := a.Cache.Get(ctx, config.CacheNSForRole, config.CacheKeyForSyncToCasbin)
		if err != nil {
			logging.Context(ctx).Error("Failed to get cache", zap.Error(err), zap.String("key", config.CacheKeyForSyncToCasbin))
			continue
		} else if !ok {
			continue // 无更新信号
		}

		// 2. 解析时间戳
		updated, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			logging.Context(ctx).Error("Failed to parse cache value", zap.Error(err), zap.String("val", val))
			continue
		}

		// 3. 如果有新更新，则重载策略
		if lastUpdated < updated {
			if err := a.load(ctx); err != nil {
				logging.Context(ctx).Error("Failed to load casbin policy", zap.Error(err))
			} else {
				lastUpdated = updated // 更新最后加载时间
			}
		}
	}
}

// TriggerReload 触发 Casbin 策略重载（写入缓存信号，由 autoLoad 检测）
func (a *Casbinx) TriggerReload(ctx context.Context) {
	if a.Cache == nil {
		return
	}
	_ = a.Cache.Set(ctx, config.CacheNSForRole, config.CacheKeyForSyncToCasbin,
		fmt.Sprintf("%d", time.Now().UnixNano()), 10*time.Second)
}

// Release 释放 Casbinx 资源（停止定时器）。
func (a *Casbinx) Release(ctx context.Context) error {
	if a.ticker != nil {
		a.ticker.Stop()
	}
	return nil
}
