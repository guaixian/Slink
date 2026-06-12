package model

import (
	"embed"
	"fmt"
	"os"

	"gorm.io/gorm"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

// DBConfig 数据库配置
// 存储连接各种数据库所需的配置信息
type DBConfig struct {
	Type     string // sqlite, mysql, postgres, sqlserver
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string // for postgres
	Charset  string // for mysql
}

// GetDBConfigFromEnv 从环境变量获取数据库配置
// 支持通过环境变量配置数据库连接参数
func GetDBConfigFromEnv() *DBConfig {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	}

	config := &DBConfig{
		Type:     dbType,
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		Charset:  os.Getenv("DB_CHARSET"),
	}

	// 设置默认值
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	if config.Charset == "" {
		config.Charset = "utf8mb4"
	}
	if config.DBName == "" {
		config.DBName = "data.db"
	}

	// 解析端口
	portStr := os.Getenv("DB_PORT")
	if portStr != "" {
		fmt.Sscanf(portStr, "%d", &config.Port)
	} else {
		// 设置默认端口
		switch config.Type {
		case "mysql":
			config.Port = 3306
		case "postgres", "postgresql":
			config.Port = 5432
		case "sqlserver", "mssql":
			config.Port = 1433
		}
	}

	return config
}

// InitDBWithConfig 使用配置初始化数据库
// 根据配置类型连接相应的数据库并执行初始化SQL
func InitDBWithConfig(config *DBConfig) error {
	var db *gorm.DB
	var err error

	switch config.Type {
	case "mysql":
		db, err = InitMySQL(config)
	case "postgres", "postgresql":
		db, err = InitPostgreSQL(config)
	case "sqlserver", "mssql":
		db, err = InitSQLServer(config)
	case "sqlite":
		fallthrough
	default:
		db, err = InitSQLite(config)
	}

	if err != nil {
		return fmt.Errorf("数据库初始化失败: %w", err)
	}

	DB = db

	// 非 SQLite 初始化脚本可能不含该表；SQLite 亦在此再次 AutoMigrate 保证存在
	if err := db.AutoMigrate(&GlobalUploadPolicy{}); err != nil {
		return fmt.Errorf("迁移 global_upload_policies 失败: %w", err)
	}
	if err := EnsureUploadPolicySeeded(db); err != nil {
		return fmt.Errorf("初始化全局上传策略失败: %w", err)
	}

	return nil
}

// CloseDB 关闭全局连接并将 DB 置空，便于全量恢复后重新 InitDBWithConfig。
func CloseDB() error {
	if DB == nil {
		return nil
	}
	sq, err := DB.DB()
	if err != nil {
		DB = nil
		return err
	}
	err = sq.Close()
	DB = nil
	return err
}

// executeSQLFile 执行SQL文件
// 从sql目录读取并执行对应的SQL初始化脚本
func executeSQLFile(db *gorm.DB, dbType string) error {
	sqlFileName := fmt.Sprintf("sql/%s.sql", dbType)
	if dbType == "sqlite" {
		return nil // SQLite使用GORM AutoMigrate
	}

	data, err := sqlFiles.ReadFile(sqlFileName)
	if err != nil {
		return fmt.Errorf("读取SQL文件失败: %w", err)
	}

	// 执行SQL语句
	sqlContent := string(data)
	if err := db.Exec(sqlContent).Error; err != nil {
		return fmt.Errorf("执行SQL失败: %w", err)
	}

	return nil
}

// initDefaultData 初始化默认数据（SQLite专用）
// 创建系统运行所需的初始配置数据
func initDefaultData(db *gorm.DB) error {
	// 初始化默认配置
	if err := InitDefaultConfigs(db); err != nil {
		return err
	}

	// 初始化默认用户组
	if err := InitDefaultGroups(db); err != nil {
		return err
	}

	// 初始化默认存储策略
	if err := InitDefaultStrategies(db); err != nil {
		return err
	}

	return nil
}
