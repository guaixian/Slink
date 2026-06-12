-- PostgreSQL 数据库初始化脚本
-- 创建数据库和表结构

-- 注意：数据库创建需要手动执行或通过程序处理
-- 执行前请先创建数据库: CREATE DATABASE slink;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL DEFAULT 1,
    name VARCHAR(191) NOT NULL DEFAULT '集帅',
    email VARCHAR(191) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    is_admin BIGINT NOT NULL DEFAULT 0,
    capacity BIGINT NOT NULL DEFAULT 0,
    configs TEXT NOT NULL,
    image_nums BIGINT NOT NULL DEFAULT 0,
    registered_ip VARCHAR(191) NOT NULL DEFAULT '',
    created_at TIMESTAMP(3) DEFAULT NULL,
    updated_at TIMESTAMP(3) DEFAULT NULL
);
COMMENT ON TABLE users IS '用户表 - 存储系统用户信息';
COMMENT ON COLUMN users.id IS '用户ID';
COMMENT ON COLUMN users.group_id IS '用户组ID';
COMMENT ON COLUMN users.name IS '用户名称';
COMMENT ON COLUMN users.email IS '邮箱地址';
COMMENT ON COLUMN users.password IS '加密后的密码';
COMMENT ON COLUMN users.is_admin IS '是否管理员 0-否 1-是';
COMMENT ON COLUMN users.capacity IS '存储容量限制(字节)';
COMMENT ON COLUMN users.configs IS '用户配置(JSON格式)';
COMMENT ON COLUMN users.image_nums IS '图片数量';
COMMENT ON COLUMN users.registered_ip IS '注册IP地址';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '更新时间';

-- 系统配置表
CREATE TABLE IF NOT EXISTS configs (
    id BIGSERIAL PRIMARY KEY,
    config_key VARCHAR(191) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    description VARCHAR(191),
    created_at TIMESTAMP(3) DEFAULT NULL,
    updated_at TIMESTAMP(3) DEFAULT NULL
);
COMMENT ON TABLE configs IS '系统配置表 - 存储系统运行配置';
COMMENT ON COLUMN configs.id IS '配置ID';
COMMENT ON COLUMN configs.config_key IS '配置键名';
COMMENT ON COLUMN configs.value IS '配置值';
COMMENT ON COLUMN configs.description IS '配置描述';
COMMENT ON COLUMN configs.created_at IS '创建时间';
COMMENT ON COLUMN configs.updated_at IS '更新时间';

-- 用户组表
CREATE TABLE IF NOT EXISTS groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(191) NOT NULL,
    is_default BIGINT NOT NULL DEFAULT 0,
    is_guest BIGINT NOT NULL DEFAULT 0,
    configs TEXT NOT NULL,
    created_at TIMESTAMP(3) DEFAULT NULL,
    updated_at TIMESTAMP(3) DEFAULT NULL
);
COMMENT ON TABLE groups IS '用户组表 - 定义用户分组用于权限控制';
COMMENT ON COLUMN groups.id IS '组ID';
COMMENT ON COLUMN groups.name IS '组名称';
COMMENT ON COLUMN groups.is_default IS '是否默认组 0-否 1-是';
COMMENT ON COLUMN groups.is_guest IS '是否游客组 0-否 1-是';
COMMENT ON COLUMN groups.configs IS '组配置(JSON格式)';
COMMENT ON COLUMN groups.created_at IS '创建时间';
COMMENT ON COLUMN groups.updated_at IS '更新时间';

-- 个人访问令牌表
CREATE TABLE IF NOT EXISTS personal_access_tokens (
    id BIGSERIAL PRIMARY KEY,
    api_name VARCHAR(191) NOT NULL,
    username VARCHAR(191) NOT NULL,
    token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP(3) DEFAULT NULL,
    updated_at TIMESTAMP(3) DEFAULT NULL
);
COMMENT ON TABLE personal_access_tokens IS '个人访问令牌表 - 存储用户API访问令牌';
COMMENT ON COLUMN personal_access_tokens.id IS '令牌ID';
COMMENT ON COLUMN personal_access_tokens.api_name IS 'API名称';
COMMENT ON COLUMN personal_access_tokens.username IS '所属用户名';
COMMENT ON COLUMN personal_access_tokens.token IS '访问令牌';
COMMENT ON COLUMN personal_access_tokens.created_at IS '创建时间';
COMMENT ON COLUMN personal_access_tokens.updated_at IS '更新时间';

-- 用户组与存储策略关联表
CREATE TABLE IF NOT EXISTS group_strategies (
    group_id BIGINT NOT NULL,
    strategy_id BIGINT NOT NULL,
    created_at BIGINT DEFAULT NULL,
    updated_at BIGINT DEFAULT NULL,
    PRIMARY KEY (group_id, strategy_id)
);
COMMENT ON TABLE group_strategies IS '用户组与存储策略关联表 - 实现多对多关系';
COMMENT ON COLUMN group_strategies.group_id IS '用户组ID';
COMMENT ON COLUMN group_strategies.strategy_id IS '存储策略ID';
COMMENT ON COLUMN group_strategies.created_at IS '创建时间';
COMMENT ON COLUMN group_strategies.updated_at IS '更新时间';

-- 存储策略表
CREATE TABLE IF NOT EXISTS strategies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(191) NOT NULL UNIQUE,
    introduction VARCHAR(255) NOT NULL,
    strategy_key VARCHAR(191) NOT NULL UNIQUE,
    configs TEXT NOT NULL,
    created_at TIMESTAMP(3) DEFAULT NULL,
    updated_at TIMESTAMP(3) DEFAULT NULL
);
COMMENT ON TABLE strategies IS '存储策略表 - 定义图片存储后端配置';
COMMENT ON COLUMN strategies.id IS '策略ID';
COMMENT ON COLUMN strategies.name IS '策略名称';
COMMENT ON COLUMN strategies.introduction IS '策略简介';
COMMENT ON COLUMN strategies.strategy_key IS '策略标识符';
COMMENT ON COLUMN strategies.configs IS '策略配置(JSON格式)';
COMMENT ON COLUMN strategies.created_at IS '创建时间';
COMMENT ON COLUMN strategies.updated_at IS '更新时间';

-- 图片信息表
CREATE TABLE IF NOT EXISTS images (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    strategy_id BIGINT NOT NULL,
    image_key VARCHAR(191) NOT NULL,
    path VARCHAR(191) NOT NULL,
    name VARCHAR(191) NOT NULL,
    origin_name VARCHAR(191) NOT NULL,
    size BIGINT NOT NULL,
    mimetype VARCHAR(191) NOT NULL,
    extension VARCHAR(191) NOT NULL,
    md5 VARCHAR(191) NOT NULL,
    sha1 VARCHAR(191) NOT NULL,
    width BIGINT NOT NULL,
    height BIGINT NOT NULL,
    permissions BIGINT NOT NULL DEFAULT 0,
    is_unhealthy BIGINT NOT NULL DEFAULT 0,
    upload_ip VARCHAR(191) NOT NULL,
    created_at TIMESTAMP(3) DEFAULT NULL,
    updated_at TIMESTAMP(3) DEFAULT NULL
);
COMMENT ON TABLE images IS '图片信息表 - 存储所有图片元数据';
COMMENT ON COLUMN images.id IS '图片ID';
COMMENT ON COLUMN images.user_id IS '上传用户ID';
COMMENT ON COLUMN images.group_id IS '所属用户组ID';
COMMENT ON COLUMN images.strategy_id IS '使用的存储策略ID';
COMMENT ON COLUMN images.image_key IS '图片唯一标识符';
COMMENT ON COLUMN images.path IS '图片存储路径';
COMMENT ON COLUMN images.name IS '图片文件名';
COMMENT ON COLUMN images.origin_name IS '原始文件名';
COMMENT ON COLUMN images.size IS '文件大小(字节)';
COMMENT ON COLUMN images.mimetype IS 'MIME类型';
COMMENT ON COLUMN images.extension IS '文件扩展名';
COMMENT ON COLUMN images.md5 IS '文件MD5值';
COMMENT ON COLUMN images.sha1 IS '文件SHA1值';
COMMENT ON COLUMN images.width IS '图片宽度';
COMMENT ON COLUMN images.height IS '图片高度';
COMMENT ON COLUMN images.permissions IS '访问权限 0-公开 1-私有';
COMMENT ON COLUMN images.is_unhealthy IS '是否不健康 0-否 1-是';
COMMENT ON COLUMN images.upload_ip IS '上传IP地址';
COMMENT ON COLUMN images.created_at IS '创建时间';
COMMENT ON COLUMN images.updated_at IS '更新时间';
CREATE INDEX IF NOT EXISTS idx_images_user_id ON images (user_id);
CREATE INDEX IF NOT EXISTS idx_images_group_id ON images (group_id);

-- 图片分享记录表
CREATE TABLE IF NOT EXISTS shares (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    image_id BIGINT NOT NULL,
    share_code VARCHAR(32) NOT NULL UNIQUE,
    password VARCHAR(255),
    expires_at TIMESTAMP(3),
    view_count INT NOT NULL DEFAULT 0,
    max_views INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP(3) DEFAULT NULL,
    updated_at TIMESTAMP(3) DEFAULT NULL
);
COMMENT ON TABLE shares IS '图片分享记录表 - 存储用户分享图片记录';
COMMENT ON COLUMN shares.id IS '分享ID';
COMMENT ON COLUMN shares.user_id IS '用户ID';
COMMENT ON COLUMN shares.image_id IS '图片ID';
COMMENT ON COLUMN shares.share_code IS '分享码';
COMMENT ON COLUMN shares.password IS '分享密码(加密存储)';
COMMENT ON COLUMN shares.expires_at IS '过期时间';
COMMENT ON COLUMN shares.view_count IS '查看次数';
COMMENT ON COLUMN shares.max_views IS '最大查看次数 0表示无限制';
COMMENT ON COLUMN shares.is_active IS '是否激活';
COMMENT ON COLUMN shares.created_at IS '创建时间';
COMMENT ON COLUMN shares.updated_at IS '更新时间';
CREATE INDEX IF NOT EXISTS idx_shares_user_id ON shares (user_id);
CREATE INDEX IF NOT EXISTS idx_shares_image_id ON shares (image_id);
