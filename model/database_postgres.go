package model

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitPostgreSQL 初始化PostgreSQL数据库连接
// 连接到PostgreSQL数据库并创建数据库（如果不存在）
func InitPostgreSQL(config *DBConfig) (*gorm.DB, error) {
	// 先连接postgres系统数据库
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接PostgreSQL服务器失败: %w", err)
	}

	// 创建数据库
	createDBsql := fmt.Sprintf("CREATE DATABASE \"%s\"", config.DBName)
	if err := db.Exec(createDBsql).Error; err != nil {
		// 数据库可能已存在，忽略错误
	}

	// 关闭当前连接
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.Close()

	// 连接到指定数据库
	dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 不再执行旧版 sql/postgresql.sql:其表结构与 GORM 模型已漂移
	//(缺 policy_group_id 列、email 唯一约束名与 uniqueIndex 冲突会导致
	// AutoMigrate DROP CONSTRAINT uni_users_email 失败),表结构统一由
	// migrateAllTables 的 AutoMigrate 创建/演进。
	return db, nil
}
