-- /*
--  * SPDX-License-Identifier: MIT
--  *
--  * Copyright (c) 2026-present 李星云 (lxy911)
--  *
--  * Project: Blogo
--  * Repository: https://github.com/zhian9/Blogo
--  */

-- =============================================================
-- 博客文章阅读权限系统 — 数据库迁移
-- =============================================================
-- 说明：
--   1. 本项目的表名格式为 {table_prefix}article，前缀通过 .env / 配置文件注入。
--      以下 SQL 使用占位符 `{prefix}` 表示前缀。
--      例如默认无前缀：{prefix}article → article
--   2. GORM AutoMigrate 会在服务启动时自动创建 article_visible_user 表，
--      本迁移文件主要供手动升级或审计使用。
--   3. article 表的 visibility 字段已存在（VARCHAR(20), default 'public'），
--      本次仅需确认其支持 'partial_visible' 值——无需修改表结构。
-- =============================================================

-- -------------------------------------------------------------
-- 1. 创建文章可见用户关联表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `{prefix}article_visible_user` (
    `id`         VARCHAR(20)  NOT NULL COMMENT '主键ID (XID)',
    `article_id` VARCHAR(20)  NOT NULL COMMENT '文章ID',
    `user_id`    VARCHAR(20)  NOT NULL COMMENT '用户ID',
    `created_at` DATETIME     NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_article_user` (`article_id`, `user_id`),
    INDEX `idx_article_id` (`article_id`),
    INDEX `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章可见用户关联表';

-- -------------------------------------------------------------
-- 2. 补充：article 表 visibility 字段已存在，无需 ALTER TABLE。
--    确认默认值和允许值即可：
--      visibility VARCHAR(20) NOT NULL DEFAULT 'public'
--    有效值：'public' | 'private' | 'partial_visible'
-- -------------------------------------------------------------
-- 如需手动迁移，可执行：
-- ALTER TABLE `{prefix}article`
--     MODIFY COLUMN `visibility` VARCHAR(20) NOT NULL DEFAULT 'public' COMMENT '可见性：public/private/partial_visible';
