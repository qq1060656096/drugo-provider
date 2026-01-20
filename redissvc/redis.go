// Package redissvc 提供基于 mgredis.Group 的 Redis 服务。
package redissvc

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/qq1060656096/drugo/kernel"
	"github.com/qq1060656096/mgredis"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const Name = "redis"

var _ kernel.Service = (*RedisService)(nil)

// RedisService 使用单一 mgredis.Group 管理多个 Redis 实例
type RedisService struct {
	name   string
	config *viper.Viper
	logger *zap.Logger

	group mgredis.Group

	once    sync.Once
	bootErr error
}

// New 创建 RedisService
func New() *RedisService {
	return &RedisService{
		name:  Name,
		group: mgredis.New(),
	}
}

func (s *RedisService) Name() string {
	return s.name
}

func (s *RedisService) Boot(ctx context.Context) error {
	s.once.Do(func() {
		s.bootErr = s.boot(ctx)
	})
	return s.bootErr
}

func (s *RedisService) boot(ctx context.Context) error {
	k := kernel.MustFromContext(ctx)
	s.config = k.Config().MustGet(s.name)
	s.logger = k.Logger().MustGet(s.name)

	s.logger.Info("redis service config",
		zap.Any("config", s.config.AllSettings()),
	)

	// 获取所有顶层配置项（redis 实例名称）
	allSettings := s.config.AllSettings()
	for name := range allSettings {
		cfg := s.config.Sub(name)
		if cfg == nil {
			continue
		}

		redisCfg, err := s.buildRedisConfig(cfg)
		if err != nil {
			return fmt.Errorf("build redis config %s: %w", name, err)
		}

		s.logger.Info("register redis",
			zap.String("name", name),
			zap.String("addr", redisCfg.Addr),
			zap.Int("db", redisCfg.DB),
		)

		s.group.Register(ctx, name, redisCfg)
	}

	return nil
}

// buildRedisConfig 构建 mgredis.RedisConfig
func (s *RedisService) buildRedisConfig(v *viper.Viper) (mgredis.RedisConfig, error) {
	cfg := mgredis.RedisConfig{
		Name:         v.GetString("name"),
		Addr:         v.GetString("addr"),
		Password:     v.GetString("password"),
		DB:           v.GetInt("db"),
		PoolSize:     v.GetInt("pool_size"),
		MinIdleConns: v.GetInt("min_idle_conns"),
		DialTimeout:  v.GetDuration("dial_timeout"),
		ReadTimeout:  v.GetDuration("read_timeout"),
		WriteTimeout: v.GetDuration("write_timeout"),
	}

	if cfg.Addr == "" {
		return mgredis.RedisConfig{}, errors.New("redis addr is empty")
	}

	return cfg, nil
}

// Group 暴露 group（可选，infra 层使用）
func (s *RedisService) Group() mgredis.Group {
	return s.group
}

// Close 关闭所有 Redis 连接
func (s *RedisService) Close(ctx context.Context) error {
	if s.group == nil {
		return nil
	}
	errs := s.group.Close(ctx)
	if len(errs) > 0 {
		err := errors.Join(errs...)
		if s.logger != nil {
			s.logger.Error("redis service failed to close", zap.Error(err))
		}
		return err
	}
	if s.logger != nil {
		s.logger.Info("redis service closed")
	}
	return nil
}
