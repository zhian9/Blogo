// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package rbac

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/crypto/hash"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	seedAdminUsername = "admin"
	seedAdminPassword = "Admin123456"
	seedAdminRoleCode = "super_admin"
	seedAdminRoleName = "超级管理员"
	seedUserRoleCode  = "user"
	seedUserRoleName  = "用户"
	seedGuestRoleCode = "guest"
	seedGuestRoleName = "游客"
)

// InitAdminAccount 初始化超级管理员账号 + 全站角色数据清洗（幂等）。
func (r *RBAC) InitAdminAccount(ctx context.Context) error {
	db := r.DB

	// 1. 确保五个角色存在
	allRoles := map[string]string{
		seedAdminRoleCode:   seedAdminRoleName,
		"content_manager":   "内容管理员",
		"comment_moderator": "评论审核员",
		seedUserRoleCode:    seedUserRoleName,
		seedGuestRoleCode:   seedGuestRoleName,
	}
	for code, name := range allRoles {
		if err := ensureRole(db, code, name); err != nil {
			return err
		}
	}

	// 2. 旧角色码迁移：admin → super_admin（避免重复）
	var oldAdminRole schema.Role
	if err := db.Where("code = ?", "admin").First(&oldAdminRole).Error; err == nil {
		var superExists schema.Role
		if db.Where("code = ?", "super_admin").First(&superExists).Error == nil {
			// super_admin 已存在 → 删除旧 admin 角色，将关联用户迁移到 super_admin
			db.Model(&schema.UserRole{}).Where("role_id = ?", oldAdminRole.ID).Update("role_id", superExists.ID)
			db.Delete(&oldAdminRole)
		} else {
			db.Model(&oldAdminRole).Update("code", "super_admin")
		}
		logging.Context(ctx).Info("migrated legacy role code", zap.String("from", "admin"), zap.String("to", "super_admin"))
	}

	// 3. 全站数据清洗：通过邮箱/用户名定位超级管理员
	var adminUser schema.User
	err := db.Where("email = ? OR username = ?", "admin@blogo.local", seedAdminUsername).First(&adminUser).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == nil {
		if err := cleanseAdminUser(ctx, db, &adminUser, allRoles); err != nil {
			return err
		}
	} else {
		// 超管不存在 → 创建
		if err := createAdminUser(ctx, db, allRoles); err != nil {
			return err
		}
	}

	// 4. 全站其他用户强制重置为 user 角色
	if err := cleanseOtherUsers(ctx, db, allRoles, adminUser.ID); err != nil {
		logging.Context(ctx).Error("failed to cleanse other users", zap.Error(err))
	}

	return nil
}

// cleanseAdminUser 强制将超管账号的角色设为 super_admin（数据修复）
func cleanseAdminUser(ctx context.Context, db *gorm.DB, user *schema.User, allRoles map[string]string) error {
	// 确保账号处于激活状态且昵称正确
	if user.Name != "系统管理员" || user.Status != schema.UserStatusActivated {
		db.Model(user).Updates(map[string]interface{}{
			"name": "系统管理员", "status": schema.UserStatusActivated,
		})
	}

	// 查找 super_admin 角色
	superAdminRole, err := getRoleByCode(db, seedAdminRoleCode)
	if err != nil {
		return err
	}

	// 剥离其他所有角色
	db.Where("user_id = ?", user.ID).Delete(&schema.UserRole{})

	// 绑定 super_admin 角色
	if err := ensureUserHasRole(db, user.ID, superAdminRole.ID); err != nil {
		return err
	}

	logging.Context(ctx).Info("super admin data cleansed",
		zap.String("username", user.Username),
		zap.String("user_id", user.ID),
		zap.String("role", seedAdminRoleCode),
	)
	return nil
}

// createAdminUser 创建超级管理员用户
func createAdminUser(ctx context.Context, db *gorm.DB, allRoles map[string]string) error {
	hashedPwd, err := hash.GeneratePassword(seedAdminPassword)
	if err != nil {
		return err
	}

	adminUser := schema.User{
		ID:        util.NewXID(),
		Username:  seedAdminUsername,
		Name:      "系统管理员",
		Password:  hashedPwd,
		Email:     "admin@blogo.local",
		Status:    schema.UserStatusActivated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(&adminUser).Error; err != nil {
		return err
	}

	superAdminRole, err := getRoleByCode(db, seedAdminRoleCode)
	if err != nil {
		return err
	}
	if err := ensureUserHasRole(db, adminUser.ID, superAdminRole.ID); err != nil {
		return err
	}

	logging.Context(ctx).Info("super admin account created",
		zap.String("username", seedAdminUsername),
		zap.String("user_id", adminUser.ID),
	)
	return nil
}

// cleanseOtherUsers 将所有非超管用户强制重置为 user 角色（user_role.id 复用 user.id，20 字符不截断）
func cleanseOtherUsers(ctx context.Context, db *gorm.DB, allRoles map[string]string, adminUserID string) error {
	userRole, err := getRoleByCode(db, seedUserRoleCode)
	if err != nil {
		return err
	}

	// 删除所有非超管用户的现有角色绑定
	db.Exec("DELETE ur FROM user_role ur WHERE ur.user_id != ?", adminUserID)

	// 为所有非超管用户重新绑定 user 角色
	db.Exec(`
		INSERT INTO user_role (id, user_id, role_id, created_at, updated_at)
		SELECT u.id, u.id, ?, NOW(), NOW()
		FROM user u
		WHERE u.id != ? AND u.id NOT IN (
			SELECT ur.user_id FROM user_role ur WHERE ur.role_id = ? AND ur.user_id != ?
		)
	`, userRole.ID, adminUserID, userRole.ID, adminUserID)

	logging.Context(ctx).Info("all other users reset to user role",
		zap.String("role", seedUserRoleCode),
	)
	return nil
}

// ensureRole 确保指定 code 的角色存在
func ensureRole(db *gorm.DB, code, name string) error {
	var existing schema.Role
	err := db.Where("code = ?", code).First(&existing).Error
	if err == nil {
		if existing.Status != schema.RoleStatusEnabled {
			return db.Model(&existing).Update("status", schema.RoleStatusEnabled).Error
		}
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	seq := 0
	switch code {
	case seedAdminRoleCode:
		seq = 1
	case "content_manager":
		seq = 2
	case "comment_moderator":
		seq = 3
	case seedUserRoleCode:
		seq = 4
	case seedGuestRoleCode:
		seq = 5
	}

	role := schema.Role{
		ID:          util.NewXID(),
		Code:        code,
		Name:        name,
		Description: name,
		Sequence:    seq,
		Status:      schema.RoleStatusEnabled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return db.Create(&role).Error
}

func getRoleByCode(db *gorm.DB, code string) (*schema.Role, error) {
	var role schema.Role
	err := db.Where("code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func ensureUserHasRole(db *gorm.DB, userID, roleID string) error {
	var count int64
	if err := db.Model(&schema.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	ur := schema.UserRole{
		ID:        util.NewXID(),
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return db.Create(&ur).Error
}

func GetUserRoleCode(db *gorm.DB, userID, rootID string) string {
	if userID == rootID {
		return seedAdminRoleCode
	}
	var ur schema.UserRole
	err := db.Where("user_id = ?", userID).First(&ur).Error
	if err != nil {
		return seedGuestRoleCode
	}
	var role schema.Role
	err = db.Where("id = ?", ur.RoleID).First(&role).Error
	if err != nil {
		return seedGuestRoleCode
	}
	return role.Code
}

func (r *RBAC) EnsureRegisterRole(ctx context.Context, userID string) error {
	userRole, err := getRoleByCode(r.DB, seedUserRoleCode)
	if err != nil {
		return err
	}
	return ensureUserHasRole(r.DB, userID, userRole.ID)
}
