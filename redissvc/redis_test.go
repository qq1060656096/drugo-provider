package redissvc

import (
	"context"
	"path/filepath"
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
	return nil
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
func createMockKernel(t *testing.T, serviceName string, redisConfigs map[string]map[string]interface{}) *mockKernel {
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

	// 设置 redis 配置
	for instanceName, instanceConfig := range redisConfigs {
		for key, value := range instanceConfig {
			v.Set(serviceName+"."+instanceName+"."+key, value)
		}
	}

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
func createTestContext(t *testing.T, serviceName string, redisConfigs map[string]map[string]interface{}) context.Context {
	k := createMockKernel(t, serviceName, redisConfigs)
	return kernel.WithContext(context.Background(), k)
}

// TestNew 测试 New 函数
func TestNew(t *testing.T) {
	service := New()
	assert.NotNil(t, service)
	assert.Equal(t, Name, service.name)
	assert.NotNil(t, service.group)
	assert.Nil(t, service.config)
	assert.Nil(t, service.logger)
}

// TestRedisService_Name 测试 Name 方法
func TestRedisService_Name(t *testing.T) {
	service := New()
	assert.Equal(t, "redis", service.Name())
}

// TestRedisService_Group 测试 Group 方法
func TestRedisService_Group(t *testing.T) {
	service := New()
	group := service.Group()
	assert.NotNil(t, group)
	assert.Equal(t, service.group, group)
}

// TestRedisService_buildRedisConfig 测试 buildRedisConfig 方法
func TestRedisService_buildRedisConfig(t *testing.T) {
	service := New()

	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
		validate    func(*testing.T, map[string]interface{})
	}{
		{
			name: "完整有效配置",
			config: map[string]interface{}{
				"name":           "test-redis",
				"addr":           "127.0.0.1:6379",
				"password":       "secret",
				"db":             1,
				"pool_size":      10,
				"min_idle_conns": 5,
				"dial_timeout":   "5s",
				"read_timeout":   "3s",
				"write_timeout":  "3s",
			},
			expectError: false,
			validate: func(t *testing.T, cfg map[string]interface{}) {
				assert.Equal(t, "test-redis", cfg["name"])
				assert.Equal(t, "127.0.0.1:6379", cfg["addr"])
				assert.Equal(t, "secret", cfg["password"])
				assert.Equal(t, 1, cfg["db"])
				assert.Equal(t, 10, cfg["pool_size"])
			},
		},
		{
			name: "最小有效配置",
			config: map[string]interface{}{
				"addr": "localhost:6379",
			},
			expectError: false,
			validate: func(t *testing.T, cfg map[string]interface{}) {
				assert.Equal(t, "localhost:6379", cfg["addr"])
			},
		},
		{
			name: "缺少addr",
			config: map[string]interface{}{
				"db": 0,
			},
			expectError: true,
		},
		{
			name:        "空配置",
			config:      map[string]interface{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			for key, value := range tt.config {
				v.Set(key, value)
			}

			cfg, err := service.buildRedisConfig(v)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					// 将配置转换为 map 以便验证
					cfgMap := map[string]interface{}{
						"name":           cfg.Name,
						"addr":           cfg.Addr,
						"password":       cfg.Password,
						"db":             cfg.DB,
						"pool_size":      cfg.PoolSize,
						"min_idle_conns": cfg.MinIdleConns,
						"dial_timeout":   cfg.DialTimeout,
						"read_timeout":   cfg.ReadTimeout,
						"write_timeout":  cfg.WriteTimeout,
					}
					tt.validate(t, cfgMap)
				}
			}
		})
	}
}

// TestRedisService_Boot 测试 Boot 方法
func TestRedisService_Boot(t *testing.T) {
	tests := []struct {
		name        string
		configs     map[string]map[string]interface{}
		expectError bool
	}{
		{
			name: "单个redis实例",
			configs: map[string]map[string]interface{}{
				"main": {
					"addr": "127.0.0.1:6379",
					"db":   0,
				},
			},
			expectError: false,
		},
		{
			name: "多个redis实例",
			configs: map[string]map[string]interface{}{
				"main": {
					"addr": "127.0.0.1:6379",
					"db":   0,
				},
				"cache": {
					"addr": "127.0.0.1:6379",
					"db":   1,
				},
				"session": {
					"addr": "127.0.0.1:6380",
					"db":   0,
				},
			},
			expectError: false,
		},
		{
			name: "包含无效实例",
			configs: map[string]map[string]interface{}{
				"main": {
					"addr": "127.0.0.1:6379",
					"db":   0,
				},
				"invalid": {
					"db": 1,
					// 缺少 addr
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := New()
			ctx := createTestContext(t, "redis", tt.configs)

			err := service.Boot(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service.config)
				assert.NotNil(t, service.logger)
			}
		})
	}
}

// TestRedisService_Boot_Once 测试 Boot 方法只执行一次
func TestRedisService_Boot_Once(t *testing.T) {
	service := New()
	configs := map[string]map[string]interface{}{
		"main": {
			"addr": "127.0.0.1:6379",
			"db":   0,
		},
	}
	ctx := createTestContext(t, "redis", configs)

	// 第一次调用
	err1 := service.Boot(ctx)
	assert.NoError(t, err1)

	// 第二次调用应该返回相同的结果（不会重新执行）
	err2 := service.Boot(ctx)
	assert.NoError(t, err2)
	assert.Equal(t, err1, err2)

	// 验证配置只初始化一次
	assert.NotNil(t, service.config)
	assert.NotNil(t, service.logger)
}

// TestRedisService_Boot_Error_Persists 测试启动错误会被保存
func TestRedisService_Boot_Error_Persists(t *testing.T) {
	service := New()
	configs := map[string]map[string]interface{}{
		"invalid": {
			"db": 0,
			// 缺少 addr，会导致错误
		},
	}
	ctx := createTestContext(t, "redis", configs)

	// 第一次调用应该返回错误
	err1 := service.Boot(ctx)
	assert.Error(t, err1)

	// 第二次调用应该返回相同的错误
	err2 := service.Boot(ctx)
	assert.Error(t, err2)
	assert.Equal(t, err1, err2)
}

// TestRedisService_Boot_WithTimeout 测试带超时的启动
func TestRedisService_Boot_WithTimeout(t *testing.T) {
	service := New()
	configs := map[string]map[string]interface{}{
		"main": {
			"addr":          "127.0.0.1:6379",
			"db":            0,
			"dial_timeout":  "5s",
			"read_timeout":  "3s",
			"write_timeout": "3s",
		},
	}
	ctx := createTestContext(t, "redis", configs)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := service.Boot(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, service.config)
	assert.NotNil(t, service.logger)
}

// TestRedisService_Close 测试 Close 方法
func TestRedisService_Close(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*testing.T) (*RedisService, context.Context)
		expectError bool
	}{
		{
			name: "正常关闭",
			setup: func(t *testing.T) (*RedisService, context.Context) {
				service := New()
				configs := map[string]map[string]interface{}{
					"main": {
						"addr": "127.0.0.1:6379",
						"db":   0,
					},
				}
				ctx := createTestContext(t, "redis", configs)
				err := service.Boot(ctx)
				require.NoError(t, err)
				return service, ctx
			},
			expectError: false,
		},
		{
			name: "未启动的服务",
			setup: func(t *testing.T) (*RedisService, context.Context) {
				service := New()
				ctx := context.Background()
				return service, ctx
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, ctx := tt.setup(t)

			err := service.Close(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRedisService_Integration 集成测试
func TestRedisService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	service := New()
	configs := map[string]map[string]interface{}{
		"main": {
			"addr":           "127.0.0.1:6379",
			"db":             0,
			"pool_size":      10,
			"min_idle_conns": 2,
			"dial_timeout":   "5s",
			"read_timeout":   "3s",
			"write_timeout":  "3s",
		},
		"cache": {
			"addr":           "127.0.0.1:6379",
			"db":             1,
			"pool_size":      5,
			"min_idle_conns": 1,
		},
	}
	ctx := createTestContext(t, "redis", configs)

	// 启动服务
	err := service.Boot(ctx)
	require.NoError(t, err)

	// 验证服务状态
	assert.NotNil(t, service.config)
	assert.NotNil(t, service.logger)
	assert.NotNil(t, service.Group())

	// 验证配置
	assert.NotNil(t, service.config.Sub("main"))
	assert.NotNil(t, service.config.Sub("cache"))

	// 关闭服务
	err = service.Close(ctx)
	assert.NoError(t, err)
}

// TestRedisService_ConfigTypes 测试不同类型的配置值
func TestRedisService_ConfigTypes(t *testing.T) {
	service := New()

	tests := []struct {
		name   string
		config map[string]interface{}
		check  func(*testing.T, map[string]interface{})
	}{
		{
			name: "字符串类型",
			config: map[string]interface{}{
				"addr":     "localhost:6379",
				"password": "test123",
			},
			check: func(t *testing.T, cfg map[string]interface{}) {
				assert.Equal(t, "localhost:6379", cfg["addr"])
				assert.Equal(t, "test123", cfg["password"])
			},
		},
		{
			name: "整数类型",
			config: map[string]interface{}{
				"addr":           "localhost:6379",
				"db":             2,
				"pool_size":      20,
				"min_idle_conns": 10,
			},
			check: func(t *testing.T, cfg map[string]interface{}) {
				assert.Equal(t, 2, cfg["db"])
				assert.Equal(t, 20, cfg["pool_size"])
				assert.Equal(t, 10, cfg["min_idle_conns"])
			},
		},
		{
			name: "持续时间类型",
			config: map[string]interface{}{
				"addr":          "localhost:6379",
				"dial_timeout":  "10s",
				"read_timeout":  "5s",
				"write_timeout": "5s",
			},
			check: func(t *testing.T, cfg map[string]interface{}) {
				assert.Equal(t, 10*time.Second, cfg["dial_timeout"])
				assert.Equal(t, 5*time.Second, cfg["read_timeout"])
				assert.Equal(t, 5*time.Second, cfg["write_timeout"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			for key, value := range tt.config {
				v.Set(key, value)
			}

			cfg, err := service.buildRedisConfig(v)
			assert.NoError(t, err)

			// 转换为 map 用于检查
			cfgMap := map[string]interface{}{
				"addr":           cfg.Addr,
				"password":       cfg.Password,
				"db":             cfg.DB,
				"pool_size":      cfg.PoolSize,
				"min_idle_conns": cfg.MinIdleConns,
				"dial_timeout":   cfg.DialTimeout,
				"read_timeout":   cfg.ReadTimeout,
				"write_timeout":  cfg.WriteTimeout,
			}

			tt.check(t, cfgMap)
		})
	}
}

// BenchmarkNew 性能测试：创建服务
func BenchmarkNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}

// BenchmarkRedisService_buildRedisConfig 性能测试：构建配置
func BenchmarkRedisService_buildRedisConfig(b *testing.B) {
	service := New()
	v := viper.New()
	v.Set("addr", "127.0.0.1:6379")
	v.Set("db", 0)
	v.Set("pool_size", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.buildRedisConfig(v)
	}
}
