package api

import (
	"Slink/applog"
	"Slink/model"
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	backupManifestName   = "manifest.json"
	backupSQLiteInZip    = "database.sqlite"
	backupConfigDir      = "config"
	backupStaticDir      = "static"
	backupMaxImportBytes = 600 << 20 // 600 MiB
)

// backupManifestV1 与压缩包中 manifest.json 对应
type backupManifestV1 struct {
	Version         int    `json:"version"`
	App             string `json:"app"`
	CreatedAt       string `json:"created_at"`
	DBType          string `json:"db_type"`
	SQLiteInZip     bool   `json:"sqlite_in_zip"`
	ExportDBName    string `json:"export_db_name"`
	SQLiteInZipName string `json:"sqlite_in_zip_name,omitempty"`
	Note            string `json:"note,omitempty"`
}

func isAllowedImportZipPath(p string) bool {
	p = filepath.ToSlash(strings.TrimSpace(p))
	if p == "" || p == ".." || strings.HasPrefix(p, "/") {
		return false
	}
	if strings.Contains(p, "../") {
		return false
	}
	if p == backupManifestName || p == backupSQLiteInZip {
		return true
	}
	if strings.HasPrefix(p, "config/") {
		return true
	}
	if p == "static" || strings.HasPrefix(p, "static/") {
		return true
	}
	return false
}

func isSqliteDBType(t string) bool {
	t = strings.ToLower(strings.TrimSpace(t))
	return t == "" || t == "sqlite" || t == "sqlite3"
}

// ExportFullBackup 打包 config 目录、SQLite 数据文件、static 下的本地资源为 zip
func ExportFullBackup(c *gin.Context) {
	cfg, err := model.LoadDatabaseConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "读取数据库配置失败", "error": err.Error()})
		return
	}

	manifest := backupManifestV1{
		Version:   1,
		App:       "slink",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		DBType:    strings.ToLower(strings.TrimSpace(cfg.Type)),
	}
	if manifest.DBType == "" {
		manifest.DBType = "sqlite"
	}
	manifest.SQLiteInZip = isSqliteDBType(manifest.DBType)
	manifest.ExportDBName = cfg.DBName
	manifest.SQLiteInZipName = backupSQLiteInZip
	if !manifest.SQLiteInZip {
		manifest.Note = "当前为非 SQLite 数据库，压缩包中不包含数据库转储，仅含 config 与 static；恢复数据库请使用外部工具。"
	}

	c.Header("Content-Type", "application/zip")
	fname := fmt.Sprintf("slink-backup-%s.zip", time.Now().Format("20060102-150405"))
	c.Header("Content-Disposition", "attachment; filename="+fname)
	c.Status(http.StatusOK)

	zw := zip.NewWriter(c.Writer)
	defer func() { _ = zw.Close() }()

	manData, _ := json.MarshalIndent(manifest, "", "  ")
	if err := writeZipFile(zw, backupManifestName, manData); err != nil {
		applog.Logger.Error("export backup: write manifest", "error", err)
		return
	}

	_ = os.MkdirAll(backupConfigDir, 0755)
	if err := filepath.Walk(backupConfigDir, func(path string, info os.FileInfo, werr error) error {
		if werr != nil {
			if os.IsNotExist(werr) {
				return nil
			}
			return werr
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(backupConfigDir, path)
		if err != nil {
			return err
		}
		zipName := "config/" + filepath.ToSlash(rel)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return writeZipFile(zw, zipName, data)
	}); err != nil {
		applog.Logger.Error("export backup: config walk", "error", err)
		return
	}

	if manifest.SQLiteInZip {
		dbPath := strings.TrimSpace(cfg.DBName)
		if dbPath == "" {
			dbPath = "data.db"
		}
		if data, rerr := os.ReadFile(dbPath); rerr == nil {
			if err := writeZipFile(zw, backupSQLiteInZip, data); err != nil {
				applog.Logger.Error("export backup: add sqlite", "error", err)
				return
			}
		} else {
			applog.Logger.Warn("export backup: db file 不可读，仅导出配置与 static", "path", dbPath, "error", rerr)
		}
	}

	_ = os.MkdirAll(backupStaticDir, 0755)
	if err := filepath.Walk(backupStaticDir, func(path string, info os.FileInfo, werr error) error {
		if werr != nil {
			if os.IsNotExist(werr) {
				return nil
			}
			return werr
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(backupStaticDir, path)
		if err != nil {
			return err
		}
		zipName := "static/" + filepath.ToSlash(rel)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return writeZipFile(zw, zipName, data)
	}); err != nil {
		applog.Logger.Error("export backup: static walk", "error", err)
		return
	}
}

func writeZipFile(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

const restoreConfirmToken = "RESTORE"

// ImportFullBackup 从 zip 恢复 config、可选 SQLite、static
func ImportFullBackup(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, backupMaxImportBytes)
	if err := c.Request.ParseMultipartForm(backupMaxImportBytes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "解析上传失败，文件可能超过大小限制 (600MB)", "error": err.Error()})
		return
	}
	if strings.TrimSpace(c.PostForm("confirm")) != restoreConfirmToken {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "拒绝恢复：请在上传时附带字段 confirm=RESTORE 以确认将覆盖本机数据",
		})
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "需要 multipart 文件字段 file", "error": err.Error()})
		return
	}
	if !strings.HasSuffix(strings.ToLower(fh.Filename), ".zip") {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "只接受 .zip 文件"})
		return
	}

	zr, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "无法读取上传", "error": err.Error()})
		return
	}
	defer zr.Close()

	tmpZip, err := os.CreateTemp("", "slink-import-*.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "创建临时文件失败", "error": err.Error()})
		return
	}
	tmpPath := tmpZip.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err = io.Copy(tmpZip, zr); err != nil {
		_ = tmpZip.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "保存临时包失败", "error": err.Error()})
		return
	}
	_ = tmpZip.Close()

	reader, err := zip.OpenReader(tmpPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "不是有效的 zip 包", "error": err.Error()})
		return
	}
	defer reader.Close()

	extractDir, err := os.MkdirTemp("", "slink-extract-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "创建解压目录失败", "error": err.Error()})
		return
	}
	defer func() { _ = os.RemoveAll(extractDir) }()

	for _, file := range reader.File {
		if !isAllowedImportZipPath(file.Name) {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "包内存在非法路径: " + file.Name})
			return
		}
		dest := filepath.Join(extractDir, filepath.FromSlash(file.Name))
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(dest, 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "创建目录失败", "error": err.Error()})
				return
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "创建父目录失败", "error": err.Error()})
			return
		}
		rf, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "读取包内文件失败", "error": err.Error()})
			return
		}
		out, err := os.Create(dest)
		if err != nil {
			_ = rf.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "写出解压文件失败", "error": err.Error()})
			return
		}
		_, err = io.Copy(out, rf)
		_ = rf.Close()
		_ = out.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "保存解压数据失败", "error": err.Error()})
			return
		}
	}

	manPath := filepath.Join(extractDir, backupManifestName)
	manData, err := os.ReadFile(manPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "包内缺少 manifest.json"})
		return
	}
	var manifest backupManifestV1
	if err := json.Unmarshal(manData, &manifest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "manifest.json 无法解析", "error": err.Error()})
		return
	}
	if manifest.Version != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "不支持的备份版本", "version": manifest.Version})
		return
	}

	extractedConfig := filepath.Join(extractDir, backupConfigDir)
	if _, err := os.Stat(extractedConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "包内缺少 config 目录", "error": err.Error()})
		return
	}
	extractedDBJSON := filepath.Join(extractDir, "config", "database.json")
	if _, err := os.Stat(extractedDBJSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "包内 config/database.json 缺失", "error": err.Error()})
		return
	}

	pendingCfg, err := model.LoadDatabaseConfigFromFile(extractedDBJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "包内 database.json 无法解析", "error": err.Error()})
		return
	}
	sqliteInExtract := filepath.Join(extractDir, backupSQLiteInZip)
	if isSqliteDBType(pendingCfg.Type) {
		if _, err := os.Stat(sqliteInExtract); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "包内缺少 database.sqlite，无法按 SQLite 恢复", "error": err.Error()})
			return
		}
	}

	// 通过校验后再覆盖本机
	if err := os.MkdirAll(backupConfigDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "创建 config 目录失败", "error": err.Error()})
		return
	}
	ents, _ := os.ReadDir(backupConfigDir)
	for _, e := range ents {
		_ = os.RemoveAll(filepath.Join(backupConfigDir, e.Name()))
	}
	if err := copyDirContents(extractedConfig, backupConfigDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "写入 config 失败", "error": err.Error()})
		return
	}

	if err := model.CloseDB(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "关闭当前数据库失败", "error": err.Error()})
		return
	}

	if isSqliteDBType(pendingCfg.Type) {
		dbName := strings.TrimSpace(pendingCfg.DBName)
		if dbName == "" {
			dbName = "data.db"
		}
		if err := os.MkdirAll(filepath.Dir(dbName), 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "创建数据库目录失败", "error": err.Error()})
			return
		}
		if err := copyFile(sqliteInExtract, dbName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "写入 SQLite 文件失败", "error": err.Error()})
			return
		}
	}

	extractedStatic := filepath.Join(extractDir, backupStaticDir)
	if st, sterr := os.Stat(extractedStatic); sterr == nil && st.IsDir() {
		_ = os.RemoveAll(backupStaticDir)
		if err := copyDirTree(extractedStatic, backupStaticDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "恢复 static 失败，请检查权限后手动修复并重启", "error": err.Error()})
			return
		}
	} else {
		_ = os.MkdirAll(backupStaticDir, 0755)
	}

	if err := model.InitDBWithConfig(pendingCfg); err != nil {
		applog.Logger.Error("import backup: 重新打开数据库失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "恢复后重新打开数据库失败，请检查 config/database.json 与数据文件是否匹配，或重启进程",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "恢复成功。若访问异常，请核对存储策略中的「访问网址」与 static 目录。",
	})
	applog.Logger.Info("full backup import 完成", "db_type", pendingCfg.Type)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	_ = out.Close()
	return err
}

func copyDirContents(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		out := filepath.Join(dstDir, rel)
		if info.IsDir() {
			return os.MkdirAll(out, 0755)
		}
		if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
			return err
		}
		return copyFile(path, out)
	})
}

func copyDirTree(srcRoot, dstRoot string) error {
	_ = os.MkdirAll(dstRoot, 0755)
	return copyDirContents(srcRoot, dstRoot)
}
