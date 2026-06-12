package middleware

import (
	"Slink/applog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ctxRequestID = "request_id"

// RequestID 返回当前请求的 X-Request-ID / 内部 ID（若中间件已设置）。
func RequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// RequestLogger 记录每个 HTTP 请求的详细信息（方法、路径、状态、耗时、客户端、用户 ID 等）。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(ctxRequestID, rid)
		c.Writer.Header().Set("X-Request-ID", rid)

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		applog.Logger.Debug("http request start",
			"request_id", rid,
			"method", c.Request.Method,
			"path", path,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"content_length", c.Request.ContentLength,
		)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		uidVal, hasUser := c.Get("userID")
		logArgs := []any{
			"request_id", rid,
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"bytes_written", c.Writer.Size(),
		}
		if hasUser {
			logArgs = append(logArgs, "user_id", uidVal)
		}
		if len(c.Errors) > 0 {
			logArgs = append(logArgs, "gin_errors", c.Errors.String())
		}

		if status >= 500 {
			applog.Logger.Error("http request done", logArgs...)
		} else if status >= 400 {
			applog.Logger.Warn("http request done", logArgs...)
		} else {
			applog.Logger.Info("http request done", logArgs...)
		}
	}
}
