package imaging

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// makePNG 生成一张纯色测试 PNG
func makePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func decodeDims(t *testing.T, data []byte) (int, int, string) {
	t.Helper()
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode result: %v", err)
	}
	return cfg.Width, cfg.Height, format
}

func TestResizeByWidthKeepsAspect(t *testing.T) {
	src := makePNG(t, 200, 100)
	res, err := Process(src, Options{Operation: OpResize, Width: 100})
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	w, h, _ := decodeDims(t, res.Data)
	if w != 100 || h != 50 {
		t.Fatalf("expected 100x50, got %dx%d", w, h)
	}
}

func TestUpscaleByScale(t *testing.T) {
	src := makePNG(t, 50, 40)
	res, err := Process(src, Options{Operation: OpResize, Scale: 2})
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if res.Width != 100 || res.Height != 80 {
		t.Fatalf("expected 100x80, got %dx%d", res.Width, res.Height)
	}
}

func TestConvertPNGToJPEG(t *testing.T) {
	src := makePNG(t, 64, 64)
	res, err := Process(src, Options{Operation: OpConvert, Format: "jpg", Quality: 80})
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if res.Format != "jpeg" {
		t.Fatalf("expected jpeg, got %s", res.Format)
	}
	_, _, format := decodeDims(t, res.Data)
	if format != "jpeg" {
		t.Fatalf("decoded format expected jpeg, got %s", format)
	}
}

func TestThumbnailFitsWithinBox(t *testing.T) {
	src := makePNG(t, 800, 400)
	res, err := Process(src, Options{Operation: OpThumbnail, MaxSize: 200})
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if res.Width != 200 || res.Height != 100 {
		t.Fatalf("expected 200x100, got %dx%d", res.Width, res.Height)
	}
}

func TestThumbnailDoesNotUpscale(t *testing.T) {
	src := makePNG(t, 100, 100)
	res, err := Process(src, Options{Operation: OpThumbnail, MaxSize: 500})
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if res.Width != 100 || res.Height != 100 {
		t.Fatalf("expected unchanged 100x100, got %dx%d", res.Width, res.Height)
	}
}

func TestUnsupportedOutputFormatRejected(t *testing.T) {
	src := makePNG(t, 32, 32)
	if _, err := Process(src, Options{Operation: OpConvert, Format: "webp"}); err == nil {
		t.Fatal("expected error for webp output, got nil")
	}
}

func TestOversizedScaleRejected(t *testing.T) {
	src := makePNG(t, 4000, 4000)
	if _, err := Process(src, Options{Operation: OpResize, Scale: 5}); err == nil {
		t.Fatal("expected error for oversized target, got nil")
	}
}
