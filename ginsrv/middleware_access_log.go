package ginsrv

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	maxBodySize          = 4 * 1024 // 最大记录4KB的body
	defaultAccessLogName = "gin.access"
)

// responseWriter 用于捕获响应的body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.body.Len() < maxBodySize {
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// AccessLogger 是用于记录请求、响应日志的中间件
func AccessLogger(lmg interface{ MustGet(string) *zap.Logger }, accessLogName string, errLogName string) gin.HandlerFunc {
	if accessLogName == "" {
		accessLogName = defaultAccessLogName
	}
	if errLogName == "" {
		errLogName = defaultErrorLogName
	}

	accessLogger := lmg.MustGet(accessLogName)
	errorLogger := lmg.MustGet(errLogName)

	return func(c *gin.Context) {
		start := time.Now()

		// ⭐ 获取 trace_id
		traceID := GetTraceID(c)

		// 读取请求body
		var requestBody []byte
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize))
			requestBody = bodyBytes
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 重新设置请求body
		}

		// 替换响应Writer以捕获响应body
		bw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = bw

		// 处理请求
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := getClientIP(c)

		// ⭐ 统一日志字段（加入 trace_id）
		fields := []zap.Field{
			zap.String("trace_id", traceID), // ⭐ 新增
			zap.Int("status", statusCode),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", clientIP),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			zap.Int("size", c.Writer.Size()),
			zap.ByteString("request", requestBody),
			zap.ByteString("response", bw.body.Bytes()),
		}

		// 处理业务错误
		if len(c.Errors) > 0 {
			errorLogger.Error("request error", append(fields, zap.String("errors", c.Errors.String()))...)
			return
		}

		// 处理HTTP错误
		if statusCode >= http.StatusInternalServerError {
			errorLogger.Error("server error", fields...)
			return
		} else if statusCode >= http.StatusBadRequest {
			errorLogger.Warn("client error", fields...)
			return
		}

		// 正常访问日志
		accessLogger.Info("request success", fields...)
	}
}
