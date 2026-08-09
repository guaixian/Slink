package model

import (
	"embed"
	"fmt"
	"os"

	mysqlDriver "gorm.io/driver/mysql"
	postgresDriver "gorm.io/driver/postgres"
	sqlserverDriver "gorm.io/driver/sqlserver"
	sqliteDriver "github.com/glebarez/sqlite"
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

	// 自动迁移所有表
	if err := migrateAllTables(db); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	return nil
}

// migrateAllTables 自动迁移所有模型表
func migrateAllTables(db *gorm.DB) error {
	tables := []interface{}{
		&User{},
		&Config{},
		&Groups{},
		&PersonalAccessToken{},
		&GroupStrategy{},
		&UserStrategy{},
		&Strategies{},
		&Images{},
		&Share{},
		&GlobalUploadPolicy{},
		&UploadPolicyGroup{},
	}
	for _, t := range tables {
		if err := db.AutoMigrate(t); err != nil {
			return err
		}
	}

	// 初始化默认数据
	if err := initAllDefaultData(db); err != nil {
		return err
	}

	return nil
}

// initAllDefaultData 初始化所有默认数据
func initAllDefaultData(db *gorm.DB) error {
	if err := InitDefaultConfigs(db); err != nil {
		return err
	}
	if err := InitDefaultGroups(db); err != nil {
		return err
	}
	if err := InitDefaultStrategies(db); err != nil {
		return err
	}
	if err := InitDefaultUploadPolicyGroup(db); err != nil {
		return err
	}
	if err := MigrateLegacyUploadPolicy(db); err != nil {
		return err
	}
	if err := EnsureUploadPolicySeeded(db); err != nil {
		return err
	}
	// 迁移现有用户（旧 group_id=0 的用户）
	db.Model(&User{}).Where("group_id = ?", 0).Update("group_id", 1)
	// 注意：不再自动创建 admin@slink.org/123456 默认管理员（固定口令属于后门账号），
	// 管理员一律由 /api/init/setup 初始化流程创建。
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

// TestDBConnection 仅测试数据库连通性：打开连接并 Ping 后立即关闭。
// 不创建数据库、不执行迁移、不修改全局 DB（供初始化向导的"测试连接"使用，
// 避免测试动作污染正在运行的系统或目标库）。
func TestDBConnection(config *DBConfig) error {
	var db *gorm.DB
	var err error

	switch config.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local&timeout=5s",
			config.User, config.Password, config.Host, config.Port, config.Charset)
		db, err = gorm.Open(mysqlDriver.Open(dsn), &gorm.Config{})
	case "postgres", "postgresql":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s connect_timeout=5",
			config.Host, config.Port, config.User, config.Password, config.SSLMode)
		db, err = gorm.Open(postgresDriver.Open(dsn), &gorm.Config{})
	case "sqlserver", "mssql":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=master&connection+timeout=5",
			config.User, config.Password, config.Host, config.Port)
		db, err = gorm.Open(sqlserverDriver.Open(dsn), &gorm.Config{})
	case "sqlite":
		fallthrough
	default:
		// SQLite 无需网络连接，仅校验能否打开（目录可写性由正式初始化保证）
		db, err = gorm.Open(sqliteDriver.Open(config.DBName), &gorm.Config{})
	}
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取连接失败: %w", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("Ping 失败: %w", err)
	}
	return nil
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
