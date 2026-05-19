// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/ossx"
	"github.com/zhian9/blogo-server/pkg/util"
)

// Image 是图片管理业务的核心对象。
type Image struct {
	Trans    *util.Trans // 事务管理器
	ImageDAL *dal.Image  // 图片数据访问层
}

// Query 查询图片列表（分页 + 分类筛选）。
func (i *Image) Query(ctx context.Context, params schema.ImageQueryParam) (*schema.ImageQueryResult, error) {
	params.Pagination = true

	result, err := i.ImageDAL.Query(ctx, params, schema.ImageQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "created_at", Direction: util.DESC},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get 获取单张图片。
func (i *Image) Get(ctx context.Context, id string) (*schema.Image, error) {
	image, err := i.ImageDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if image == nil {
		return nil, errors.NotFound("", "Image not found")
	}
	return image, nil
}

// Create 创建新图片记录（通常由上传服务调用）。
func (i *Image) Create(ctx context.Context, imageForm *schema.ImageForm) (*schema.Image, error) {
	if err := imageForm.Validate(); err != nil {
		return nil, err
	}

	image := &schema.Image{
		ID:        util.NewXID(),
		CreatedAt: time.Now(),
	}
	imageForm.FillTo(image)

	err := i.Trans.Exec(ctx, func(ctx context.Context) error {
		return i.ImageDAL.Create(ctx, image)
	})
	if err != nil {
		return nil, err
	}
	return i.Get(ctx, image.ID)
}

// Update 更新图片信息。
func (i *Image) Update(ctx context.Context, id string, imageForm *schema.ImageForm) error {
	image, err := i.ImageDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if image == nil {
		return errors.NotFound("", "Image not found")
	}

	if err := imageForm.Validate(); err != nil {
		return err
	}

	imageForm.FillTo(image)
	image.UpdatedAt = time.Now()

	return i.Trans.Exec(ctx, func(ctx context.Context) error {
		return i.ImageDAL.Update(ctx, image)
	})
}

// Delete 删除图片。
func (i *Image) Delete(ctx context.Context, id string) error {
	exists, err := i.ImageDAL.ExistsID(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Image not found")
	}

	return i.Trans.Exec(ctx, func(ctx context.Context) error {
		return i.ImageDAL.Delete(ctx, id)
	})
}

// DeleteByIds 批量删除图片。
func (i *Image) DeleteByIds(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return i.Trans.Exec(ctx, func(ctx context.Context) error {
		return i.ImageDAL.DeleteByIds(ctx, ids)
	})
}

// Upload 处理图片文件上传。
// R2 启用时上传到 Cloudflare R2，否则写入本地磁盘。
func (i *Image) Upload(ctx context.Context, fileHeader *multipart.FileHeader, category string) (*schema.Image, error) {
	const maxSize int64 = 5 * 1024 * 1024
	if fileHeader.Size > maxSize {
		return nil, errors.BadRequest(config.ErrImageTooLarge, "Image size exceeds 5MB limit")
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if !util.IsImageMimeType(mimeType) {
		return nil, errors.BadRequest(config.ErrImageInvalidType, "Unsupported image type: %s", mimeType)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, errors.BadRequest(config.ErrImageUploadFailed, "Failed to open uploaded file")
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == ".jpeg" {
		ext = ".jpg"
	}
	now := time.Now()
	imageID := util.NewXID()
	objectKey := fmt.Sprintf("uploads/%s/%s/%s/%s%s",
		category, now.Format("2006"), now.Format("01"), imageID, ext)

	var fileURL string
	var filePath string

	if ossx.Enabled() {
		// ---- R2 上传 ----
		fileURL = ossx.PublicURL(objectKey)
		filePath = objectKey
		if _, err := ossx.Upload(ctx, objectKey, src, mimeType); err != nil {
			return nil, errors.InternalServerError(config.ErrImageUploadFailed, "R2 upload failed: %s", err.Error())
		}
	} else {
		// ---- 本地磁盘上传 ----
			workDir, _ := os.Getwd()
			urlPart := filepath.Join(category, now.Format("2006"), now.Format("01"), imageID+ext)
			relDir := filepath.Join("storage", "uploads", category, now.Format("2006"), now.Format("01"))
			relPath := filepath.Join(relDir, imageID+ext)
		absDir := filepath.Join(workDir, relDir)
		absPath := filepath.Join(workDir, relPath)

		if err := os.MkdirAll(absDir, 0755); err != nil {
			return nil, errors.BadRequest(config.ErrImageUploadFailed, "Failed to create upload directory")
		}

		dst, err := os.Create(absPath)
		if err != nil {
			return nil, errors.BadRequest(config.ErrImageUploadFailed, "Failed to create file on disk")
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return nil, errors.BadRequest(config.ErrImageUploadFailed, "Failed to write file")
		}

		fileURL = "/uploads/" + filepath.ToSlash(urlPart)
		filePath = relPath
	}

	image := &schema.Image{
		ID:        imageID,
		URL:       fileURL,
		Path:      filePath,
		Name:      fileHeader.Filename,
		Size:      fileHeader.Size,
		Type:      mimeType,
		Category:  category,
		CreatedAt: now,
	}

	if err := i.Trans.Exec(ctx, func(ctx context.Context) error {
		return i.ImageDAL.Create(ctx, image)
	}); err != nil {
		if ossx.Enabled() {
			_ = ossx.Delete(ctx, objectKey)
		}
		return nil, errors.InternalServerError(config.ErrImageUploadFailed, "Failed to save image record: %s", err.Error())
	}

	return image, nil
}

// GetByCategory 获取某分类下的所有图片（不分页）。
func (i *Image) GetByCategory(ctx context.Context, category string) (schema.Images, error) {
	return i.ImageDAL.GetByCategory(ctx, category)
}
