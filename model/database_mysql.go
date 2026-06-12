package model

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitMySQL 初始化MySQL数据库连接
// 连接到MySQL数据库并创建数据库（如果不存在）
func InitMySQL(config *DBConfig) (*gorm.DB, error) {
	// 先连接MySQL服务器（不指定数据库）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local&multiStatements=true",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Charset,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接MySQL服务器失败: %w", err)
	}

	// 创建数据库
	createDBsql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET %s COLLATE utf8mb4_unicode_ci",
		config.DBName,
		config.Charset,
	)
	if err := db.Exec(createDBsql).Error; err != nil {
		return nil, fmt.Errorf("创建数据库失败: %w", err)
	}

	// 关闭当前连接
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.Close()

	// 连接到指定数据库
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&default_storage_engine=InnoDB",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.Charset,
	)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 执行SQL初始化脚本
	if err := executeSQLFile(db, "mysql"); err != nil {
		return nil, err
	}

	return db, nil
}
