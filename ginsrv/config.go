package ginsrv

import "time"

type Config struct {
	Mode            string        `yaml:"mode"`             // debug, release, test
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"` // 优雅关闭超时，默认 30s
	ReadTimeout     time.Duration `yaml:"read_timeout"`     // HTTP 读取超时，默认 15s
	WriteTimeout    time.Duration `yaml:"write_timeout"`    // HTTP 写入超时，默认 15s
	IdleTimeout     time.Duration `yaml:"idle_timeout"`     // HTTP 空闲超时，默认 60s
	Host            string        `yaml:"host"`
	Http            struct {
		Enabled bool `yaml:"enabled"`
		Port    int  `yaml:"port"`
	} `yaml:"http"`
	Https struct {
		Enabled  bool   `yaml:"enabled"`
		Port     int    `yaml:"port"`
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
		ForceSsl bool   `yaml:"force_ssl"`
	} `yaml:"https"`
}
