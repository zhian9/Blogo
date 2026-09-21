// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/util"
)

// resolveTagIDs 把「标签 ID 或标签名称」统一解析成标签 ID，必要时按名称创建新标签。
//
// 为什么需要它：ArticleForm.TagIDs / ProjectForm.TagIDs 在前端有两种来源
//   - 从标签选择器里勾选的已有标签 —— 传的是标签 ID（如 "d881cmgapfts73b0lu60"）
//   - 在编辑器里手打的新标签     —— 传的是标签名称（如 "golang"）
//
// 旧实现把两者都当作「名称」去查（GetByNames），于是传 ID 时查不到，就新建了一个
// 「名字等于该 ID」的标签，文章/项目被挂到这种垃圾标签上，真正的标签下自然搜不到内容。
//
// 现在的顺序：先按 ID 精确匹配 → 再按名称精确匹配 → 都没有才按名称新建。
// 返回值按顺序去重，避免同一标签产生重复关联。
func resolveTagIDs(ctx context.Context, tagDAL *dal.Tag, values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}

	tagIDs := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	appendID := func(id string) {
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		tagIDs = append(tagIDs, id)
	}

	for _, value := range values {
		if value == "" {
			continue
		}

		tag, err := tagDAL.Get(ctx, value)
		if err != nil {
			return nil, err
		}
		if tag != nil {
			appendID(tag.ID)
			continue
		}

		tag, err = tagDAL.GetByName(ctx, value)
		if err != nil {
			return nil, err
		}
		if tag != nil {
			appendID(tag.ID)
			continue
		}

		// 3. 名称也不存在：按名称新建标签
		newTag := &schema.Tag{
			ID:        util.NewXID(),
			Name:      value,
			CreatedAt: time.Now(),
		}
		if err := tagDAL.Create(ctx, newTag); err != nil {
			return nil, err
		}
		appendID(newTag.ID)
	}

	return tagIDs, nil
}
