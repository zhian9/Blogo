// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/util"
)

// ProjectPermission 项目阅读权限服务
type ProjectPermission struct {
	ProjectVisibleUserDAL *dal.ProjectVisibleUser
}

// CanUserReadProject 判断指定用户是否有权阅读指定项目
func (p *ProjectPermission) CanUserReadProject(ctx context.Context, userID string, project *schema.Project) bool {
	if project == nil {
		return false
	}

	switch project.Visibility {
	case schema.ProjectVisibilityPublic:
		return true

	case schema.ProjectVisibilityPrivate:
		return p.isAuthorOrAdmin(ctx, userID, project)

	case schema.ProjectVisibilityPartialVisible:
		return p.canReadPartialVisible(ctx, userID, project)

	default:
		return p.isAuthorOrAdmin(ctx, userID, project)
	}
}

func (p *ProjectPermission) isAuthorOrAdmin(ctx context.Context, userID string, project *schema.Project) bool {
	if userID == "" {
		return false
	}
	if project.AuthorID == userID {
		return true
	}
	if util.FromIsRootUser(ctx) {
		return true
	}
	return false
}

func (p *ProjectPermission) canReadPartialVisible(ctx context.Context, userID string, project *schema.Project) bool {
	if userID == "" {
		return false
	}
	if p.isAuthorOrAdmin(ctx, userID, project) {
		return true
	}
	ok, _ := p.ProjectVisibleUserDAL.IsUserVisible(ctx, project.ID, userID)
	return ok
}
