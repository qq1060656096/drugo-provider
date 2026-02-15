// Package i18nsvc 提供基于 mi18n 的国际化服务。
package i18nsvc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/qq1060656096/drugo/kernel"
	"github.com/qq1060656096/mi18n"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const Name = "i18n"

// 编译时检查，确保 I18nService 实现了 kernel.Service 接口。
var _ kernel.Service = (*I18nService)(nil)

// I18nService 通过 mi18n.I18n 管理多语言翻译。
type I18nService struct {
	name        string
	config      *viper.Viper
	logger      *zap.Logger
	i18n        *mi18n.I18n
	localeDir   string
	defaultLang string

	once    sync.Once
	bootErr error
}

// NewI18nService 创建一个新的 I18nService，默认名称为 "i18n"。
func NewI18nService() *I18nService {
	return &I18nService{
		name: Name,
	}
}

// Name 返回服务名称。
func (s *I18nService) Name() string {
	return s.name
}

// Boot 初始化国际化服务。它读取配置、创建mi18n实例并加载翻译文件。
// 此方法是幂等的，后续调用不会产生任何效果。
func (s *I18nService) Boot(ctx context.Context) error {
	s.once.Do(func() {
		s.bootErr = s.boot(ctx)
	})
	return s.bootErr
}

// boot 执行实际的初始化逻辑。
func (s *I18nService) boot(ctx context.Context) error {
	k := kernel.MustFromContext(ctx)

	// 检查配置是否存在，如果不存在则跳过初始化
	cfg, err := k.Config().Get(s.name)
	if err != nil {
		// 空配置是允许的，只是不会初始化i18n
		return nil
	}

	s.config = cfg
	s.logger = k.Logger().MustGet(s.name)

	s.logger.Info(s.Name()+" service config", zap.Any("config", s.config.AllSettings()))

	// 构建配置
	if err := s.buildConfig(ctx); err != nil {
		return fmt.Errorf("build i18n config: %w", err)
	}

	// 创建mi18n实例
	s.i18n = mi18n.New(s.localeDir, s.defaultLang)

	s.logger.Info("i18n service initialized",
		zap.String("locale_dir", s.localeDir),
		zap.String("default_lang", s.defaultLang),
	)

	return nil
}

// buildConfig 从 viper 配置构建服务配置。
func (s *I18nService) buildConfig(ctx context.Context) error {
	s.localeDir = s.config.GetString("locale_dir")
	if s.localeDir == "" {
		return errors.New("locale_dir is required")
	}

	// 转换为绝对路径
	if !filepath.IsAbs(s.localeDir) {
		// 尝试从kernel获取根目录来解析相对路径
		if k := kernel.MustFromContext(ctx); k.Root() != "" {
			s.localeDir = filepath.Join(k.Root(), s.localeDir)
		}
		// 如果无法获取根目录，则保持相对路径，用户需要确保路径正确
	}

	s.defaultLang = s.config.GetString("default_lang")
	if s.defaultLang == "" {
		s.defaultLang = "en" // 默认使用英文
	}

	return nil
}

// I18n 返回底层的 mi18n.I18n 实例。
// 如果 Boot 尚未被调用，则返回 nil。
func (s *I18nService) I18n() *mi18n.I18n {
	return s.i18n
}

// T 根据指定的语言和键获取翻译文本。
func (s *I18nService) T(lang, key string, data map[string]any) string {
	if s.i18n == nil {
		return key
	}
	return s.i18n.T(lang, key, data)
}

// TCtx 从context中获取语言信息并翻译文本。
func (s *I18nService) TCtx(ctx context.Context, key string, data map[string]any) string {
	if s.i18n == nil {
		return key
	}
	return s.i18n.TCtx(ctx, key, data)
}

// WithLang 将语言信息写入context。
func (s *I18nService) WithLang(ctx context.Context, lang string) context.Context {
	return mi18n.WithLang(ctx, lang)
}

// Lang 从context中获取语言信息。
func (s *I18nService) Lang(ctx context.Context) string {
	return mi18n.Lang(ctx)
}

// GetSupportedLanguages 返回支持的语言列表。
// 这个方法会扫描locale目录下的所有翻译文件，返回支持的语言代码。
func (s *I18nService) GetSupportedLanguages() []string {
	if s.i18n == nil || s.localeDir == "" {
		return []string{}
	}

	// 读取locale目录下的文件
	entries, err := os.ReadDir(s.localeDir)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to read locale directory", zap.String("dir", s.localeDir), zap.Error(err))
		}
		return []string{}
	}

	var languages []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 提取语言代码（文件名去掉扩展名）
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != "" {
			lang := name[:len(name)-len(ext)]
			languages = append(languages, lang)
		}
	}

	return languages
}

// Reload 重新加载翻译文件。
// 当翻译文件更新后，可以调用此方法重新加载。
func (s *I18nService) Reload() error {
	if s.localeDir == "" || s.defaultLang == "" {
		return errors.New("i18n service not properly initialized")
	}

	// 重新创建mi18n实例
	s.i18n = mi18n.New(s.localeDir, s.defaultLang)

	if s.logger != nil {
		s.logger.Info("i18n service reloaded",
			zap.String("locale_dir", s.localeDir),
			zap.String("default_lang", s.defaultLang),
		)
	}

	return nil
}

// Close 释放国际化服务资源。
func (s *I18nService) Close(ctx context.Context) error {
	if s.logger != nil {
		s.logger.Info("i18n service closed")
	}
	return nil
}

func New() *I18nService {
	return &I18nService{
		name:        "i18n",
		once:        sync.Once{},
		config:      nil,
		logger:      nil,
		i18n:        nil,
		localeDir:   "",
		defaultLang: "",
	}
}
