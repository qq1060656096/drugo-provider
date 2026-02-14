package ginsrv

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const TraceIDKey = "trace_id"

func TraceMiddleware(traceKey string) gin.HandlerFunc {
	if traceKey == "" {
		traceKey = "X-Request-ID"
	}

	return func(c *gin.Context) {
		traceID := c.GetHeader(traceKey)
		if traceID == "" {
			traceID = uuid.NewString()
		}

		// gin context
		c.Set(TraceIDKey, traceID)

		// request context（关键）
		ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		// response header
		c.Writer.Header().Set(traceKey, traceID)

		c.Next()
	}
}

func GetTraceID(c *gin.Context) string {
	v, ok := c.Get(TraceIDKey)
	if !ok {
		return ""
	}

	s, _ := v.(string)
	return s
}
