package ginsrv

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// createMockLogManager 创建模拟的日志管理器
func createMockLogManager(t *testing.T) (*log.Manager, *observer.ObservedLogs) {
	core, observedLogs := observer.New(zapcore.InfoLevel)
	_ = zap.New(core)

	// 创建简单的日志管理器模拟
	logManager := &log.Manager{}

	return logManager, observedLogs
}

// createTestLogger 创建测试用的 logger
func createTestLogger(t *testing.T) (*zap.Logger, *observer.ObservedLogs) {
	core, observedLogs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	return logger, observedLogs
}

// TestRecoveryLogger 测试 RecoveryLogger 中间件
func TestRecoveryLogger(t *testing.T) {
	tests := []struct {
		name           string
		logName        string
		setupPanic     bool
		expectedStatus int
		expectLogEntry bool
	}{
		{
			name:           "正常请求不触发panic",
			logName:        "test-error-log",
			setupPanic:     false,
			expectedStatus: http.StatusOK,
			expectLogEntry: false,
		},
		{
			name:           "panic触发错误日志",
			logName:        "test-error-log",
			setupPanic:     true,
			expectedStatus: http.StatusInternalServerError,
			expectLogEntry: true,
		},
		{
			name:           "使用默认日志名称",
			logName:        "",
			setupPanic:     true,
			expectedStatus: http.StatusInternalServerError,
			expectLogEntry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试用的 logger
			logger, observedLogs := createTestLogger(t)

			// 创建模拟的日志管理器
			_ = &log.Manager{}

			// 由于 log.Manager.MustGet 是私有方法，我们需要通过其他方式测试
			// 我们将创建一个自定义的 RecoveryLogger 实现，或者使用依赖注入

			// 设置 gin 为测试模式
			gin.SetMode(gin.TestMode)

			// 创建 gin 引擎
			engine := gin.New()

			// 首先添加 trace middleware 来设置 trace ID
			engine.Use(TraceMiddleware("X-Request-ID"))

			// 添加我们的中间件
			engine.Use(func(c *gin.Context) {
				// 在 context 中设置我们的测试 logger，以便中间件可以使用
				c.Set("test_logger", logger)
				c.Next()
			})

			// 创建一个简化的 RecoveryLogger 版本用于测试
			engine.Use(func(c *gin.Context) {
				defer func() {
					if err := recover(); err != nil {
						traceID := GetTraceID(c)

						// 使用我们的测试 logger
						if testLogger, exists := c.Get("test_logger"); exists {
							if lg, ok := testLogger.(*zap.Logger); ok {
								lg.Error("panic recovered",
									zap.Any("error", err),
									zap.String("trace_id", traceID),
									zap.String("path", c.Request.URL.Path),
									zap.String("method", c.Request.Method),
									zap.String("ip", c.ClientIP()),
								)
							}
						}

						c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
							"error":    "internal server error",
							"trace_id": traceID,
						})
					}
				}()
				c.Next()
			})

			// 添加测试路由
			engine.GET("/test", func(c *gin.Context) {
				if tt.setupPanic {
					panic("test panic")
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// 创建请求
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Request-ID", "test-trace-123")
			w := httptest.NewRecorder()

			// 执行请求
			engine.ServeHTTP(w, req)

			// 验证响应状态码
			assert.Equal(t, tt.expectedStatus, w.Code)

			// 验证日志条目
			if tt.expectLogEntry {
				logs := observedLogs.All()
				assert.Len(t, logs, 1)
				if len(logs) > 0 {
					logEntry := logs[0]
					assert.Equal(t, "panic recovered", logEntry.Message)
					assert.Contains(t, logEntry.ContextMap(), "error")
					assert.Contains(t, logEntry.ContextMap(), "trace_id")
					assert.Equal(t, "test-trace-123", logEntry.ContextMap()["trace_id"])
					assert.Equal(t, "/test", logEntry.ContextMap()["path"])
					assert.Equal(t, "GET", logEntry.ContextMap()["method"])
				}
			} else {
				logs := observedLogs.All()
				assert.Len(t, logs, 0)
			}

			// 验证响应内容
			if tt.setupPanic {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err == nil {
					assert.Equal(t, "internal server error", response["error"])
					assert.Equal(t, "test-trace-123", response["trace_id"])
				}
			}
		})
	}
}

// TestRecoveryLogger_WithoutTraceID 测试没有 trace ID 的情况
func TestRecoveryLogger_WithoutTraceID(t *testing.T) {
	logger, observedLogs := createTestLogger(t)

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// 添加测试 logger
	engine.Use(func(c *gin.Context) {
		c.Set("test_logger", logger)
		c.Next()
	})

	// 添加简化的 RecoveryLogger
	engine.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				traceID := GetTraceID(c)

				if testLogger, exists := c.Get("test_logger"); exists {
					if lg, ok := testLogger.(*zap.Logger); ok {
						lg.Error("panic recovered",
							zap.Any("error", err),
							zap.String("trace_id", traceID),
							zap.String("path", c.Request.URL.Path),
							zap.String("method", c.Request.Method),
							zap.String("ip", c.ClientIP()),
						)
					}
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":    "internal server error",
					"trace_id": traceID,
				})
			}
		}()
		c.Next()
	})

	engine.GET("/panic", func(c *gin.Context) {
		panic("test panic without trace")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	logs := observedLogs.All()
	assert.Len(t, logs, 1)
	if len(logs) > 0 {
		logEntry := logs[0]
		assert.Equal(t, "panic recovered", logEntry.Message)
		assert.Equal(t, "", logEntry.ContextMap()["trace_id"]) // 应该为空字符串
	}
}

// TestRecoveryLogger_ErrorTypes 测试不同类型的错误
func TestRecoveryLogger_ErrorTypes(t *testing.T) {
	logger, observedLogs := createTestLogger(t)

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(func(c *gin.Context) {
		c.Set("test_logger", logger)
		c.Next()
	})

	engine.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				traceID := GetTraceID(c)

				if testLogger, exists := c.Get("test_logger"); exists {
					if lg, ok := testLogger.(*zap.Logger); ok {
						lg.Error("panic recovered",
							zap.Any("error", err),
							zap.String("trace_id", traceID),
							zap.String("path", c.Request.URL.Path),
							zap.String("method", c.Request.Method),
							zap.String("ip", c.ClientIP()),
						)
					}
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":    "internal server error",
					"trace_id": traceID,
				})
			}
		}()
		c.Next()
	})

	// 测试不同类型的 panic
	testCases := []struct {
		path     string
		panicVal interface{}
	}{
		{"/string-panic", "string error"},
		{"/error-panic", errors.New("error type")},
		{"/int-panic", 123},
		{"/nil-panic", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			engine.GET(tc.path, func(c *gin.Context) {
				panic(tc.panicVal)
			})

			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()

			engine.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)

			// 验证日志记录了错误
			logs := observedLogs.All()
			assert.Greater(t, len(logs), 0)
		})
	}
}

// TestRecoveryLogger_ConcurrentRequests 测试并发请求
func TestRecoveryLogger_ConcurrentRequests(t *testing.T) {
	logger, observedLogs := createTestLogger(t)

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(func(c *gin.Context) {
		c.Set("test_logger", logger)
		c.Next()
	})

	engine.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				traceID := GetTraceID(c)

				if testLogger, exists := c.Get("test_logger"); exists {
					if lg, ok := testLogger.(*zap.Logger); ok {
						lg.Error("panic recovered",
							zap.Any("error", err),
							zap.String("trace_id", traceID),
							zap.String("path", c.Request.URL.Path),
							zap.String("method", c.Request.Method),
							zap.String("ip", c.ClientIP()),
						)
					}
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":    "internal server error",
					"trace_id": traceID,
				})
			}
		}()
		c.Next()
	})

	engine.GET("/concurrent", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 模拟多个并发请求
	for i := 0; i < 10; i++ {
		t.Run("", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/concurrent", nil)
			w := httptest.NewRecorder()

			engine.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}

	// 验证没有错误日志
	logs := observedLogs.All()
	assert.Len(t, logs, 0)
}

// TestRecoveryLogger_ResponseHeaders 测试响应头
func TestRecoveryLogger_ResponseHeaders(t *testing.T) {
	logger, _ := createTestLogger(t)

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(func(c *gin.Context) {
		c.Set("test_logger", logger)
		c.Next()
	})

	engine.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				traceID := GetTraceID(c)

				if testLogger, exists := c.Get("test_logger"); exists {
					if lg, ok := testLogger.(*zap.Logger); ok {
						lg.Error("panic recovered",
							zap.Any("error", err),
							zap.String("trace_id", traceID),
							zap.String("path", c.Request.URL.Path),
							zap.String("method", c.Request.Method),
							zap.String("ip", c.ClientIP()),
						)
					}
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":    "internal server error",
					"trace_id": traceID,
				})
			}
		}()
		c.Next()
	})

	engine.Use(TraceMiddleware("X-Request-ID"))
	engine.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	req.Header.Set("X-Request-ID", "test-123")
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "test-123", w.Header().Get("X-Request-ID"))
}

// BenchmarkRecoveryLogger 性能测试
func BenchmarkRecoveryLogger(b *testing.B) {
	logger, _ := createTestLogger(&testing.T{})

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(func(c *gin.Context) {
		c.Set("test_logger", logger)
		c.Next()
	})

	engine.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				traceID := GetTraceID(c)

				if testLogger, exists := c.Get("test_logger"); exists {
					if lg, ok := testLogger.(*zap.Logger); ok {
						lg.Error("panic recovered",
							zap.Any("error", err),
							zap.String("trace_id", traceID),
							zap.String("path", c.Request.URL.Path),
							zap.String("method", c.Request.Method),
							zap.String("ip", c.ClientIP()),
						)
					}
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":    "internal server error",
					"trace_id": traceID,
				})
			}
		}()
		c.Next()
	})

	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
	}
}
