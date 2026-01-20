package dbsvc

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
	"go.uber.org/zap"
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
func createMockKernel(t *testing.T, serviceName string, dbConfigs map[string]interface{}) *mockKernel {
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

	// 设置 db 配置
	for key, value := range dbConfigs {
		v.Set(serviceName+"."+key, value)
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
func createTestContext(t *testing.T, serviceName string, dbConfigs map[string]interface{}) context.Context {
	k := createMockKernel(t, serviceName, dbConfigs)
	return kernel.WithContext(context.Background(), k)
}

// setupTestKernel 设置测试环境，返回 kernel 和 context
func setupTestKernel(t *testing.T, dbConfigs map[string]interface{}) (kernel.Kernel, context.Context) {
	k := createMockKernel(t, Name, dbConfigs)
	ctx := kernel.WithContext(context.Background(), k)
	return k, ctx
}

func TestNewDbService(t *testing.T) {
	svc := NewDbService()
	assert.NotNil(t, svc)
	assert.Equal(t, Name, svc.name)
	assert.Nil(t, svc.config)
	assert.Nil(t, svc.logger)
	assert.Nil(t, svc.mgorm)
}

func TestNew(t *testing.T) {
	svc := New()
	assert.NotNil(t, svc)
	assert.Equal(t, "db", svc.name)
	assert.Nil(t, svc.config)
	assert.Nil(t, svc.logger)
	assert.Nil(t, svc.mgorm)
}

func TestDbService_Name(t *testing.T) {
	svc := NewDbService()
	assert.Equal(t, Name, svc.Name())
}

func TestDbService_Manager_BeforeBoot(t *testing.T) {
	svc := NewDbService()
	assert.Nil(t, svc.Manager())
}

func TestDbService_Boot_Success_SQLite(t *testing.T) {
	configMap := map[string]interface{}{
		"public.common.name":              "common",
		"public.common.driver_type":       "sqlite",
		"public.common.dsn":               ":memory:",
		"public.common.max_idle_conns":    2,
		"public.common.max_open_conns":    10,
		"public.common.conn_max_lifetime": "1h",
	}

	ctx := createTestContext(t, Name, configMap)
	svc := NewDbService()

	err := svc.Boot(ctx)
	require.NoError(t, err)

	// 验证服务已初始化
	assert.NotNil(t, svc.config)
	assert.NotNil(t, svc.logger)
	assert.NotNil(t, svc.mgorm)
	assert.NotNil(t, svc.Manager())

	// 验证数据库组已创建
	group, err := svc.mgorm.Group("public")
	require.NoError(t, err)
	assert.NotNil(t, group)

	// 验证数据库已注册
	db, err := group.Get(ctx, "common")
	require.NoError(t, err)
	assert.NotNil(t, db)

	// 清理
	err = svc.Close(ctx)
	assert.NoError(t, err)
}

func TestDbService_Boot_Multiple_Groups_And_DBs(t *testing.T) {
	configMap := map[string]interface{}{
		"public.common.name":                     "common",
		"public.common.driver_type":              "sqlite",
		"public.common.dsn":                      ":memory:",
		"public.common.max_idle_conns":           2,
		"public.common.max_open_conns":           10,
		"public.common.conn_max_lifetime":        "1h",
		"business.test_data_1.name":              "test_data_1",
		"business.test_data_1.driver_type":       "sqlite",
		"business.test_data_1.dsn":               ":memory:",
		"business.test_data_1.max_idle_conns":    5,
		"business.test_data_1.max_open_conns":    20,
		"business.test_data_1.conn_max_lifetime": "2h",
		"business.test_data_2.name":              "test_data_2",
		"business.test_data_2.driver_type":       "sqlite",
		"business.test_data_2.dsn":               ":memory:",
	}

	ctx := createTestContext(t, Name, configMap)
	svc := NewDbService()

	err := svc.Boot(ctx)
	require.NoError(t, err)

	// 验证第一个组
	publicGroup, err := svc.mgorm.Group("public")
	require.NoError(t, err)
	assert.NotNil(t, publicGroup)
	commonDB, err := publicGroup.Get(ctx, "common")
	require.NoError(t, err)
	assert.NotNil(t, commonDB)

	// 验证第二个组
	businessGroup, err := svc.mgorm.Group("business")
	require.NoError(t, err)
	assert.NotNil(t, businessGroup)
	testData1DB, err := businessGroup.Get(ctx, "test_data_1")
	require.NoError(t, err)
	assert.NotNil(t, testData1DB)
	testData2DB, err := businessGroup.Get(ctx, "test_data_2")
	require.NoError(t, err)
	assert.NotNil(t, testData2DB)

	// 清理
	err = svc.Close(ctx)
	assert.NoError(t, err)
}

func TestDbService_Boot_Idempotent(t *testing.T) {
	configMap := map[string]interface{}{
		"public.common.name":        "common",
		"public.common.driver_type": "sqlite",
		"public.common.dsn":         ":memory:",
	}

	ctx := createTestContext(t, Name, configMap)
	svc := NewDbService()

	// 第一次 Boot
	err1 := svc.Boot(ctx)
	require.NoError(t, err1)

	mgorm1 := svc.mgorm

	// 第二次 Boot 应该是幂等的
	err2 := svc.Boot(ctx)
	assert.NoError(t, err2)

	// mgorm 应该是同一个实例
	assert.Equal(t, mgorm1, svc.mgorm)

	// 清理
	err := svc.Close(ctx)
	assert.NoError(t, err)
}

func TestDbService_Boot_InvalidDriverType(t *testing.T) {
	configMap := map[string]interface{}{
		"public.common.name":        "common",
		"public.common.driver_type": "invalid_driver",
		"public.common.dsn":         "some_dsn",
	}

	ctx := createTestContext(t, Name, configMap)
	svc := NewDbService()

	err := svc.Boot(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown driver type")
}

func TestDbService_Boot_EmptyConfig(t *testing.T) {
	configMap := map[string]interface{}{}

	ctx := createTestContext(t, Name, configMap)
	svc := NewDbService()

	err := svc.Boot(ctx)
	// 空配置应该成功，但不会注册任何数据库
	assert.NoError(t, err)
	assert.NotNil(t, svc.mgorm)
}

func TestDbService_Boot_InvalidConfig_MissingGroupName(t *testing.T) {
	configMap := map[string]interface{}{
		"common.name":        "common",
		"common.driver_type": "sqlite",
		"common.dsn":         ":memory:",
	}

	ctx := createTestContext(t, Name, configMap)
	svc := NewDbService()

	// 单级配置 key 会被跳过
	err := svc.Boot(ctx)
	assert.NoError(t, err)
}

func TestDbService_buildDBConfig(t *testing.T) {
	svc := NewDbService()
	svc.logger = zap.NewNop()

	tests := []struct {
		name        string
		config      map[string]interface{}
		wantErr     bool
		errContains string
		validate    func(t *testing.T, cfg interface{})
	}{
		{
			name: "valid sqlite config",
			config: map[string]interface{}{
				"name":              "test_db",
				"driver_type":       "sqlite",
				"dsn":               ":memory:",
				"max_idle_conns":    5,
				"max_open_conns":    10,
				"conn_max_lifetime": "1h30m",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg interface{}) {
				// Type assertion would fail in real scenario without proper type
				// This is just to show the structure
			},
		},
		{
			name: "valid mysql config",
			config: map[string]interface{}{
				"name":              "mysql_db",
				"driver_type":       "mysql",
				"dsn":               "user:pass@tcp(127.0.0.1:3306)/dbname",
				"max_idle_conns":    3,
				"max_open_conns":    15,
				"conn_max_lifetime": "2h",
			},
			wantErr: false,
		},
		{
			name: "invalid driver type",
			config: map[string]interface{}{
				"name":        "invalid_db",
				"driver_type": "unknown",
				"dsn":         "some_dsn",
			},
			wantErr:     true,
			errContains: "unknown driver type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			for key, value := range tt.config {
				v.Set(key, value)
			}

			cfg, err := svc.buildDBConfig(v)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg.Dialector)
				assert.Equal(t, tt.config["name"], cfg.Name)
				assert.Equal(t, tt.config["driver_type"], cfg.DriverType)
				assert.Equal(t, tt.config["dsn"], cfg.DSN)
				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

func TestDbService_createDialector(t *testing.T) {
	svc := NewDbService()

	tests := []struct {
		name       string
		driverType string
		dsn        string
		wantErr    bool
	}{
		{
			name:       "sqlite",
			driverType: "sqlite",
			dsn:        ":memory:",
			wantErr:    false,
		},
		{
			name:       "mysql",
			driverType: "mysql",
			dsn:        "user:pass@tcp(localhost:3306)/dbname",
			wantErr:    false,
		},
		{
			name:       "postgres",
			driverType: "postgres",
			dsn:        "host=localhost user=user password=pass dbname=testdb port=5432",
			wantErr:    false,
		},
		{
			name:       "sqlserver",
			driverType: "sqlserver",
			dsn:        "sqlserver://user:pass@localhost:1433?database=testdb",
			wantErr:    false,
		},
		{
			name:       "unknown driver",
			driverType: "unknown",
			dsn:        "some_dsn",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialector, err := svc.createDialector(tt.driverType, tt.dsn)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, dialector)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, dialector)
			}
		})
	}
}

func TestDbService_Close_BeforeBoot(t *testing.T) {
	svc := NewDbService()
	err := svc.Close(context.Background())
	assert.NoError(t, err)
}

func TestDbService_Close_AfterBoot(t *testing.T) {
	configMap := map[string]interface{}{
		"public.common.name":        "common",
		"public.common.driver_type": "sqlite",
		"public.common.dsn":         ":memory:",
	}

	_, ctx := setupTestKernel(t, configMap)
	svc := NewDbService()

	err := svc.Boot(ctx)
	require.NoError(t, err)

	err = svc.Close(ctx)
	assert.NoError(t, err)
}

func TestDbService_registerDB_Success(t *testing.T) {
	configMap := map[string]interface{}{
		"public.common.name":              "common",
		"public.common.driver_type":       "sqlite",
		"public.common.dsn":               ":memory:",
		"public.common.max_idle_conns":    2,
		"public.common.max_open_conns":    10,
		"public.common.conn_max_lifetime": "1h",
	}

	ctx := createTestContext(t, Name, configMap)
	svc := NewDbService()

	// 由于 registerDB 需要 mgorm，我们需要先初始化它
	err := svc.Boot(ctx)
	require.NoError(t, err)

	// 验证注册成功
	group, err := svc.mgorm.Group("public")
	require.NoError(t, err)
	assert.NotNil(t, group)
	db, err := group.Get(ctx, "common")
	require.NoError(t, err)
	assert.NotNil(t, db)

	// 清理
	err = svc.Close(ctx)
	assert.NoError(t, err)
}

func TestDbService_Boot_WithDurationConfig(t *testing.T) {
	configMap := map[string]interface{}{
		"public.common.name":              "common",
		"public.common.driver_type":       "sqlite",
		"public.common.dsn":               ":memory:",
		"public.common.conn_max_lifetime": "90m",
	}

	ctx := createTestContext(t, Name, configMap)
	svc := NewDbService()

	err := svc.Boot(ctx)
	require.NoError(t, err)

	// 验证配置被正确解析
	assert.NotNil(t, svc.mgorm)

	// 清理
	err = svc.Close(ctx)
	assert.NoError(t, err)
}

func TestDbService_Boot_ErrorPersistsOnRetry(t *testing.T) {
	configMap := map[string]interface{}{
		"public.common.name":        "common",
		"public.common.driver_type": "invalid",
		"public.common.dsn":         "dsn",
	}

	ctx := createTestContext(t, Name, configMap)
	svc := NewDbService()

	// 第一次 Boot 失败
	err1 := svc.Boot(ctx)
	require.Error(t, err1)

	// 第二次 Boot 应该返回相同的错误（幂等性）
	err2 := svc.Boot(ctx)
	assert.Error(t, err2)
	assert.Equal(t, err1.Error(), err2.Error())
}

func TestDbService_Integration_FullLifecycle(t *testing.T) {
	configMap := map[string]interface{}{
		"public.db1.name":         "db1",
		"public.db1.driver_type":  "sqlite",
		"public.db1.dsn":          ":memory:",
		"private.db2.name":        "db2",
		"private.db2.driver_type": "sqlite",
		"private.db2.dsn":         ":memory:",
	}

	ctx := createTestContext(t, Name, configMap)
	svc := NewDbService()

	// 1. 测试初始状态
	assert.Equal(t, Name, svc.Name())
	assert.Nil(t, svc.Manager())

	// 2. Boot 服务
	err := svc.Boot(ctx)
	require.NoError(t, err)
	assert.NotNil(t, svc.Manager())

	// 3. 验证数据库连接可用
	publicGroup := svc.mgorm.MustGroup("public")
	db1 := publicGroup.MustGet(ctx, "db1")
	assert.NotNil(t, db1)

	privateGroup := svc.mgorm.MustGroup("private")
	db2 := privateGroup.MustGet(ctx, "db2")
	assert.NotNil(t, db2)

	// 4. 测试数据库操作（创建表）
	err = db1.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error
	assert.NoError(t, err)

	// 5. 插入数据
	err = db1.Exec("INSERT INTO test (id, name) VALUES (?, ?)", 1, "test").Error
	assert.NoError(t, err)

	// 6. 查询数据
	var count int64
	err = db1.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// 7. 关闭服务
	err = svc.Close(ctx)
	assert.NoError(t, err)
}

func TestDbService_ConfigParsing_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		configMap map[string]interface{}
		shouldErr bool
	}{
		{
			name: "empty group name",
			configMap: map[string]interface{}{
				".db1.name":        "db1",
				".db1.driver_type": "sqlite",
				".db1.dsn":         ":memory:",
			},
			shouldErr: false, // 会被跳过，不会报错
		},
		{
			name: "empty db name",
			configMap: map[string]interface{}{
				"group..name":        "db1",
				"group..driver_type": "sqlite",
				"group..dsn":         ":memory:",
			},
			shouldErr: false, // 会被跳过，不会报错
		},
		{
			name: "single level key",
			configMap: map[string]interface{}{
				"single": "value",
			},
			shouldErr: false, // 会被跳过
		},
		{
			name: "zero values for connection pool",
			configMap: map[string]interface{}{
				"public.db1.name":              "db1",
				"public.db1.driver_type":       "sqlite",
				"public.db1.dsn":               ":memory:",
				"public.db1.max_idle_conns":    0,
				"public.db1.max_open_conns":    0,
				"public.db1.conn_max_lifetime": 0,
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext(t, Name, tt.configMap)
			svc := NewDbService()

			err := svc.Boot(ctx)

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 清理
			_ = svc.Close(ctx)
		})
	}
}

func TestDbService_ConnMaxLifetime_Parsing(t *testing.T) {
	tests := []struct {
		name     string
		lifetime interface{}
		expected time.Duration
	}{
		{
			name:     "duration string",
			lifetime: "1h30m",
			expected: 90 * time.Minute,
		},
		{
			name:     "seconds",
			lifetime: "3600s",
			expected: 1 * time.Hour,
		},
		{
			name:     "milliseconds",
			lifetime: "1000ms",
			expected: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			v.Set("conn_max_lifetime", tt.lifetime)

			duration := v.GetDuration("conn_max_lifetime")
			assert.Equal(t, tt.expected, duration)
		})
	}
}
