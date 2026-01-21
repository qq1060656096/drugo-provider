package ginsrv

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo/drugo"
	"github.com/qq1060656096/drugo/kernel"
	"go.uber.org/zap"
)

const Name = "gin"

var _ kernel.Service = (*GinService)(nil)
var _ kernel.Runner = (*GinService)(nil)

// Service 结构体名简练化，调用者使用 ginsrv.Service
type GinService struct {
	name       string // ← 添加 name 字段
	engine     *gin.Engine
	config     *Config
	httpServer *http.Server
	tlsServer  *http.Server
	once       sync.Once
}

// Name 实现 kernel.Service 接口
func (s *GinService) Name() string {
	return s.name
}

// Boot 初始化基础资源
func (s *GinService) Boot(ctx context.Context) error {
	k := kernel.MustFromContext(ctx)
	logger := k.Logger().MustGet(s.Name())
	s.init()
	logger.Info("booting", zap.String("name", s.Name()))
	return nil
}

// Close 优雅关闭，注意：标准库风格中，Context 应透传给 Shutdown
func (s *GinService) Close(ctx context.Context) error {
	k := kernel.MustFromContext(ctx)
	logger := k.Logger().MustGet(s.Name())
	logger.Info("closing gin service")

	// 使用配置的超时时间，默认 30 秒
	timeout := s.config.ShutdownTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var errs []error
	if s.httpServer != nil {
		logger.Info("shutting down http server", zap.String("addr", s.httpServer.Addr))
		if err := s.httpServer.Shutdown(timeoutCtx); err != nil {
			logger.Error("http server shutdown failed", zap.String("addr", s.httpServer.Addr), zap.Error(err))
			errs = append(errs, fmt.Errorf("http shutdown: %w", err))
		} else {
			logger.Info("http server shutdown completed", zap.String("addr", s.httpServer.Addr))
		}
	}

	if s.tlsServer != nil {
		logger.Info("shutting down https server", zap.String("addr", s.tlsServer.Addr))
		if err := s.tlsServer.Shutdown(timeoutCtx); err != nil {
			logger.Error("https server shutdown failed", zap.String("addr", s.tlsServer.Addr), zap.Error(err))
			errs = append(errs, fmt.Errorf("https shutdown: %w", err))
		} else {
			logger.Info("https server shutdown completed", zap.String("addr", s.tlsServer.Addr))
		}
	}

	if len(errs) > 0 {
		logger.Error("gin service closed with errors", zap.Int("error_count", len(errs)))
		return errors.Join(errs...)
	}

	logger.Info("gin service closed successfully")
	return nil
}

func (s *GinService) Run(ctx context.Context) error {
	k := kernel.MustFromContext(ctx)
	logger := k.Logger().MustGet(s.Name())

	logger.Info(s.Name() + " service starting")

	// 1. 配置加载
	logger.Debug("loading config")
	confGetter := k.Config().MustGet(s.Name())
	if err := confGetter.Unmarshal(s.config); err != nil {
		logger.Error("failed to unmarshal config", zap.Error(err))
		return fmt.Errorf("unmarshal config: %w", err)
	}
	logger.Info("config loaded",
		zap.String("mode", s.config.Mode),
		zap.String("host", s.config.Host),
		zap.Bool("http_enabled", s.config.Http.Enabled),
		zap.Int("http_port", s.config.Http.Port),
		zap.Bool("https_enabled", s.config.Https.Enabled),
		zap.Int("https_port", s.config.Https.Port),
	)

	// 2. 设置 Gin 模式
	if s.config.Mode != "" {
		gin.SetMode(s.config.Mode)
		logger.Info("gin mode set", zap.String("mode", s.config.Mode))
	}

	// 3. 获取超时配置，使用默认值
	readTimeout := s.config.ReadTimeout
	if readTimeout <= 0 {
		readTimeout = 15 * time.Second
	}
	writeTimeout := s.config.WriteTimeout
	if writeTimeout <= 0 {
		writeTimeout = 15 * time.Second
	}
	idleTimeout := s.config.IdleTimeout
	if idleTimeout <= 0 {
		idleTimeout = 60 * time.Second
	}

	errChan := make(chan error, 2)

	// 4. HTTP Server 启动
	if s.config.Http.Enabled {
		s.httpServer = &http.Server{
			Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Http.Port),
			Handler:      s.engine,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		}
		url := fmt.Sprintf("http://%s:%d", s.config.Host, s.config.Http.Port)
		if s.config.Host == "" || s.config.Host == "0.0.0.0" {
			url = fmt.Sprintf("http://%s:%d", "localhost", s.config.Http.Port)
		}
		logger.Info("starting http server",
			zap.String("url", url),
			zap.String("addr", s.httpServer.Addr),
			zap.String("protocol", "http"),
			zap.Duration("read_timeout", readTimeout),
			zap.Duration("write_timeout", writeTimeout),
			zap.Duration("idle_timeout", idleTimeout),
		)
		go func() {
			if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("http server error", zap.String("addr", s.httpServer.Addr), zap.Error(err))
				errChan <- err
			}
		}()
	} else {
		logger.Debug("http server disabled")
	}

	// 5. HTTPS Server 启动
	if s.config.Https.Enabled {
		s.tlsServer = &http.Server{
			Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Https.Port),
			Handler:      s.engine,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
			TLSConfig: &tls.Config{
				NextProtos: []string{"http/1.1"},
			},
		}
		logger.Info("starting https server",
			zap.String("addr", s.tlsServer.Addr),
			zap.String("protocol", "https"),
			zap.String("cert_file", s.config.Https.CertFile),
			zap.String("key_file", s.config.Https.KeyFile),
			zap.Duration("read_timeout", readTimeout),
			zap.Duration("write_timeout", writeTimeout),
			zap.Duration("idle_timeout", idleTimeout),
		)
		go func() {
			if err := s.tlsServer.ListenAndServeTLS(s.config.Https.CertFile, s.config.Https.KeyFile); err != nil && err != http.ErrServerClosed {
				logger.Error("https server error",
					zap.String("addr", s.tlsServer.Addr),
					zap.String("cert_file", s.config.Https.CertFile),
					zap.Error(err),
				)
				errChan <- err
			}
		}()
	} else {
		logger.Debug("https server disabled")
	}

	logger.Info("gin service running")

	// 6. 阻塞等待
	select {
	case <-ctx.Done():
		logger.Info("gin service received stop signal", zap.Error(ctx.Err()))
		return nil
	case err := <-errChan:
		logger.Error("gin service stopped due to server error", zap.Error(err))
		return fmt.Errorf("server error: %w", err)
	}
}

// Engine 获取 Gin 引擎实例
func (s *GinService) Engine() *gin.Engine {
	s.init()
	return s.engine
}

// SetEngineContextAppVar 设置 gin app变量
func (s *GinService) SetEngineContextAppVar(app kernel.Kernel) {
	s.init()
	s.engine.Use(func(c *gin.Context) {
		c.Set(drugo.Name, app)
		c.Next()
	})
}

// init 替换 doOnce，更符合内部初始化命名习惯
func (s *GinService) init() {
	s.once.Do(func() {
		s.config = &Config{}
		s.engine = gin.New()
		// 默认 Ping 路由放在初始化里
		s.engine.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
	})
}

type Option func(*GinService)

func New(opts ...Option) *GinService {
	s := &GinService{name: Name}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithName(name string) Option {
	return func(s *GinService) { s.name = name }
}
