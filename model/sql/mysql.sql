-- MySQL 数据库初始化脚本
-- 创建数据库和表结构

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `slink` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `slink`;

-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `group_id` bigint unsigned NOT NULL DEFAULT 1 COMMENT '用户组ID',
    `name` varchar(191) NOT NULL DEFAULT '集帅' COMMENT '用户名称',
    `email` varchar(191) NOT NULL COMMENT '邮箱地址',
    `password` longtext NOT NULL COMMENT '加密后的密码',
    `is_admin` bigint unsigned NOT NULL DEFAULT 0 COMMENT '是否管理员 0-否 1-是',
    `capacity` bigint unsigned NOT NULL DEFAULT 0 COMMENT '存储容量限制(字节)',
    `configs` longtext NOT NULL COMMENT '用户配置(JSON格式)',
    `image_nums` bigint unsigned NOT NULL DEFAULT 0 COMMENT '图片数量',
    `registered_ip` varchar(191) NOT NULL DEFAULT '' COMMENT '注册IP地址',
    `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表 - 存储系统用户信息';

-- 系统配置表
CREATE TABLE IF NOT EXISTS `configs` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '配置ID',
    `config_key` varchar(191) NOT NULL COMMENT '配置键名',
    `value` longtext NOT NULL COMMENT '配置值',
    `description` varchar(191) DEFAULT NULL COMMENT '配置描述',
    `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_configs_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表 - 存储系统运行配置';

-- 用户组表
CREATE TABLE IF NOT EXISTS `groups` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '组ID',
    `name` varchar(191) NOT NULL COMMENT '组名称',
    `is_default` bigint unsigned NOT NULL DEFAULT 0 COMMENT '是否默认组 0-否 1-是',
    `is_guest` bigint unsigned NOT NULL DEFAULT 0 COMMENT '是否游客组 0-否 1-是',
    `configs` longtext NOT NULL COMMENT '组配置(JSON格式)',
    `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户组表 - 定义用户分组用于权限控制';

-- 个人访问令牌表
CREATE TABLE IF NOT EXISTS `personal_access_tokens` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '令牌ID',
    `api_name` varchar(191) NOT NULL COMMENT 'API名称',
    `username` varchar(191) NOT NULL COMMENT '所属用户名',
    `token` varchar(255) NOT NULL COMMENT '访问令牌',
    `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='个人访问令牌表 - 存储用户API访问令牌';

-- 用户组与存储策略关联表
CREATE TABLE IF NOT EXISTS `group_strategies` (
    `group_id` bigint unsigned NOT NULL COMMENT '用户组ID',
    `strategy_id` bigint unsigned NOT NULL COMMENT '存储策略ID',
    `created_at` bigint DEFAULT NULL COMMENT '创建时间',
    `updated_at` bigint DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`group_id`, `strategy_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户组与存储策略关联表 - 实现多对多关系';

-- 存储策略表
CREATE TABLE IF NOT EXISTS `strategies` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '策略ID',
    `name` varchar(191) NOT NULL COMMENT '策略名称',
    `introduction` varchar(255) NOT NULL COMMENT '策略简介',
    `strategy_key` varchar(191) NOT NULL COMMENT '策略标识符',
    `configs` text NOT NULL COMMENT '策略配置(JSON格式)',
    `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_strategies_name` (`name`),
    UNIQUE INDEX `idx_strategies_key` (`strategy_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='存储策略表 - 定义图片存储后端配置';

-- 图片信息表
CREATE TABLE IF NOT EXISTS `images` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '图片ID',
    `user_id` bigint unsigned NOT NULL COMMENT '上传用户ID',
    `group_id` bigint unsigned NOT NULL COMMENT '所属用户组ID',
    `strategy_id` bigint unsigned NOT NULL COMMENT '使用的存储策略ID',
    `image_key` varchar(191) NOT NULL COMMENT '图片唯一标识符',
    `path` varchar(191) NOT NULL COMMENT '图片存储路径',
    `name` varchar(191) NOT NULL COMMENT '图片文件名',
    `origin_name` varchar(191) NOT NULL COMMENT '原始文件名',
    `size` bigint NOT NULL COMMENT '文件大小(字节)',
    `mimetype` varchar(191) NOT NULL COMMENT 'MIME类型',
    `extension` varchar(191) NOT NULL COMMENT '文件扩展名',
    `md5` varchar(191) NOT NULL COMMENT '文件MD5值',
    `sha1` varchar(191) NOT NULL COMMENT '文件SHA1值',
    `width` bigint NOT NULL COMMENT '图片宽度',
    `height` bigint NOT NULL COMMENT '图片高度',
    `permissions` bigint unsigned NOT NULL DEFAULT 0 COMMENT '访问权限 0-公开 1-私有',
    `is_unhealthy` bigint unsigned NOT NULL DEFAULT 0 COMMENT '是否不健康 0-否 1-是',
    `upload_ip` varchar(191) NOT NULL COMMENT '上传IP地址',
    `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_images_user_id` (`user_id`),
    INDEX `idx_images_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图片信息表 - 存储所有图片元数据';

-- 图片分享记录表
CREATE TABLE IF NOT EXISTS `shares` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '分享ID',
    `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
    `image_id` bigint unsigned NOT NULL COMMENT '图片ID',
    `share_code` varchar(32) NOT NULL COMMENT '分享码',
    `password` varchar(255) DEFAULT NULL COMMENT '分享密码(加密存储)',
    `expires_at` datetime(3) DEFAULT NULL COMMENT '过期时间',
    `view_count` int NOT NULL DEFAULT 0 COMMENT '查看次数',
    `max_views` int NOT NULL DEFAULT 0 COMMENT '最大查看次数 0表示无限制',
    `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否激活',
    `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_shares_code` (`share_code`),
    INDEX `idx_shares_user_id` (`user_id`),
    INDEX `idx_shares_image_id` (`image_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图片分享记录表 - 存储用户分享图片记录';
