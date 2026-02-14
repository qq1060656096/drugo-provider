package ginsrv

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTraceMiddleware_DefaultTraceKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建 gin 引擎并添加中间件
	r := gin.New()
	r.Use(TraceMiddleware(""))
	r.GET("/test", func(c *gin.Context) {
		traceID := GetTraceID(c)
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	// 创建请求
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// 执行请求
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "trace_id")
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestTraceMiddleware_CustomTraceKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	customKey := "X-Custom-Trace-ID"
	r := gin.New()
	r.Use(TraceMiddleware(customKey))
	r.GET("/test", func(c *gin.Context) {
		traceID := GetTraceID(c)
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "trace_id")
	assert.NotEmpty(t, w.Header().Get(customKey))
	assert.Empty(t, w.Header().Get("X-Request-ID"))
}

func TestTraceMiddleware_WithExistingTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	existingTraceID := "existing-trace-123"
	r := gin.New()
	r.Use(TraceMiddleware("X-Request-ID"))
	r.GET("/test", func(c *gin.Context) {
		traceID := GetTraceID(c)
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingTraceID)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), existingTraceID)
	assert.Equal(t, existingTraceID, w.Header().Get("X-Request-ID"))
}

func TestTraceMiddleware_ContextValue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(TraceMiddleware(""))
	r.GET("/test", func(c *gin.Context) {
		// 检查 context 中的 trace ID
		traceID := c.Request.Context().Value(TraceIDKey)
		c.JSON(200, gin.H{"context_trace_id": traceID})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "context_trace_id")
}

func TestGetTraceID_WhenNotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)

	traceID := GetTraceID(c)
	assert.Empty(t, traceID)
}

func TestGetTraceID_WhenSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)

	expectedTraceID := "test-trace-id"
	c.Set(TraceIDKey, expectedTraceID)

	traceID := GetTraceID(c)
	assert.Equal(t, expectedTraceID, traceID)
}

func TestTraceMiddleware_ChainRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(TraceMiddleware(""))
	r.Use(func(c *gin.Context) {
		// 在后续中间件中验证 trace ID 存在
		traceID := GetTraceID(c)
		assert.NotEmpty(t, traceID)

		// 验证 context 中也有 trace ID
		ctxTraceID := c.Request.Context().Value(TraceIDKey)
		assert.NotEmpty(t, ctxTraceID)

		c.Next()
	})
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestTraceMiddleware_DifferentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(TraceMiddleware(""))
	r.GET("/test", func(c *gin.Context) {
		traceID := GetTraceID(c)
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	// 第一个请求
	req1, _ := http.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// 第二个请求
	req2, _ := http.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	// 验证两个请求有不同的 trace ID
	assert.NotEmpty(t, w1.Header().Get("X-Request-ID"))
	assert.NotEmpty(t, w2.Header().Get("X-Request-ID"))
	assert.NotEqual(t, w1.Header().Get("X-Request-ID"), w2.Header().Get("X-Request-ID"))
}
