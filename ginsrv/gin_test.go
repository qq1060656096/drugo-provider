package ginsrv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/qq1060656096/drugo/config"
	"github.com/qq1060656096/drugo/kernel"
	"github.com/qq1060656096/drugo/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockKernel 模拟 kernel 接口
type mockKernel struct {
	logger *log.Manager
	config *config.Manager
	name   string
}

func (m *mockKernel) Container() kernel.Container[kernel.Service] {
	return nil // 简化实现
}

func (m *mockKernel) Boot(ctx context.Context) error {
	return nil
}

func (m *mockKernel) Run(ctx context.Context) error {
	return nil
}

func (m *mockKernel) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockKernel) Root() string {
	return "/tmp"
}

func (m *mockKernel) Config() *config.Manager {
	return m.config
}

func (m *mockKernel) Logger() *log.Manager {
	return m.logger
}

func (m *mockKernel) Serve(ctx context.Context) error {
	return nil
}

func (m *mockKernel) Name() string {
	return m.name
}

// createMockKernel 创建模拟的 kernel 实例
func createMockKernel(t *testing.T, serviceName string, ginConfig *Config) *mockKernel {
	// 创建日志管理器
	logConfig := log.Config{
		Level:  "info",
		Format: "console",
		Dir:    t.TempDir(),
	}
	logManager, err := log.NewManager(logConfig)
	require.NoError(t, err)

	// 创建配置目录和文件
	configDir := t.TempDir()
	configFile := filepath.Join(configDir, serviceName+".yaml")

	// 创建 viper 实例并设置配置
	v := viper.New()
	// 创建根配置，包含服务名称作为顶级键
	v.Set(serviceName+".mode", ginConfig.Mode)
	v.Set(serviceName+".shutdown_timeout", ginConfig.ShutdownTimeout)
	v.Set(serviceName+".read_timeout", ginConfig.ReadTimeout)
	v.Set(serviceName+".write_timeout", ginConfig.WriteTimeout)
	v.Set(serviceName+".idle_timeout", ginConfig.IdleTimeout)
	v.Set(serviceName+".host", ginConfig.Host)
	v.Set(serviceName+".http.enabled", ginConfig.Http.Enabled)
	v.Set(serviceName+".http.port", ginConfig.Http.Port)
	v.Set(serviceName+".https.enabled", ginConfig.Https.Enabled)
	v.Set(serviceName+".https.port", ginConfig.Https.Port)
	v.Set(serviceName+".https.cert_file", ginConfig.Https.CertFile)
	v.Set(serviceName+".https.key_file", ginConfig.Https.KeyFile)
	v.Set(serviceName+".https.force_ssl", ginConfig.Https.ForceSsl)

	// 写入配置文件
	err = v.WriteConfigAs(configFile)
	require.NoError(t, err)

	// 创建配置管理器
	configManager, err := config.NewManager(configDir)
	require.NoError(t, err)

	return &mockKernel{
		logger: logManager,
		config: configManager,
		name:   serviceName,
	}
}

// createTestContext 创建测试用的上下文
func createTestContext(t *testing.T, serviceName string, ginConfig *Config) context.Context {
	k := createMockKernel(t, serviceName, ginConfig)
	return kernel.WithContext(context.Background(), k)
}

// TestNew 测试 New 函数
func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		opts     []Option
		expected string
	}{
		{
			name:     "默认创建",
			opts:     nil,
			expected: "gin",
		},
		{
			name:     "自定义名称",
			opts:     []Option{WithName("custom-gin")},
			expected: "custom-gin",
		},
		{
			name:     "多个选项",
			opts:     []Option{WithName("multi-gin")},
			expected: "multi-gin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := New(tt.opts...)
			assert.NotNil(t, service)
			assert.Equal(t, tt.expected, service.Name())
		})
	}
}

// TestWithName 测试 WithName 选项函数
func TestWithName(t *testing.T) {
	service := &GinService{}
	opt := WithName("test-service")
	opt(service)
	assert.Equal(t, "test-service", service.name)
}

// TestGinService_Name 测试 Name 方法
func TestGinService_Name(t *testing.T) {
	service := New(WithName("test-name"))
	assert.Equal(t, "test-name", service.Name())
}

// TestGinService_Engine 测试 Engine 方法
func TestGinService_Engine(t *testing.T) {
	service := New()

	// 第一次调用应该初始化引擎
	engine1 := service.Engine()
	assert.NotNil(t, engine1)

	// 第二次调用应该返回同一个引擎实例
	engine2 := service.Engine()
	assert.Same(t, engine1, engine2)

	// 验证默认路由存在
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	engine1.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, "pong", response["message"])
}

// TestGinService_init 测试 init 方法
func TestGinService_init(t *testing.T) {
	service := New()

	// 初始化前应该为空
	assert.Nil(t, service.engine)
	assert.Nil(t, service.config)

	// 调用初始化
	service.init()

	// 初始化后应该有值
	assert.NotNil(t, service.engine)
	assert.NotNil(t, service.config)

	// 验证默认路由
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	service.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGinService_Boot 测试 Boot 方法
func TestGinService_Boot(t *testing.T) {
	service := New(WithName("test-gin"))
	config := &Config{Mode: "test"}
	ctx := createTestContext(t, "test-gin", config)

	// 测试正常启动
	err := service.Boot(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, service.engine)
}

// TestGinService_Close 测试 Close 方法
func TestGinService_Close(t *testing.T) {
	tests := []struct {
		name           string
		setupService   func(*testing.T) *GinService
		expectedError  bool
		expectedFields []string
	}{
		{
			name: "无服务器运行",
			setupService: func(t *testing.T) *GinService {
				service := New(WithName("test-close"))
				service.init() // 初始化配置
				return service
			},
			expectedError: false,
		},
		{
			name: "只有HTTP服务器",
			setupService: func(t *testing.T) *GinService {
				service := New(WithName("test-close-http"))
				service.init() // 初始化配置
				service.httpServer = &http.Server{Addr: ":8080"}
				return service
			},
			expectedError: false,
		},
		{
			name: "只有HTTPS服务器",
			setupService: func(t *testing.T) *GinService {
				service := New(WithName("test-close-https"))
				service.init() // 初始化配置
				service.tlsServer = &http.Server{Addr: ":8443"}
				return service
			},
			expectedError: false,
		},
		{
			name: "两个服务器都存在",
			setupService: func(t *testing.T) *GinService {
				service := New(WithName("test-close-both"))
				service.init() // 初始化配置
				service.httpServer = &http.Server{Addr: ":8080"}
				service.tlsServer = &http.Server{Addr: ":8443"}
				return service
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			config := &Config{ShutdownTimeout: 5 * time.Second}
			ctx := createTestContext(t, service.Name(), config)

			err := service.Close(ctx)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGinService_Close_Timeout 测试关闭超时场景
func TestGinService_Close_Timeout(t *testing.T) {
	t.Skip("跳过超时测试，需要更复杂的设置")

	service := New(WithName("test-timeout"))
	service.init() // 初始化配置

	// TODO: 实现真正的超时测试
	// 需要创建一个在关闭时会阻塞的服务器
	config := &Config{ShutdownTimeout: 100 * time.Millisecond}
	ctx := createTestContext(t, service.Name(), config)

	err := service.Close(ctx)
	assert.NoError(t, err) // 暂时期望成功
}

// TestGinService_Run 测试 Run 方法
func TestGinService_Run(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectedError bool
		setupInvalid  bool // 新增字段用于标记无效配置场景
	}{
		{
			name: "只有HTTP服务器",
			config: &Config{
				Mode: "test",
				Host: "localhost",
				Http: struct {
					Enabled bool `yaml:"enabled"`
					Port    int  `yaml:"port"`
				}{
					Enabled: true,
					Port:    0, // 使用随机端口
				},
				Https: struct {
					Enabled  bool   `yaml:"enabled"`
					Port     int    `yaml:"port"`
					CertFile string `yaml:"cert_file"`
					KeyFile  string `yaml:"key_file"`
					ForceSsl bool   `yaml:"force_ssl"`
				}{
					Enabled: false,
				},
			},
			expectedError: false,
		},
		{
			name:          "配置解析失败",
			config:        &Config{Mode: "test"},
			expectedError: true,
			setupInvalid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := New(WithName("test-run"))

			var ctx context.Context
			if tt.setupInvalid {
				// 为无效配置场景创建空的配置目录
				configDir := t.TempDir()
				configManager, err := config.NewManager(configDir)
				require.NoError(t, err)

				k := &mockKernel{
					logger: createMockKernel(t, "test-run", &Config{}).logger,
					config: configManager,
					name:   "test-run",
				}
				ctx = kernel.WithContext(context.Background(), k)
			} else {
				ctx = createTestContext(t, "test-run", tt.config)
			}

			// 先调用 Boot 初始化服务
			err := service.Boot(ctx)
			if err != nil {
				if tt.expectedError {
					assert.Error(t, err)
					return
				} else {
					t.Fatalf("Boot failed: %v", err)
				}
			}

			// 使用带超时的上下文
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if tt.expectedError {
				// 对于期望错误的场景，检查是否 panic
				assert.Panics(t, func() {
					service.Run(ctx)
				})
			} else {
				err = service.Run(ctx)
				// 正常情况下应该因为上下文取消而返回
				assert.NoError(t, err)
			}
		})
	}
}

// TestGinService_Run_ContextCancellation 测试上下文取消
func TestGinService_Run_ContextCancellation(t *testing.T) {
	service := New(WithName("test-cancel"))
	config := &Config{
		Mode: "test",
		Host: "localhost",
		Http: struct {
			Enabled bool `yaml:"enabled"`
			Port    int  `yaml:"port"`
		}{
			Enabled: true,
			Port:    0, // 随机端口
		},
	}

	ctx := createTestContext(t, "test-cancel", config)

	// 先调用 Boot 初始化服务
	err := service.Boot(ctx)
	require.NoError(t, err)

	// 创建一个会很快取消的上下文
	ctx, cancel := context.WithCancel(ctx)

	// 启动服务
	errChan := make(chan error, 1)
	go func() {
		errChan <- service.Run(ctx)
	}()

	// 等待一小段时间后取消
	time.Sleep(100 * time.Millisecond)
	cancel()

	// 等待服务停止
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("服务未在预期时间内停止")
	}
}

// TestGinService_Run_DefaultTimeouts 测试默认超时值
func TestGinService_Run_DefaultTimeouts(t *testing.T) {
	service := New(WithName("test-timeouts"))
	config := &Config{
		Mode:         "test",
		Host:         "localhost",
		ReadTimeout:  0, // 使用默认值
		WriteTimeout: 0, // 使用默认值
		IdleTimeout:  0, // 使用默认值
		Http: struct {
			Enabled bool `yaml:"enabled"`
			Port    int  `yaml:"port"`
		}{
			Enabled: true,
			Port:    0,
		},
	}

	ctx := createTestContext(t, "test-timeouts", config)

	// 先调用 Boot 初始化服务
	err := service.Boot(ctx)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err = service.Run(ctx)
	assert.NoError(t, err)
}

// TestGinService_ConcurrentAccess 测试并发访问
func TestGinService_ConcurrentAccess(t *testing.T) {
	service := New(WithName("test-concurrent"))

	var wg sync.WaitGroup
	numGoroutines := 10

	// 并发调用 Engine 方法
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine := service.Engine()
			assert.NotNil(t, engine)
		}()
	}

	wg.Wait()

	// 验证只有一个引擎实例
	engine1 := service.Engine()
	engine2 := service.Engine()
	assert.Same(t, engine1, engine2)
}

// TestGinService_GinMode 测试 Gin 模式设置
func TestGinService_GinMode(t *testing.T) {
	tests := []struct {
		name  string
		mode  string
		valid bool
	}{
		{"debug模式", "debug", true},
		{"release模式", "release", true},
		{"test模式", "test", true},
		{"空模式", "", true},
		{"无效模式", "invalid", false}, // 无效模式会导致 panic
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := New(WithName("test-mode"))
			config := &Config{
				Mode: tt.mode,
				Http: struct {
					Enabled bool `yaml:"enabled"`
					Port    int  `yaml:"port"`
				}{
					Enabled: false, // 禁用服务器避免实际启动
				},
			}

			ctx := createTestContext(t, "test-mode", config)

			// 先调用 Boot 初始化服务
			err := service.Boot(ctx)
			require.NoError(t, err)

			if !tt.valid {
				// 对于无效模式，期望 panic
				assert.Panics(t, func() {
					ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
					defer cancel()
					service.Run(ctx)
				})
			} else {
				// 对于有效模式，应该正常运行
				ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
				defer cancel()
				err = service.Run(ctx)
				// 应该因为上下文取消而正常返回
				assert.NoError(t, err)
			}
		})
	}
}

// BenchmarkGinService_Engine 性能测试：Engine 方法
func BenchmarkGinService_Engine(b *testing.B) {
	service := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Engine()
	}
}

// BenchmarkGinService_ConcurrentEngine 性能测试：并发访问 Engine
func BenchmarkGinService_ConcurrentEngine(b *testing.B) {
	service := New()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			service.Engine()
		}
	})
}

// TestGinService_Integration 集成测试
func TestGinService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	service := New(WithName("integration-test"))
	config := &Config{
		Mode:            "test",
		Host:            "localhost",
		ShutdownTimeout: 5 * time.Second,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     30 * time.Second,
		Http: struct {
			Enabled bool `yaml:"enabled"`
			Port    int  `yaml:"port"`
		}{
			Enabled: true,
			Port:    0, // 随机端口
		},
		Https: struct {
			Enabled  bool   `yaml:"enabled"`
			Port     int    `yaml:"port"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
			ForceSsl bool   `yaml:"force_ssl"`
		}{
			Enabled: false,
		},
	}

	ctx := createTestContext(t, "integration-test", config)

	// 启动服务
	err := service.Boot(ctx)
	require.NoError(t, err)

	// 运行服务（在后台）
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- service.Run(ctx)
	}()

	// 等待服务运行完成
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("集成测试超时")
	}

	// 关闭服务
	closeCtx := createTestContext(t, "integration-test", config)
	err = service.Close(closeCtx)
	assert.NoError(t, err)
}
