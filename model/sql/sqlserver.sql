-- SQLServer 数据库初始化脚本
-- 创建数据库和表结构

-- 创建数据库
IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = 'slink')
BEGIN
    CREATE DATABASE [slink]
END
GO

USE [slink]
GO

-- 用户表
IF OBJECT_ID('users', 'U') IS NOT NULL DROP TABLE [users]
GO

CREATE TABLE [users] (
    [id] bigint IDENTITY(1,1) PRIMARY KEY,
    [group_id] bigint NOT NULL DEFAULT 1,
    [name] nvarchar(191) NOT NULL DEFAULT N'集帅',
    [email] nvarchar(191) NOT NULL,
    [password] nvarchar(max) NOT NULL,
    [is_admin] bigint NOT NULL DEFAULT 0,
    [capacity] bigint NOT NULL DEFAULT 0,
    [configs] nvarchar(max) NOT NULL,
    [image_nums] bigint NOT NULL DEFAULT 0,
    [registered_ip] nvarchar(191) NOT NULL DEFAULT '',
    [created_at] datetime2(3) NULL,
    [updated_at] datetime2(3) NULL
)
GO

CREATE UNIQUE INDEX [idx_users_email] ON [users] ([email])
GO

-- 系统配置表
IF OBJECT_ID('configs', 'U') IS NOT NULL DROP TABLE [configs]
GO

CREATE TABLE [configs] (
    [id] bigint IDENTITY(1,1) PRIMARY KEY,
    [config_key] nvarchar(191) NOT NULL,
    [value] nvarchar(max) NOT NULL,
    [description] nvarchar(191) NULL,
    [created_at] datetime2(3) NULL,
    [updated_at] datetime2(3) NULL
)
GO

CREATE UNIQUE INDEX [idx_configs_key] ON [configs] ([config_key])
GO

-- 用户组表
IF OBJECT_ID('groups', 'U') IS NOT NULL DROP TABLE [groups]
GO

CREATE TABLE [groups] (
    [id] bigint IDENTITY(1,1) PRIMARY KEY,
    [name] nvarchar(191) NOT NULL,
    [is_default] bigint NOT NULL DEFAULT 0,
    [is_guest] bigint NOT NULL DEFAULT 0,
    [configs] nvarchar(max) NOT NULL,
    [created_at] datetime2(3) NULL,
    [updated_at] datetime2(3) NULL
)
GO

-- 个人访问令牌表
IF OBJECT_ID('personal_access_tokens', 'U') IS NOT NULL DROP TABLE [personal_access_tokens]
GO

CREATE TABLE [personal_access_tokens] (
    [id] bigint IDENTITY(1,1) PRIMARY KEY,
    [api_name] nvarchar(191) NOT NULL,
    [username] nvarchar(191) NOT NULL,
    [token] nvarchar(255) NOT NULL,
    [created_at] datetime2(3) NULL,
    [updated_at] datetime2(3) NULL
)
GO

-- 用户组与存储策略关联表
IF OBJECT_ID('group_strategies', 'U') IS NOT NULL DROP TABLE [group_strategies]
GO

CREATE TABLE [group_strategies] (
    [group_id] bigint NOT NULL,
    [strategy_id] bigint NOT NULL,
    [created_at] bigint NULL,
    [updated_at] bigint NULL,
    PRIMARY KEY ([group_id], [strategy_id])
)
GO

-- 存储策略表
IF OBJECT_ID('strategies', 'U') IS NOT NULL DROP TABLE [strategies]
GO

CREATE TABLE [strategies] (
    [id] bigint IDENTITY(1,1) PRIMARY KEY,
    [name] nvarchar(191) NOT NULL,
    [introduction] nvarchar(255) NOT NULL,
    [strategy_key] nvarchar(191) NOT NULL,
    [configs] nvarchar(max) NOT NULL,
    [created_at] datetime2(3) NULL,
    [updated_at] datetime2(3) NULL
)
GO

CREATE UNIQUE INDEX [idx_strategies_name] ON [strategies] ([name])
GO
CREATE UNIQUE INDEX [idx_strategies_key] ON [strategies] ([strategy_key])
GO

-- 图片信息表
IF OBJECT_ID('images', 'U') IS NOT NULL DROP TABLE [images]
GO

CREATE TABLE [images] (
    [id] bigint IDENTITY(1,1) PRIMARY KEY,
    [user_id] bigint NOT NULL,
    [group_id] bigint NOT NULL,
    [strategy_id] bigint NOT NULL,
    [image_key] nvarchar(191) NOT NULL,
    [path] nvarchar(191) NOT NULL,
    [name] nvarchar(191) NOT NULL,
    [origin_name] nvarchar(191) NOT NULL,
    [size] bigint NOT NULL,
    [mimetype] nvarchar(191) NOT NULL,
    [extension] nvarchar(191) NOT NULL,
    [md5] nvarchar(191) NOT NULL,
    [sha1] nvarchar(191) NOT NULL,
    [width] bigint NOT NULL,
    [height] bigint NOT NULL,
    [permissions] bigint NOT NULL DEFAULT 0,
    [is_unhealthy] bigint NOT NULL DEFAULT 0,
    [upload_ip] nvarchar(191) NOT NULL,
    [created_at] datetime2(3) NULL,
    [updated_at] datetime2(3) NULL
)
GO

CREATE INDEX [idx_images_user_id] ON [images] ([user_id])
GO
CREATE INDEX [idx_images_group_id] ON [images] ([group_id])
GO

-- 图片分享记录表
IF OBJECT_ID('shares', 'U') IS NOT NULL DROP TABLE [shares]
GO

CREATE TABLE [shares] (
    [id] bigint IDENTITY(1,1) PRIMARY KEY,
    [user_id] bigint NOT NULL,
    [image_id] bigint NOT NULL,
    [share_code] nvarchar(32) NOT NULL,
    [password] nvarchar(255) NULL,
    [expires_at] datetime2(3) NULL,
    [view_count] int NOT NULL DEFAULT 0,
    [max_views] int NOT NULL DEFAULT 0,
    [is_active] bit NOT NULL DEFAULT 1,
    [created_at] datetime2(3) NULL,
    [updated_at] datetime2(3) NULL
)
GO

CREATE UNIQUE INDEX [idx_shares_code] ON [shares] ([share_code])
GO
CREATE INDEX [idx_shares_user_id] ON [shares] ([user_id])
GO
CREATE INDEX [idx_shares_image_id] ON [shares] ([image_id])
GO
