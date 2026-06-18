// Package imaging 提供纯 Go 实现的本地图片处理能力，无需任何外部服务依赖。
//
// 支持的操作：
//   - compress  ：按质量重新编码（主要作用于 JPEG）以压缩体积
//   - convert   ：在受支持的格式之间转换（jpeg/png/gif/bmp）
//   - resize    ：按目标宽高或缩放系数进行高质量重采样（CatmullRom，近似双三次）
//   - thumbnail ：在保持宽高比的前提下缩放到指定边界框内
//
// 说明：本包实现的“高清放大”是基于高质量重采样的传统插值，并非 AI 超分辨率；
// 真正的 AI 超分/智能去水印等能力请见 aiext 外置扩展接口（依赖外部模型服务）。
package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"

	xbmp "golang.org/x/image/bmp"
	xdraw "golang.org/x/image/draw"

	// 注册解码器，使 image.Decode 能识别对应格式
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

// 操作类型常量
const (
	OpCompress  = "compress"
	OpConvert   = "convert"
	OpResize    = "resize"
	OpThumbnail = "thumbnail"
)

// DefaultJPEGQuality 默认 JPEG 质量
const DefaultJPEGQuality = 85

// MaxDimension 单边最大像素数，防止超大目标尺寸导致内存耗尽
const MaxDimension = 10000

// Options 图片处理参数
type Options struct {
	Operation string  // 操作类型，见 Op* 常量；为空时按参数自动推断
	Format    string  // 目标输出格式：jpeg/jpg/png/gif/bmp；为空时沿用源格式
	Quality   int     // JPEG 质量(1-100)，0 表示使用默认值
	Width     int     // 目标宽度（resize）
	Height    int     // 目标高度（resize）
	Scale     float64 // 缩放系数（resize/thumbnail），>0 时优先于宽高
	MaxSize   int     // 缩略图边界框（最长边像素，thumbnail）
}

// Result 处理结果
type Result struct {
	Data   []byte // 处理后的图片字节
	Format string // 实际输出格式（jpeg/png/gif/bmp）
	Width  int    // 输出宽度
	Height int    // 输出高度
}

// SupportedInputFormats 列出可解码的输入格式
func SupportedInputFormats() []string {
	return []string{"jpeg", "png", "gif", "bmp", "webp"}
}

// SupportedOutputFormats 列出可编码的输出格式。
// 注意：webp 在纯 Go 下仅支持解码、不支持编码，故不在输出列表中；
// 如需 webp 输出请使用外置扩展或带 cgo 的编码器。
func SupportedOutputFormats() []string {
	return []string{"jpeg", "png", "gif", "bmp"}
}

// normalizeFormat 归一化格式名
func normalizeFormat(f string) string {
	f = strings.ToLower(strings.TrimSpace(f))
	if f == "jpg" {
		return "jpeg"
	}
	return f
}

// Process 按 Options 处理图片字节，返回处理结果。
func Process(data []byte, opts Options) (*Result, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("空的图片数据")
	}

	src, srcFormat, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %w", err)
	}

	// 决定输出格式
	outFormat := normalizeFormat(opts.Format)
	if outFormat == "" {
		outFormat = normalizeFormat(srcFormat)
	}
	if !isEncodable(outFormat) {
		return nil, fmt.Errorf("不支持的输出格式: %s（本地可编码格式：%s）",
			opts.Format, strings.Join(SupportedOutputFormats(), "/"))
	}

	// 计算目标尺寸
	dstW, dstH, err := targetSize(src.Bounds().Dx(), src.Bounds().Dy(), opts)
	if err != nil {
		return nil, err
	}

	// 若尺寸有变化则做高质量重采样
	out := src
	if dstW != src.Bounds().Dx() || dstH != src.Bounds().Dy() {
		out = resample(src, dstW, dstH)
	}

	encoded, err := encode(out, outFormat, opts.Quality)
	if err != nil {
		return nil, err
	}

	return &Result{
		Data:   encoded,
		Format: outFormat,
		Width:  out.Bounds().Dx(),
		Height: out.Bounds().Dy(),
	}, nil
}

// targetSize 根据操作与参数计算目标宽高
func targetSize(srcW, srcH int, opts Options) (int, int, error) {
	switch opts.Operation {
	case OpThumbnail:
		max := opts.MaxSize
		if opts.Scale > 0 {
			return scaledSize(srcW, srcH, opts.Scale)
		}
		if max <= 0 {
			max = 256
		}
		w, h := fitWithin(srcW, srcH, max)
		return w, h, nil
	case OpResize:
		return explicitResize(srcW, srcH, opts)
	case OpCompress, OpConvert, "":
		// 不改变尺寸；但若显式给了缩放/宽高也尊重之
		if opts.Scale > 0 || opts.Width > 0 || opts.Height > 0 {
			return explicitResize(srcW, srcH, opts)
		}
		return srcW, srcH, nil
	default:
		return 0, 0, fmt.Errorf("未知的操作类型: %s", opts.Operation)
	}
}

// explicitResize 处理 scale / width / height 组合
func explicitResize(srcW, srcH int, opts Options) (int, int, error) {
	if opts.Scale > 0 {
		return scaledSize(srcW, srcH, opts.Scale)
	}
	w, h := opts.Width, opts.Height
	switch {
	case w > 0 && h > 0:
		// 使用指定宽高
	case w > 0:
		h = int(float64(srcH) * float64(w) / float64(srcW))
	case h > 0:
		w = int(float64(srcW) * float64(h) / float64(srcH))
	default:
		return srcW, srcH, nil
	}
	return clampDim(w), clampDim(h), nil
}

func scaledSize(srcW, srcH int, scale float64) (int, int, error) {
	if scale <= 0 {
		return 0, 0, fmt.Errorf("缩放系数必须大于 0")
	}
	w := int(float64(srcW) * scale)
	h := int(float64(srcH) * scale)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	if w > MaxDimension || h > MaxDimension {
		return 0, 0, fmt.Errorf("目标尺寸 %dx%d 超过单边上限 %d", w, h, MaxDimension)
	}
	return w, h, nil
}

// fitWithin 在保持宽高比的前提下，将图片缩放到 max×max 边界框内（仅缩小或保持）
func fitWithin(srcW, srcH, max int) (int, int) {
	if srcW <= max && srcH <= max {
		return srcW, srcH
	}
	if srcW >= srcH {
		h := int(float64(srcH) * float64(max) / float64(srcW))
		if h < 1 {
			h = 1
		}
		return max, h
	}
	w := int(float64(srcW) * float64(max) / float64(srcH))
	if w < 1 {
		w = 1
	}
	return w, max
}

func clampDim(v int) int {
	if v < 1 {
		return 1
	}
	if v > MaxDimension {
		return MaxDimension
	}
	return v
}

// resample 使用 CatmullRom（近似双三次）做高质量重采样，适合放大/缩小
func resample(src image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

func isEncodable(format string) bool {
	for _, f := range SupportedOutputFormats() {
		if f == format {
			return true
		}
	}
	return false
}

// encode 将图片编码为指定格式
func encode(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		q := quality
		if q <= 0 || q > 100 {
			q = DefaultJPEGQuality
		}
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
			return nil, fmt.Errorf("JPEG 编码失败: %w", err)
		}
	case "png":
		enc := png.Encoder{CompressionLevel: png.DefaultCompression}
		if err := enc.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("PNG 编码失败: %w", err)
		}
	case "gif":
		if err := gif.Encode(&buf, img, &gif.Options{NumColors: 256}); err != nil {
			return nil, fmt.Errorf("GIF 编码失败: %w", err)
		}
	case "bmp":
		if err := xbmp.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("BMP 编码失败: %w", err)
		}
	default:
		return nil, fmt.Errorf("不支持的输出格式: %s", format)
	}
	return buf.Bytes(), nil
}

// MimeTypeForFormat 返回格式对应的 MIME 类型
func MimeTypeForFormat(format string) string {
	switch normalizeFormat(format) {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "bmp":
		return "image/bmp"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
