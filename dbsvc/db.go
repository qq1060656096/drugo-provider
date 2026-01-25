// Package dbsvc 提供基于 mgorm 的数据库服务。
package dbsvc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/qq1060656096/drugo/kernel"
	"github.com/qq1060656096/mgorm"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const Name = "db"

// 编译时检查，确保 DbService 实现了 kernel.Service 接口。
var _ kernel.Service = (*DbService)(nil)

// DbService 通过 mgorm.Manager 管理多个数据库连接。
type DbService struct {
	name    string
	config  *viper.Viper
	logger  *zap.Logger
	manager mgorm.Manager

	once    sync.Once
	bootErr error
}

// NewDbService 创建一个新的 DbService，默认名称为 "db"。
func NewDbService() *DbService {
	return &DbService{
		name: Name,
	}
}

// Name 返回服务名称。
func (s *DbService) Name() string {
	return s.name
}

// Boot 初始化数据库服务。它读取配置、创建数据库连接并注册到 mgorm.Manager。
// 此方法是幂等的，后续调用不会产生任何效果。
func (s *DbService) Boot(ctx context.Context) error {
	s.once.Do(func() {
		s.bootErr = s.boot(ctx)
	})
	return s.bootErr
}

// boot 执行实际的初始化逻辑。
func (s *DbService) boot(ctx context.Context) error {
	s.manager = mgorm.NewManager()

	k := kernel.MustFromContext(ctx)

	// 检查配置是否存在，如果不存在则跳过初始化
	cfg, err := k.Config().Get(s.name)
	if err != nil {
		// 空配置是允许的，只是不会注册任何数据库
		return nil
	}

	s.config = cfg
	s.logger = k.Logger().MustGet(s.name)

	s.logger.Info(s.Name()+" service config", zap.Any("config", s.config.AllSettings()))
	registered := make(map[string]struct{})

	for _, key := range s.config.AllKeys() {
		parts := strings.Split(key, ".")
		if len(parts) < 3 {
			continue
		}

		groupName, dbName := parts[0], parts[1]
		if groupName == "" || dbName == "" {
			continue
		}

		regKey := groupName + "." + dbName
		if _, ok := registered[regKey]; ok {
			continue
		}
		registered[regKey] = struct{}{}

		if err := s.registerDB(ctx, groupName, dbName); err != nil {
			s.logger.Error("failed to register db", zap.String("group", groupName), zap.String("db", dbName))
			return fmt.Errorf("register db %s.%s: %w", groupName, dbName, err)
		}
	}

	return nil
}

// registerDB 将单个数据库连接注册到指定的分组。
func (s *DbService) registerDB(ctx context.Context, groupName, dbName string) error {
	s.manager.AddGroup(groupName)

	groupCfg := s.config.Sub(groupName)
	if groupCfg == nil {
		s.logger.Error("group config not found", zap.String("group", groupName))
		return fmt.Errorf("group config %q not found", groupName)
	}

	dbCfg := groupCfg.Sub(dbName)
	if dbCfg == nil {
		s.logger.Error("db config not found", zap.String("group", groupName), zap.String("db", dbName))
		return fmt.Errorf("db config %q not found in group %q", dbName, groupName)
	}

	cfg, err := s.buildDBConfig(dbCfg)
	if err != nil {
		return err
	}

	s.logger.Info("registering database",
		zap.String("group", groupName),
		zap.String("db", dbName),
		zap.String("driver", cfg.DriverType),
	)

	s.manager.MustGroup(groupName).Register(ctx, dbName, cfg)
	err = s.manager.MustGroup(groupName).Ping(ctx, dbName)
	if err != nil {
		s.logger.Error("failed to ping db", zap.String("group", groupName), zap.String("db", dbName))
	}
	s.logger.Info("database registered",
		zap.String("group", groupName),
		zap.String("db", dbName),
	)

	return err
}

// buildDBConfig 从 viper 配置创建 mgorm.DBConfig。
func (s *DbService) buildDBConfig(v *viper.Viper) (mgorm.DBConfig, error) {
	cfg := mgorm.DBConfig{
		Name:            v.GetString("name"),
		DriverType:      v.GetString("driver_type"),
		DSN:             v.GetString("dsn"),
		Host:            v.GetString("host"),
		Port:            v.GetInt("port"),
		User:            v.GetString("user"),
		Password:        v.GetString("password"),
		DBName:          v.GetString("db_name"),
		Charset:         v.GetString("charset"),
		MaxIdleConns:    v.GetInt("max_idle_conns"),
		MaxOpenConns:    v.GetInt("max_open_conns"),
		ConnMaxLifetime: v.GetDuration("conn_max_lifetime"),
	}
	if cfg.DSN == "" {
		cfg.DSN = cfg.AutoDsn()
	}

	dialector, err := s.createDialector(cfg.DriverType, cfg.DSN)
	if err != nil {
		return mgorm.DBConfig{}, fmt.Errorf("create dialector: %w", err)
	}
	cfg.Dialector = dialector

	return cfg, nil
}

// createDialector 根据指定的驱动类型创建 gorm Dialector。
func (s *DbService) createDialector(driverType, dsn string) (gorm.Dialector, error) {
	return CreateDialector(driverType, dsn)
}

// Close 释放此服务管理的所有数据库连接。
func (s *DbService) Close(ctx context.Context) error {
	if s.manager == nil {
		return nil
	}
	// TODO: 当 mgorm 支持时，实现正确的连接清理
	errs := s.manager.Close(ctx)
	if len(errs) > 0 {
		err := errors.Join(errs...)
		s.logger.Error("mgorm service failed to close", zap.Error(err))
		return err
	}
	s.logger.Info("mgorm service closed")
	return nil
}

// Manager 返回底层的 mgorm.Manager 实例。
// 如果 Boot 尚未被调用，则返回 nil。
func (s *DbService) Manager() mgorm.Manager {
	return s.manager
}

func New() *DbService {
	return &DbService{
		name:    "db",
		once:    sync.Once{},
		config:  nil,
		logger:  nil,
		manager: nil,
	}
}
