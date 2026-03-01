package ginsrv

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/bizutil/errcode"
	"github.com/qq1060656096/drugo-provider/pkg/ginresp"
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
			if r := recover(); r != nil {
				var err error
				switch e := r.(type) {
				case error:
					err = e
				default:
					err = fmt.Errorf("%v", r)
				}
				// ⭐ 获取 trace_id
				traceID := GetTraceID(c)

				errorLogger.Error("panic recovered",
					zap.Any("recoverData", r),
					zap.ByteString("stack", debug.Stack()),
					zap.String("trace_id", traceID), // ⭐ 新增
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)

				err = errcode.Wrap(1500010001, err, "internal server error")
				ginresp.AbortErr(c, err, nil)
				return
			}
		}()

		c.Next()
	}
}
