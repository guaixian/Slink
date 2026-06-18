package api

import (
	"Slink/applog"
	"Slink/imaging"
	"Slink/middleware"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProcessImage 本地图片处理（无外部依赖）。
//
// 以 multipart/form-data 提交：
//   - image      文件，必填
//   - operation  compress|convert|resize|thumbnail，可选（缺省按其它参数推断）
//   - format     输出格式 jpeg/png/gif/bmp，可选
//   - quality    JPEG 质量 1-100，可选
//   - width      目标宽度（resize），可选
//   - height     目标高度（resize），可选
//   - scale      缩放系数（如 2 表示放大两倍），可选
//   - max_size   缩略图边界框最长边，可选
//   - response   download(默认，返回二进制) | json(返回 base64 与元数据)
//
// 处理在内存中完成并直接返回结果，不写入存储；如需保存可将结果再次走上传接口。
func ProcessImage(c *gin.Context) {
	rid := middleware.RequestID(c)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未选择图片"})
		return
	}

	// 以全局上传策略的大小上限作为处理输入的上限，避免超大文件占用内存
	policy, ok := readGlobalUploadPolicy(c)
	if !ok {
		return
	}
	if file.Size > int64(policy.MaximumFileSize*1024) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小超过限制"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	data, err := io.ReadAll(io.LimitReader(src, int64(policy.MaximumFileSize*1024)+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	opts := imaging.Options{
		Operation: c.PostForm("operation"),
		Format:    c.PostForm("format"),
		Quality:   atoiDefault(c.PostForm("quality"), 0),
		Width:     atoiDefault(c.PostForm("width"), 0),
		Height:    atoiDefault(c.PostForm("height"), 0),
		Scale:     atofDefault(c.PostForm("scale"), 0),
		MaxSize:   atoiDefault(c.PostForm("max_size"), 0),
	}

	result, err := imaging.Process(data, opts)
	if err != nil {
		applog.Logger.Warn("process image failed", "request_id", rid, "user_id", userID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	applog.Logger.Info("process image ok", "request_id", rid, "user_id", userID,
		"operation", opts.Operation, "out_format", result.Format,
		"out_w", result.Width, "out_h", result.Height, "out_bytes", len(result.Data))

	mime := imaging.MimeTypeForFormat(result.Format)

	if c.PostForm("response") == "json" {
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "处理成功",
			"data": gin.H{
				"format":      result.Format,
				"mimetype":    mime,
				"width":       result.Width,
				"height":      result.Height,
				"size_bytes":  len(result.Data),
				"base64":      base64.StdEncoding.EncodeToString(result.Data),
				"data_uri":    "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(result.Data),
				"origin_name": file.Filename,
			},
		})
		return
	}

	// 默认返回二进制，并通过响应头暴露处理后元数据
	c.Header("X-Image-Format", result.Format)
	c.Header("X-Image-Width", strconv.Itoa(result.Width))
	c.Header("X-Image-Height", strconv.Itoa(result.Height))
	c.Header("X-Image-Bytes", strconv.Itoa(len(result.Data)))
	c.Data(http.StatusOK, mime, result.Data)
}

// GetProcessCapabilities 返回本地图片处理能力，便于前端按需展示
func GetProcessCapabilities(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data": gin.H{
			"operations":        []string{imaging.OpCompress, imaging.OpConvert, imaging.OpResize, imaging.OpThumbnail},
			"input_formats":     imaging.SupportedInputFormats(),
			"output_formats":    imaging.SupportedOutputFormats(),
			"max_dimension":     imaging.MaxDimension,
			"default_quality":   imaging.DefaultJPEGQuality,
			"note":              "本地处理为传统高质量重采样；AI 超分/智能去水印等见 /api/image/ai/* 外置扩展",
			"webp_output_local": false,
		},
	})
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func atofDefault(s string, def float64) float64 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}
