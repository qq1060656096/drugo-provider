package ginsrv

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo/log"
	"go.uber.org/zap"
)

const (
	defaultErrorLogName = "gin.error"
)

// RecoveryLogger 捕获panic并记录错误日志
func RecoveryLogger(lmg *log.Manager, logName string) gin.HandlerFunc {
	if logName == "" {
		logName = defaultErrorLogName
	}
	errorLogger := lmg.MustGet(logName)

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

				// ⭐ 获取 trace_id
				traceID := GetTraceID(c)

				errorLogger.Error("panic recovered",
					zap.Any("error", err),
					zap.ByteString("stack", debug.Stack()),
					zap.String("trace_id", traceID), // ⭐ 新增
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":    "internal server error",
					"trace_id": traceID, // ⭐ 建议返回给客户端
				})
			}
		}()

		c.Next()
	}
}
