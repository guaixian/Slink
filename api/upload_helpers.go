package api

import (
	"Slink/model"
	"Slink/utils"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// remoteImageHTTPClient URL 抓取专用客户端：带超时，限制重定向次数
var remoteImageHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("重定向次数过多")
		}
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			return fmt.Errorf("不允许的协议: %s", req.URL.Scheme)
		}
		return nil
	},
}

// fetchRemoteImage 从 URL 下载图片数据。
// 仅允许 http/https；无论对方是否返回 Content-Length，都把读取大小限制在 maxSizeKB 内。
func fetchRemoteImage(rawURL string, maxSizeKB uint) ([]byte, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, "", fmt.Errorf("无效的URL，仅支持 http/https")
	}

	resp, err := remoteImageHTTPClient.Get(rawURL)
	if err != nil {
		return nil, "", fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", fmt.Errorf("URL不是有效的图片")
	}

	maxBytes := int64(maxSizeKB) * 1024
	if resp.ContentLength > maxBytes {
		return nil, "", fmt.Errorf("文件大小超过限制（%dKB）", maxSizeKB)
	}

	// 流式限制读取大小，Content-Length 缺失或虚报时也能拦截
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("读取图片数据失败")
	}
	if int64(len(data)) > maxBytes {
		return nil, "", fmt.Errorf("文件大小超过限制（%dKB）", maxSizeKB)
	}
	return data, contentType, nil
}


// formatBytesShort 容量提示用的紧凑格式化
func formatBytesShort(n int64) string {
	if n < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(n)/1024)
	}
	if n < 1024*1024*1024 {
		return fmt.Sprintf("%.2fMB", float64(n)/(1024*1024))
	}
	return fmt.Sprintf("%.2fGB", float64(n)/(1024*1024*1024))
}

// checkUserCapacity 校验用户剩余存储空间是否容得下本次上传;Capacity 为 0 表示无限制
func checkUserCapacity(c *gin.Context, user *model.User, incomingBytes int64) bool {
	if user.Capacity == 0 {
		return true
	}
	var used int64
	model.DB.Model(&model.Images{}).Where("user_id = ?", user.ID).Select("COALESCE(SUM(size), 0)").Scan(&used)
	if used+incomingBytes > int64(user.Capacity) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("存储空间不足（已用 %s / 上限 %s）",
				formatBytesShort(used), formatBytesShort(int64(user.Capacity))),
		})
		return false
	}
	return true
}

// hashBytesMD5 计算字节数据的 MD5（URL 上传的内存数据去重用）
func hashBytesMD5(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

// readUploadedFile 读取上传文件的全部字节（大小已被策略上限约束，可安全读入内存）
func readUploadedFile(file *multipart.FileHeader) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	return io.ReadAll(src)
}

// writeUploadSuccess 统一的上传成功响应
func writeUploadSuccess(c *gin.Context, rec *model.Images, originName string, sizeMB float64, links map[string]string) {
	c.JSON(http.StatusOK, &UploadImageResponse{
		Status:  true,
		Message: "上传成功",
		Data: &ImageData{
			ID:         rec.ID,
			Pathname:   normalizePathnameForLinks(rec.Path),
			OriginName: originName,
			Size:       sizeMB,
			Mimetype:   rec.Mimetype,
			MD5:        rec.Md5,
			SHA1:       rec.SHA1,
			Links:      links,
		},
	})
}

// attemptDedupUpload 落盘前按内容 MD5 去重。
// 命中已有原始图片时：不再写入存储，仅创建一条引用记录（RefImageID 指向原始图，
// 与原始图共享同一存储路径/URL），完成响应并返回 true；未命中返回 false 继续正常上传流程。
func attemptDedupUpload(c *gin.Context, md5Str string, userID uint, originName string, userConfig model.UserConfig) bool {
	if md5Str == "" {
		return false
	}
	existing, err := checkMD5Duplicate(model.DB, md5Str)
	if err != nil || existing == nil {
		return false
	}

	refID := existing.ID
	rec := &model.Images{
		UserID:      userID,
		GroupID:     model.EffectiveImageGroupID(model.DB),
		StrategyID:  existing.StrategyID,
		ImageKey:    existing.ImageKey,
		Path:        existing.Path,
		Name:        existing.Name,
		OriginName:  originName,
		Size:        existing.Size,
		Mimetype:    existing.Mimetype,
		Extension:   existing.Extension,
		Md5:         existing.Md5,
		SHA1:        existing.SHA1,
		Width:       existing.Width,
		Height:      existing.Height,
		Permissions: uint(userConfig.DefaultPermission),
		IsUnhealthy: 0,
		UploadIp:    c.ClientIP(),
		RefImageID:  &refID,
	}
	model.SetImagePermissionFromUserConfig(rec, userConfig)

	if err := model.CreateImage(model.DB, rec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片记录失败"})
		return true
	}
	if err := model.IncrementUserImageCount(model.DB, userID); err != nil {
		// 计数失败不影响上传结果
	}

	utilsConfig, err := strategyUtilsConfigForStoredImage(model.DB, existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取存储策略失败"})
		return true
	}
	pathname := normalizePathnameForLinks(existing.Path)
	links := utils.GenerateImageLinksWithConfig(pathname, originName, utilsConfig)
	writeUploadSuccess(c, rec, originName, float64(existing.Size)/(1024*1024), links)
	return true
}

// resolveUploadStrategy 解析本次上传使用的存储策略
// 优先使用请求中指定的 strategy_id，其次查用户分配的策略(UserStrategy表)，再回退到用户偏好中的 default_strategy，最后用第一个可用策略
func resolveUploadStrategy(db *gorm.DB, userID uint, userConfig model.UserConfig, strategyIDRaw string) (*model.Strategies, *model.StrategyConfigData, error) {
	// 1. 解析请求中指定的 strategy_id
	var requestedSID uint
	if strategyIDRaw != "" {
		if v, err := strconv.ParseUint(strategyIDRaw, 10, 32); err == nil && v > 0 {
			requestedSID = uint(v)
		}
	}

	all, err := model.GetAllStrategies(db)
	if err != nil {
		return nil, nil, err
	}
	if len(all) == 0 {
		return nil, nil, fmt.Errorf("未配置存储策略")
	}

	// 2. 获取用户分配的策略列表
	userStrategyIDs, _ := model.GetStrategyIDsByUserID(db, userID)

	// 构建可用策略ID集合（用户分配的策略，或全部策略如果是管理员/无分配记录）
	isAdmin := false
	if u, err := model.GetUserByID(db, userID); err == nil && u.IsAdmin == 1 {
		isAdmin = true
	}

	// 管理员可以看到所有策略；普通用户只能看到分配给他们的策略
	availableIDs := make(map[uint]bool)
	if isAdmin || len(userStrategyIDs) == 0 {
		for i := range all {
			availableIDs[all[i].ID] = true
		}
	} else {
		for _, sid := range userStrategyIDs {
			availableIDs[sid] = true
		}
	}

	// 3. 如果请求指定了策略ID，验证是否可用
	if requestedSID > 0 {
		if availableIDs[requestedSID] {
			if st, err := model.GetStrategyByID(db, requestedSID); err == nil {
				scd, err := model.GetStrategyConfigData(st)
				return st, scd, err
			}
		}
	}

	// 4. 回退到用户偏好中的 DefaultStrategy
	sid := userConfig.DefaultStrategy
	if availableIDs[uint(sid)] {
		if st, err := model.GetStrategyByID(db, uint(sid)); err == nil {
			scd, err := model.GetStrategyConfigData(st)
			return st, scd, err
		}
	}

	// 5. 使用第一个可用策略
	var st *model.Strategies
	for i := range all {
		if availableIDs[all[i].ID] {
			st = &all[i]
			break
		}
	}
	if st == nil {
		st = &all[0]
	}
	scd, err := model.GetStrategyConfigData(st)
	if err != nil {
		return nil, nil, err
	}
	return st, scd, nil
}

// checkMD5Duplicate 检查MD5是否已存在，返回已存在的原始图片记录
// 用于上传去重：如果相同MD5的图片已存在，则返回已有记录，新上传只需创建引用记录
func checkMD5Duplicate(db *gorm.DB, md5 string) (*model.Images, error) {
	img, err := model.GetImageByMD5(db, md5)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return img, nil
}

func strategyUtilsConfigFromModel(strategy *model.Strategies) (*utils.StrategyConfig, error) {
	sc, err := model.GetStrategyConfig(strategy)
	if err != nil {
		return nil, err
	}
	baseURL := strings.TrimSpace(sc.URL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(sc.Domain)
	}
	return &utils.StrategyConfig{
		URL:      baseURL,
		Root:     sc.Root,
		Queries:  sc.Queries,
		AuthType: sc.AuthType,
	}, nil
}

func strategyUtilsConfigForStoredImage(db *gorm.DB, img *model.Images) (*utils.StrategyConfig, error) {
	st, err := model.GetStrategyForStoredImage(db, img.StrategyID)
	if err != nil {
		return nil, err
	}
	return strategyUtilsConfigFromModel(st)
}

func normalizePathnameForLinks(path string) string {
	pathname := path
	if len(pathname) > 7 && pathname[:7] == "static/" {
		pathname = pathname[7:]
	} else if len(pathname) > 8 && pathname[:8] == `static\` {
		pathname = pathname[8:]
	}
	return pathname
}

func readGlobalUploadPolicy(c *gin.Context) (*model.GroupConfig, bool) {
	pol, err := model.GetGlobalUploadPolicy(model.DB)
	if err != nil {
		c.JSON(500, gin.H{"error": "读取全局上传策略失败"})
		return nil, false
	}
	return pol, true
}

// readUserUploadPolicy 读取用户所属上传策略组的配置（优先用户策略组，否则默认策略组，否则全局策略）
func readUserUploadPolicy(c *gin.Context, userID uint) (*model.GroupConfig, bool) {
	// 优先：用户所属上传策略组
	pg, err := model.GetUserUploadPolicyGroup(model.DB, userID)
	if err == nil && pg != nil {
		return pg.ToGroupConfig(), true
	}

	// 其次：默认策略组
	defaultPg, err := model.GetDefaultUploadPolicyGroup(model.DB)
	if err == nil && defaultPg != nil {
		return defaultPg.ToGroupConfig(), true
	}

	// 最后：旧的全局策略
	pol, err := model.GetGlobalUploadPolicy(model.DB)
	if err == nil {
		return pol, true
	}

	c.JSON(500, gin.H{"error": "未找到上传策略配置"})
	return nil, false
}

// createImageRecordWithDedup 创建图片记录（含MD5去重逻辑）
// 如果相同MD5的原始图片已存在，则创建引用记录（skip physical upload）
// 返回 imageRecord, isDuplicate, error
func createImageRecordWithDedup(imageRecord *model.Images) (*model.Images, bool, error) {
	db := model.DB

	// 检查MD5是否已存在
	existing, err := checkMD5Duplicate(db, imageRecord.Md5)
	if err != nil {
		return nil, false, fmt.Errorf("MD5去重检查失败: %w", err)
	}

	if existing != nil {
		// 命中相同图片：创建引用记录，指向已存在的原始图片
		refID := existing.ID
		imageRecord.RefImageID = &refID
		imageRecord.Path = existing.Path
		imageRecord.Name = existing.Name
		imageRecord.Size = existing.Size
		imageRecord.Mimetype = existing.Mimetype
		imageRecord.Extension = existing.Extension
		imageRecord.Width = existing.Width
		imageRecord.Height = existing.Height
		imageRecord.StrategyID = existing.StrategyID

		if err := model.CreateImage(db, imageRecord); err != nil {
			return nil, false, fmt.Errorf("创建引用图片记录失败: %w", err)
		}
		return imageRecord, true, nil
	}

	// 未命中：正常创建新记录
	if err := model.CreateImage(db, imageRecord); err != nil {
		return nil, false, fmt.Errorf("创建图片记录失败: %w", err)
	}
	return imageRecord, false, nil
}
