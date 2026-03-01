package ginsrv

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() *gin.Context
		expectedIP   string
	}{
		{
			name: "ClientIP available",
			setupContext: func() *gin.Context {
				gin.SetMode(gin.TestMode)
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set("X-Forwarded-For", "192.168.1.100")
				return c
			},
			expectedIP: "192.168.1.100",
		},
		{
			name: "ClientIP empty, use RemoteAddr",
			setupContext: func() *gin.Context {
				gin.SetMode(gin.TestMode)
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.RemoteAddr = "192.168.1.200:8080"
				return c
			},
			expectedIP: "192.168.1.200",
		},
		{
			name: "ClientIP empty, RemoteAddr invalid",
			setupContext: func() *gin.Context {
				gin.SetMode(gin.TestMode)
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.RemoteAddr = "invalid-addr"
				return c
			},
			expectedIP: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setupContext()
			ip := getClientIP(c)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func TestGetVar(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	tests := []struct {
		name       string
		setupFunc  func(*gin.Context)
		key        string
		wantValue  interface{}
		wantExists bool
	}{
		{
			name: "key exists with correct type",
			setupFunc: func(c *gin.Context) {
				c.Set("user", &User{Name: "Alice", Age: 25})
			},
			key:        "user",
			wantValue:  &User{Name: "Alice", Age: 25},
			wantExists: true,
		},
		{
			name: "key exists with wrong type",
			setupFunc: func(c *gin.Context) {
				c.Set("user", "string value")
			},
			key:        "user",
			wantValue:  (*User)(nil),
			wantExists: false,
		},
		{
			name:       "key does not exist",
			setupFunc:  func(c *gin.Context) {},
			key:        "nonexistent",
			wantValue:  (*User)(nil),
			wantExists: false,
		},
		{
			name: "key exists with string type",
			setupFunc: func(c *gin.Context) {
				c.Set("token", "abc123")
			},
			key:        "token",
			wantValue:  "abc123",
			wantExists: true,
		},
		{
			name: "key exists with int type",
			setupFunc: func(c *gin.Context) {
				c.Set("count", 42)
			},
			key:        "count",
			wantValue:  42,
			wantExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			tt.setupFunc(c)

			switch v := tt.wantValue.(type) {
			case *User:
				result, exists := GetVar[*User](c, tt.key)
				assert.Equal(t, tt.wantExists, exists)
				if exists {
					assert.Equal(t, v.Name, result.Name)
					assert.Equal(t, v.Age, result.Age)
				}
			case string:
				result, exists := GetVar[string](c, tt.key)
				assert.Equal(t, tt.wantExists, exists)
				if exists {
					assert.Equal(t, v, result)
				}
			case int:
				result, exists := GetVar[int](c, tt.key)
				assert.Equal(t, tt.wantExists, exists)
				if exists {
					assert.Equal(t, v, result)
				}
			}
		})
	}
}

func TestMustGetVar(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	tests := []struct {
		name        string
		setupFunc   func(*gin.Context)
		key         string
		expectPanic bool
		wantValue   interface{}
	}{
		{
			name: "key exists with correct type",
			setupFunc: func(c *gin.Context) {
				c.Set("user", &User{Name: "Bob", Age: 30})
			},
			key:         "user",
			expectPanic: false,
			wantValue:   &User{Name: "Bob", Age: 30},
		},
		{
			name: "key exists with wrong type",
			setupFunc: func(c *gin.Context) {
				c.Set("user", "string value")
			},
			key:         "user",
			expectPanic: true,
		},
		{
			name:        "key does not exist",
			setupFunc:   func(c *gin.Context) {},
			key:         "nonexistent",
			expectPanic: true,
		},
		{
			name: "key exists with string type",
			setupFunc: func(c *gin.Context) {
				c.Set("token", "xyz789")
			},
			key:         "token",
			expectPanic: false,
			wantValue:   "xyz789",
		},
		{
			name: "key exists with int type",
			setupFunc: func(c *gin.Context) {
				c.Set("count", 100)
			},
			key:         "count",
			expectPanic: false,
			wantValue:   100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			tt.setupFunc(c)

			if tt.expectPanic {
				assert.Panics(t, func() {
					switch tt.key {
					case "user":
						_ = MustGetVar[*User](c, tt.key)
					case "token":
						_ = MustGetVar[string](c, tt.key)
					case "count":
						_ = MustGetVar[int](c, tt.key)
					default:
						_ = MustGetVar[string](c, tt.key)
					}
				})
			} else {
				assert.NotPanics(t, func() {
					switch v := tt.wantValue.(type) {
					case *User:
						result := MustGetVar[*User](c, tt.key)
						assert.Equal(t, v.Name, result.Name)
						assert.Equal(t, v.Age, result.Age)
					case string:
						result := MustGetVar[string](c, tt.key)
						assert.Equal(t, v, result)
					case int:
						result := MustGetVar[int](c, tt.key)
						assert.Equal(t, v, result)
					}
				})
			}
		})
	}
}

// 基准测试
func BenchmarkGetVar(b *testing.B) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user", "test user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetVar[string](c, "user")
	}
}

func BenchmarkMustGetVar(b *testing.B) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user", "test user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MustGetVar[string](c, "user")
	}
}
