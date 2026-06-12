package model

import (
	"fmt"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// InitSQLServer 初始化SQLServer数据库连接
// 连接到SQLServer数据库并创建数据库（如果不存在）
func InitSQLServer(config *DBConfig) (*gorm.DB, error) {
	// 先连接到master数据库
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=master",
		config.User,
		config.Password,
		config.Host,
		config.Port,
	)

	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接SQLServer服务器失败: %w", err)
	}

	// 创建数据库
	createDBsql := fmt.Sprintf("IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = '%s') CREATE DATABASE [%s]", config.DBName, config.DBName)
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
	dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 执行SQL初始化脚本
	if err := executeSQLFile(db, "sqlserver"); err != nil {
		return nil, err
	}

	return db, nil
}
