package ginresp

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/bizutil/eresp"
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

			Fail(c, tt.code, tt.msg, nil)

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

			Err(c, tt.err, nil)

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

	Err(c, err, nil)

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
		AbortFail(ginCtx, 1001, "test error", nil)
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
		AbortErr(ginCtx, err, nil)
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

func TestFail_WithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		code     int
		msg      string
		details  any
		expected int
	}{
		{
			name:     "with string details",
			code:     1001,
			msg:      "invalid param",
			details:  "field name is required",
			expected: http.StatusOK,
		},
		{
			name:     "with map details",
			code:     1002,
			msg:      "validation failed",
			details:  map[string]string{"field": "email", "error": "invalid format"},
			expected: http.StatusOK,
		},
		{
			name:     "with slice details",
			code:     1003,
			msg:      "multiple errors",
			details:  []string{"error1", "error2"},
			expected: http.StatusOK,
		},
		{
			name:     "with nil details",
			code:     1004,
			msg:      "no details",
			details:  nil,
			expected: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Fail(c, tt.code, tt.msg, tt.details)

			assert.Equal(t, tt.expected, w.Code)
			assert.Contains(t, w.Body.String(), `"code":`+fmt.Sprintf("%d", tt.code))
			assert.Contains(t, w.Body.String(), `"message":"`+tt.msg+`"`)

			if tt.details != nil {
				assert.Contains(t, w.Body.String(), `"details":`)
			} else {
				assert.NotContains(t, w.Body.String(), `"details":`)
			}
		})
	}
}

func TestErr_WithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		err      error
		details  any
		expected int
	}{
		{
			name:     "generic error with string details",
			err:      errors.New("generic error"),
			details:  "additional context",
			expected: http.StatusInternalServerError,
		},
		{
			name:     "error code with map details",
			err:      errcode.New(1014000001, "test error"),
			details:  map[string]any{"field": "username", "value": "invalid"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "nil error with details",
			err:      nil,
			details:  "fallback details",
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Err(c, tt.err, tt.details)

			assert.Equal(t, tt.expected, w.Code)
			assert.Contains(t, w.Body.String(), `"code":`)

			if tt.details != nil && tt.err != nil {
				assert.Contains(t, w.Body.String(), `"details":`)
			}
		})
	}
}

func TestAbortFail_WithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()

	// 创建一个 gin 引擎来测试 abort 功能
	r := gin.New()
	r.Use(func(ginCtx *gin.Context) {
		ginCtx.Next()
		assert.True(t, ginCtx.IsAborted())
	})
	r.GET("/test", func(ginCtx *gin.Context) {
		details := map[string]string{"field": "email", "error": "invalid format"}
		AbortFail(ginCtx, 1001, "validation error", details)
	})

	// 模拟请求
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":1001`)
	assert.Contains(t, w.Body.String(), `"details":`)
}

func TestAbortErr_WithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()

	// 创建一个 gin 引擎来测试 abort 功能
	r := gin.New()
	r.Use(func(ginCtx *gin.Context) {
		ginCtx.Next()
		assert.True(t, ginCtx.IsAborted())
	})
	r.GET("/test", func(ginCtx *gin.Context) {
		err := errcode.New(1014000001, "test error")
		details := "error context"
		AbortErr(ginCtx, err, details)
	})

	// 模拟请求
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"code":1014000001`)
	assert.Contains(t, w.Body.String(), `"details":`)
}

func TestResolveStatus_MoreErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "401 unauthorized",
			err:      errcode.New(1014010001, "unauthorized"), // 1(占位符) + 01(模块) + 401(HTTP状态) + 0001(顺序)
			expected: http.StatusUnauthorized,
		},
		{
			name:     "403 forbidden",
			err:      errcode.New(1014030001, "forbidden"), // 1(占位符) + 01(模块) + 403(HTTP状态) + 0001(顺序)
			expected: http.StatusForbidden,
		},
		{
			name:     "404 not found",
			err:      errcode.New(1014040001, "not found"), // 1(占位符) + 01(模块) + 404(HTTP状态) + 0001(顺序)
			expected: http.StatusNotFound,
		},
		{
			name:     "422 unprocessable entity",
			err:      errcode.New(1014220001, "unprocessable"), // 1(占位符) + 01(模块) + 422(HTTP状态) + 0001(顺序)
			expected: http.StatusUnprocessableEntity,
		},
		{
			name:     "429 too many requests",
			err:      errcode.New(1014290001, "rate limit"), // 1(占位符) + 01(模块) + 429(HTTP状态) + 0001(顺序)
			expected: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveStatus(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrite_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("nil context panic", func(t *testing.T) {
		// 这个测试需要特殊处理，因为 write 需要 non-nil context
		assert.Panics(t, func() {
			write(nil, http.StatusOK, eresp.OKResp("test", ""))
		})
	})

	t.Run("empty trace ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(TraceIDKey, "")

		OK(c, "test data")
		assert.NotContains(t, w.Body.String(), "trace_id")
	})

	t.Run("non-string trace ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(TraceIDKey, 12345)

		OK(c, "test data")
		assert.NotContains(t, w.Body.String(), "trace_id")
	})
}

func TestOKMsg_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		data     any
		msg      string
		expected int
	}{
		{
			name:     "complex data structure",
			data:     map[string]any{"user": map[string]string{"name": "test", "email": "test@example.com"}, "roles": []string{"admin", "user"}},
			msg:      "complex data",
			expected: http.StatusOK,
		},
		{
			name:     "empty message with complex data",
			data:     []int{1, 2, 3, 4, 5},
			msg:      "",
			expected: http.StatusOK,
		},
		{
			name:     "nil data with message",
			data:     nil,
			msg:      "no data available",
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
		Fail(c, 1001, "benchmark error", nil)
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
		Err(c, err, nil)
		w.Body.Reset()
	}
}
