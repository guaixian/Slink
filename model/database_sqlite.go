package model

import (
	"fmt"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// preMigrateSqliteColumnRenames 在 AutoMigrate 之前，将旧版 GORM 映射的 `key` 列改名为显式列名。
// 否则已存在且含数据的表在 ADD NOT NULL 列时会出现：
// "Cannot add a NOT NULL column with default value NULL"（SQLite 限制）。
func preMigrateSqliteColumnRenames(db *gorm.DB) error {
	renames := []struct{ table, from, to string }{
		{"configs", "key", "config_key"},
		{"strategies", "key", "strategy_key"},
		{"images", "key", "image_key"},
	}
	for _, r := range renames {
		if err := sqliteRenameColumnIfNeeded(db, r.table, r.from, r.to); err != nil {
			return err
		}
	}
	return nil
}

func sqliteQuIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

// sqliteCountColumn 使用 pragma_table_info 判断列是否存在（表名写死为白名单，避免拼进不可信内容）
func sqliteCountColumn(db *gorm.DB, table, col string) (int64, error) {
	var query string
	switch table {
	case "configs":
		query = "SELECT COUNT(1) FROM pragma_table_info('configs') AS i WHERE i.name = ?"
	case "strategies":
		query = "SELECT COUNT(1) FROM pragma_table_info('strategies') AS i WHERE i.name = ?"
	case "images":
		query = "SELECT COUNT(1) FROM pragma_table_info('images') AS i WHERE i.name = ?"
	default:
		return 0, fmt.Errorf("sqlite: 不支持的表名 %q", table)
	}
	var n int64
	err := db.Raw(query, col).Scan(&n).Error
	return n, err
}

func sqliteRenameColumnIfNeeded(db *gorm.DB, table, from, to string) error {
	if !db.Migrator().HasTable(table) {
		return nil
	}
	fromCnt, err := sqliteCountColumn(db, table, from)
	if err != nil {
		return err
	}
	toCnt, err := sqliteCountColumn(db, table, to)
	if err != nil {
		return err
	}
	if fromCnt == 0 || toCnt > 0 {
		return nil
	}
	renameSQL := fmt.Sprintf(
		"ALTER TABLE %s RENAME COLUMN %s TO %s",
		sqliteQuIdent(table), sqliteQuIdent(from), sqliteQuIdent(to),
	)
	return db.Exec(renameSQL).Error
}

// InitSQLite 初始化SQLite数据库连接
// 连接到SQLite数据库文件，如果文件不存在则创建
// SQLite使用GORM AutoMigrate自动创建表结构
func InitSQLite(config *DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(config.DBName), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接SQLite数据库失败: %w", err)
	}

	if err := preMigrateSqliteColumnRenames(db); err != nil {
		return nil, fmt.Errorf("预迁移表结构(列重命名)失败: %w", err)
	}

	// 使用GORM AutoMigrate自动创建表结构
	if err := db.AutoMigrate(
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
	); err != nil {
		return nil, fmt.Errorf("迁移表结构失败: %w", err)
	}

	// 初始化默认数据
	if err := initDefaultData(db); err != nil {
		return nil, err
	}
	// 初始化默认上传策略组
	if err := InitDefaultUploadPolicyGroup(db); err != nil {
		return nil, err
	}

	return db, nil
}
