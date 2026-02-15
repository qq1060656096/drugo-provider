package i18nsvc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/qq1060656096/drugo/config"
	"github.com/qq1060656096/drugo/kernel"
	"github.com/qq1060656096/drugo/log"
	"github.com/spf13/viper"
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
func createMockKernel(t *testing.T, serviceName string, i18nConfigs map[string]interface{}) *mockKernel {
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

	// 设置 i18n 配置
	for key, value := range i18nConfigs {
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
func createTestContext(t *testing.T, serviceName string, i18nConfigs map[string]interface{}) context.Context {
	k := createMockKernel(t, serviceName, i18nConfigs)
	return kernel.WithContext(context.Background(), k)
}

func TestI18nService_Boot(t *testing.T) {
	// 创建临时目录和翻译文件
	tempDir, err := os.MkdirTemp("", "i18n_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	localeDir := filepath.Join(tempDir, "locale")
	if err := os.Mkdir(localeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建测试翻译文件
	zhFile := filepath.Join(localeDir, "zh.json")
	enFile := filepath.Join(localeDir, "en.json")

	zhContent := `[{"id": "welcome", "translation": "欢迎"}, {"id": "greeting", "translation": "你好，{{.Name}}！"}]`
	enContent := `[{"id": "welcome", "translation": "Welcome"}, {"id": "greeting", "translation": "Hello, {{.Name}}!"}]`

	if err := os.WriteFile(zhFile, []byte(zhContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(enFile, []byte(enContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建配置
	configMap := map[string]interface{}{
		"locale_dir":   localeDir,
		"default_lang": "en",
	}

	ctx := createTestContext(t, Name, configMap)

	// 创建服务并测试
	service := New()
	if err := service.Boot(ctx); err != nil {
		t.Fatal(err)
	}

	// 测试翻译功能
	result := service.T("zh", "welcome", nil)
	if result != "欢迎" {
		t.Errorf("expected '欢迎', got '%s'", result)
	}

	result = service.T("en", "welcome", nil)
	if result != "Welcome" {
		t.Errorf("expected 'Welcome', got '%s'", result)
	}

	// 测试带变量的翻译
	data := map[string]any{"Name": "张三"}
	result = service.T("zh", "greeting", data)
	if result != "你好，张三！" {
		t.Errorf("expected '你好，张三！', got '%s'", result)
	}

	// 测试Context翻译
	ctxWithLang := service.WithLang(ctx, "zh")
	result = service.TCtx(ctxWithLang, "welcome", nil)
	if result != "欢迎" {
		t.Errorf("expected '欢迎', got '%s'", result)
	}

	// 测试语言获取
	lang := service.Lang(ctxWithLang)
	if lang != "zh" {
		t.Errorf("expected 'zh', got '%s'", lang)
	}

	// 测试关闭
	if err := service.Close(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestI18nService_BootWithoutConfig(t *testing.T) {
	// 创建空配置
	configMap := map[string]interface{}{}

	ctx := createTestContext(t, Name, configMap)

	// 创建服务并测试（应该成功，因为没有配置）
	service := New()
	if err := service.Boot(ctx); err != nil {
		t.Fatal(err)
	}

	// 测试未初始化状态下的翻译（应该返回key）
	result := service.T("zh", "welcome", nil)
	if result != "welcome" {
		t.Errorf("expected 'welcome', got '%s'", result)
	}
}

func TestI18nService_BootWithInvalidConfig(t *testing.T) {
	// 创建无效配置（缺少locale_dir）
	configMap := map[string]interface{}{
		"default_lang": "en",
	}

	ctx := createTestContext(t, Name, configMap)

	// 创建服务并测试（应该失败）
	service := New()
	if err := service.Boot(ctx); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestI18nService_GetSupportedLanguages(t *testing.T) {
	// 创建临时目录和翻译文件
	tempDir, err := os.MkdirTemp("", "i18n_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	localeDir := filepath.Join(tempDir, "locale")
	if err := os.Mkdir(localeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建多个语言的翻译文件
	files := map[string]string{
		"zh.json": `[{"id": "welcome", "translation": "欢迎"}]`,
		"en.json": `[{"id": "welcome", "translation": "Welcome"}]`,
		"ja.json": `[{"id": "welcome", "translation": "ようこそ"}]`,
	}

	for filename, content := range files {
		filePath := filepath.Join(localeDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// 创建配置
	configMap := map[string]interface{}{
		"locale_dir":   localeDir,
		"default_lang": "en",
	}

	ctx := createTestContext(t, Name, configMap)

	// 创建服务并测试
	service := New()
	if err := service.Boot(ctx); err != nil {
		t.Fatal(err)
	}

	// 测试获取支持的语言
	languages := service.GetSupportedLanguages()
	expected := []string{"zh", "en", "ja"}

	// 由于文件系统读取的顺序不确定，我们检查数量和包含关系
	if len(languages) != len(expected) {
		t.Errorf("expected %d languages, got %d", len(expected), len(languages))
	}

	for _, lang := range expected {
		found := false
		for _, actual := range languages {
			if actual == lang {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected language %s not found in %v", lang, languages)
		}
	}
}

func TestI18nService_Reload(t *testing.T) {
	// 创建临时目录和翻译文件
	tempDir, err := os.MkdirTemp("", "i18n_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	localeDir := filepath.Join(tempDir, "locale")
	if err := os.Mkdir(localeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建初始翻译文件
	zhFile := filepath.Join(localeDir, "zh.json")
	zhContent := `[{"id": "welcome", "translation": "欢迎"}]`
	if err := os.WriteFile(zhFile, []byte(zhContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建配置
	configMap := map[string]interface{}{
		"locale_dir":   localeDir,
		"default_lang": "en",
	}

	ctx := createTestContext(t, Name, configMap)

	// 创建服务并测试
	service := New()
	if err := service.Boot(ctx); err != nil {
		t.Fatal(err)
	}

	// 测试初始翻译
	result := service.T("zh", "welcome", nil)
	if result != "欢迎" {
		t.Errorf("expected '欢迎', got '%s'", result)
	}

	// 更新翻译文件
	updatedContent := `[{"id": "welcome", "translation": "欢迎来到我们的世界"}]`
	if err := os.WriteFile(zhFile, []byte(updatedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 重新加载
	if err := service.Reload(); err != nil {
		t.Fatal(err)
	}

	// 测试更新后的翻译
	result = service.T("zh", "welcome", nil)
	if result != "欢迎来到我们的世界" {
		t.Errorf("expected '欢迎来到我们的世界', got '%s'", result)
	}
}

func TestI18nService_GetSupportedLanguages_WithoutInit(t *testing.T) {
	service := New()
	languages := service.GetSupportedLanguages()

	// 未初始化时应该返回空列表
	if len(languages) != 0 {
		t.Errorf("expected empty list, got %v", languages)
	}
}

func TestI18nService_Reload_WithoutInit(t *testing.T) {
	service := New()
	err := service.Reload()

	// 未初始化时应该返回错误
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
