package ginresp

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/bizutil/errcode"
	"github.com/stretchr/testify/assert"
)

func TestOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		data     any
		expected int
	}{
		{
			name:     "string data",
			data:     "success",
			expected: http.StatusOK,
		},
		{
			name:     "map data",
			data:     map[string]any{"key": "value"},
			expected: http.StatusOK,
		},
		{
			name:     "nil data",
			data:     nil,
			expected: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			OK(c, tt.data)

			assert.Equal(t, tt.expected, w.Code)
			assert.Contains(t, w.Body.String(), `"code":0`)
			assert.Contains(t, w.Body.String(), `"message":"OK"`)
		})
	}
}

func TestOKMsg(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		data     any
		msg      string
		expected int
	}{
		{
			name:     "with message",
			data:     "test",
			msg:      "custom message",
			expected: http.StatusOK,
		},
		{
			name:     "empty message",
			data:     map[string]string{"test": "value"},
			msg:      "",
			expected: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			OKMsg(c, tt.data, tt.msg)

			assert.Equal(t, tt.expected, w.Code)
			assert.Contains(t, w.Body.String(), `"code":0`)
			if tt.msg != "" {
				assert.Contains(t, w.Body.String(), `"message":"`+tt.msg+`"`)
			}
		})
	}
}

func TestFail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		code     int
		msg      string
		expected int
	}{
		{
			name:     "business error",
			code:     1001,
			msg:      "invalid param",
			expected: http.StatusOK,
		},
		{
			name:     "empty message",
			code:     2001,
			msg:      "",
			expected: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Fail(c, tt.code, tt.msg)

			assert.Equal(t, tt.expected, w.Code)
			assert.Contains(t, w.Body.String(), `"code":`+fmt.Sprintf("%d", tt.code))
			if tt.msg != "" {
				assert.Contains(t, w.Body.String(), `"message":"`+tt.msg+`"`)
			}
		})
	}
}

func TestErr(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Err(c, tt.err)

			assert.Equal(t, tt.expected, w.Code)
			assert.Contains(t, w.Body.String(), `"code":`)
		})
	}
}

func TestErr_WithErrorCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建一个带有 HTTP 状态码的错误
	// 使用 9 位错误码：1(占位符) + 01(模块) + 400(HTTP状态) + 0001(顺序)
	err := errcode.New(1014000001, "test error")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Err(c, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"code":1014000001`)
	assert.Contains(t, w.Body.String(), `"message":"test error"`)
}

func TestAbortOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()

	// 创建一个 gin 引擎来测试 abort 功能
	r := gin.New()
	r.Use(func(ginCtx *gin.Context) {
		ginCtx.Next()
		assert.True(t, ginCtx.IsAborted())
	})
	r.GET("/test", func(ginCtx *gin.Context) {
		AbortOK(ginCtx, "test data")
	})

	// 模拟请求
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
}

func TestAbortFail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()

	// 创建一个 gin 引擎来测试 abort 功能
	r := gin.New()
	r.Use(func(ginCtx *gin.Context) {
		ginCtx.Next()
		assert.True(t, ginCtx.IsAborted())
	})
	r.GET("/test", func(ginCtx *gin.Context) {
		AbortFail(ginCtx, 1001, "test error")
	})

	// 模拟请求
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":1001`)
}

func TestAbortErr(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()

	// 创建一个 gin 引擎来测试 abort 功能
	r := gin.New()
	r.Use(func(ginCtx *gin.Context) {
		ginCtx.Next()
		assert.True(t, ginCtx.IsAborted())
	})
	r.GET("/test", func(ginCtx *gin.Context) {
		err := errors.New("test error")
		AbortErr(ginCtx, err)
	})

	// 模拟请求
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		setup    func(*gin.Context)
		expected string
	}{
		{
			name:     "nil context",
			setup:    func(c *gin.Context) {},
			expected: "",
		},
		{
			name: "no trace id",
			setup: func(c *gin.Context) {
				// 不设置任何值
			},
			expected: "",
		},
		{
			name: "trace id exists",
			setup: func(c *gin.Context) {
				c.Set(TraceIDKey, "test-trace-id")
			},
			expected: "test-trace-id",
		},
		{
			name: "trace id exists but not string",
			setup: func(c *gin.Context) {
				c.Set(TraceIDKey, 12345)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "nil context" {
				result := getTraceID(nil)
				assert.Equal(t, tt.expected, result)
			} else {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				tt.setup(c)

				result := getTraceID(c)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestResolveStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: http.StatusInternalServerError,
		},
		{
			name:     "error code with custom status",
			err:      errcode.New(1014000001, "test"), // 1(占位符) + 01(模块) + 400(HTTP状态) + 0001(顺序)
			expected: http.StatusBadRequest,
		},
		{
			name:     "error code with default status",
			err:      errcode.New(1015000001, "test"), // 1(占位符) + 01(模块) + 500(HTTP状态) + 0001(顺序)
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveStatus(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrite_WithTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(TraceIDKey, "test-trace-123")

	// 测试 write 函数是否能正确处理 trace ID
	OK(c, "test data")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"trace_id":"test-trace-123"`)
}

func TestWrite_WithoutTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// 不设置 trace ID

	OK(c, "test data")

	assert.Equal(t, http.StatusOK, w.Code)
	// 应该不包含 trace_id 字段
	assert.NotContains(t, w.Body.String(), "trace_id")
}

// 基准测试
func BenchmarkOK(b *testing.B) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OK(c, "benchmark data")
		w.Body.Reset()
	}
}

func BenchmarkFail(b *testing.B) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Fail(c, 1001, "benchmark error")
		w.Body.Reset()
	}
}

func BenchmarkErr(b *testing.B) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	err := errors.New("benchmark error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Err(c, err)
		w.Body.Reset()
	}
}
