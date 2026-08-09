// loadtest Slink 压测工具
//
// 支持两种模式:
//   upload  multipart 上传(每个请求内容唯一,避免 MD5 去重影响测量)
//   get     普通 GET(图片访问 / 列表接口等)
//
// 用法:
//   go run ./tools/loadtest -mode upload -url http://127.0.0.1:8080/api/image/upload -token <JWT> -n 300 -c 10 -size 64
//   go run ./tools/loadtest -mode get    -url http://127.0.0.1:8080/static/2026/08/09/xxx.png   -n 5000 -c 50
//   go run ./tools/loadtest -mode get    -url "http://127.0.0.1:8080/api/image/list?page=1&limit=20" -token <JWT> -n 2000 -c 20
//
// 输出:成功/失败数、总耗时、QPS、p50/p95/p99/max 延迟。
// 服务端资源占用可在压测期间用以下方式采样:
//   while true; do ps -o %cpu=,rss= -p <PID>; sleep 0.3; done
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sort"
	"sync/atomic"
	"time"
)

func main() {
	mode := flag.String("mode", "get", "upload|get")
	url := flag.String("url", "", "target url")
	token := flag.String("token", "", "bearer token")
	n := flag.Int("n", 1000, "total requests")
	c := flag.Int("c", 10, "concurrency")
	sizeKB := flag.Int("size", 64, "upload payload size in KB")
	flag.Parse()

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        512,
			MaxIdleConnsPerHost: 512,
		},
	}

	// 合法 PNG 头 + 填充;upload 模式每个请求尾部写入序号保证内容唯一
	basePNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	payload := make([]byte, *sizeKB*1024)
	copy(payload, basePNG)
	for i := len(basePNG); i < len(payload); i++ {
		payload[i] = byte(i * 31)
	}

	var idx int64
	lat := make([]time.Duration, *n)
	var okCount, errCount int64

	start := time.Now()
	jobs := make(chan int64, *c*2)
	done := make(chan struct{})
	for w := 0; w < *c; w++ {
		go func() {
			for i := range jobs {
				var req *http.Request
				var err error
				if *mode == "upload" {
					body := &bytes.Buffer{}
					mw := multipart.NewWriter(body)
					fw, _ := mw.CreateFormFile("image", fmt.Sprintf("bench-%d.png", i))
					p := append([]byte{}, payload...)
					copy(p[len(p)-8:], []byte(fmt.Sprintf("%08d", i)))
					fw.Write(p)
					mw.Close()
					req, err = http.NewRequest("POST", *url, body)
					req.Header.Set("Content-Type", mw.FormDataContentType())
				} else {
					req, err = http.NewRequest("GET", *url, nil)
				}
				if err == nil {
					if *token != "" {
						req.Header.Set("Authorization", "Bearer "+*token)
					}
					t0 := time.Now()
					resp, err2 := client.Do(req)
					if err2 != nil {
						atomic.AddInt64(&errCount, 1)
					} else {
						io.Copy(io.Discard, resp.Body)
						resp.Body.Close()
						if resp.StatusCode >= 400 {
							atomic.AddInt64(&errCount, 1)
						} else {
							atomic.AddInt64(&okCount, 1)
						}
						lat[i] = time.Since(t0)
					}
				} else {
					atomic.AddInt64(&errCount, 1)
				}
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < *n; i++ {
		jobs <- atomic.AddInt64(&idx, 1) - 1
	}
	close(jobs)
	for w := 0; w < *c; w++ {
		<-done
	}
	elapsed := time.Since(start)

	valid := lat[:0]
	for _, l := range lat {
		if l > 0 {
			valid = append(valid, l)
		}
	}
	sort.Slice(valid, func(a, b int) bool { return valid[a] < valid[b] })
	pct := func(p float64) time.Duration {
		if len(valid) == 0 {
			return 0
		}
		i := int(float64(len(valid)) * p)
		if i >= len(valid) {
			i = len(valid) - 1
		}
		return valid[i]
	}

	fmt.Printf("mode=%s n=%d c=%d ok=%d err=%d elapsed=%s QPS=%.1f\n",
		*mode, *n, *c, okCount, errCount, elapsed.Round(time.Millisecond), float64(*n)/elapsed.Seconds())
	if len(valid) > 0 {
		fmt.Printf("latency: p50=%s p95=%s p99=%s max=%s\n",
			pct(0.50).Round(time.Microsecond), pct(0.95).Round(time.Microsecond),
			pct(0.99).Round(time.Microsecond), valid[len(valid)-1].Round(time.Microsecond))
	}
}
