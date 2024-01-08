package orbit

import (
	"strings"
	"time"

	m "github.com/shengyanli1982/orbit/internal/middleware"
	"go.uber.org/zap"
)

var (
	defaultConsoleLogger     = NewLogger(nil)                            // default console logger
	defaultHttpListemAddress = "127.0.0.0"                               // default http listen address
	defaultHttpListemPort    = uint16(8080)                              // default http listen port
	defaultIdleTimeout       = uint32((15 * time.Second).Milliseconds()) // http idle timeout
)

// 配置
// configuration
type Config struct {
	Address               string             `json:"address,omitempty" yaml:"address,omitempty"`
	Port                  uint16             `json:"port,omitempty" yaml:"port,omitempty"`
	ReleaseMode           bool               `json:"releaseMode,omitempty" yaml:"releaseMode,omitempty"`
	HttpReadTimeout       uint32             `json:"httpReadTimeout,omitempty" yaml:"httpReadTimeout,omitempty"`
	HttpWriteTimeout      uint32             `json:"httpWriteTimeout,omitempty" yaml:"httpWriteTimeout,omitempty"`
	HttpReadHeaderTimeout uint32             `json:"httpReadHeaderTimeout,omitempty" yaml:"httpReadHeaderTimeout,omitempty"`
	Logger                *zap.SugaredLogger `json:"-" yaml:"-"`
	AccessLogEventFunc    m.LogEventFunc     `json:"-" yaml:"-"`
	RecoveryLogEventFunc  m.LogEventFunc     `json:"-" yaml:"-"`
}

// NewConfig 创建一个新的配置
// NewConfig creates a new config
func NewConfig() *Config {
	return &Config{
		Address:               defaultHttpListemAddress,
		Port:                  defaultHttpListemPort,
		ReleaseMode:           false,
		HttpReadTimeout:       defaultIdleTimeout,
		HttpWriteTimeout:      defaultIdleTimeout,
		HttpReadHeaderTimeout: defaultIdleTimeout,
		Logger:                defaultConsoleLogger.S().Named(defaultLoggerName),
		AccessLogEventFunc:    m.DefaultAcceseEventFunc,
		RecoveryLogEventFunc:  m.DefaultRecoveryEventFunc,
	}
}

// WithSugaredLogger 设置一个新的日志记录器
// WithSugaredLogger sets a new sugared logger
func (c *Config) WithSugaredLogger(logger *zap.SugaredLogger) *Config {
	c.Logger = logger
	return c
}

// WithLogger 设置一个新的日志记录器
// WithLogger sets a new logger
func (c *Config) WithLogger(logger *zap.Logger) *Config {
	c.Logger = logger.Sugar()
	return c
}

// WithAddress 设置一个新的地址
// WithAddress sets a new address
func (c *Config) WithAddress(address string) *Config {
	c.Address = address
	return c
}

// WithPort 设置一个新的端口
// WithPort sets a new port
func (c *Config) WithPort(port uint16) *Config {
	c.Port = port
	return c
}

// WithRelease 设置为 Release 模式
// WithRelease sets to Release mode
func (c *Config) WithRelease() *Config {
	c.ReleaseMode = true
	return c
}

// WithHttpReadTimeout 设置一个新的 Http 读取超时时间
// WithHttpReadTimeout sets a new Http read timeout
func (c *Config) WithHttpReadTimeout(timeout uint32) *Config {
	c.HttpReadTimeout = timeout
	return c
}

// WithHttpWriteTimeout 设置一个新的 Http 写入超时时间
// WithHttpWriteTimeout sets a new Http write timeout
func (c *Config) WithHttpWriteTimeout(timeout uint32) *Config {
	c.HttpWriteTimeout = timeout
	return c
}

// WithHttpReadHeaderTimeout 设置一个新的 Http 读取头部超时时间
// WithHttpReadHeaderTimeout sets a new Http read header timeout
func (c *Config) WithHttpReadHeaderTimeout(timeout uint32) *Config {
	c.HttpReadHeaderTimeout = timeout
	return c
}

func (c *Config) WithAccessLogEventFunc(fn m.LogEventFunc) *Config {
	c.AccessLogEventFunc = fn
	return c
}

func (c *Config) WithRecoveryLogEventFunc(fn m.LogEventFunc) *Config {
	c.RecoveryLogEventFunc = fn
	return c
}

// DefaultConfig 返回一个默认配置
// DefaultConfig returns a default config
func DefaultConfig() *Config {
	return NewConfig()
}

// isConfigValid 检查配置是否有效
// isConfigValid checks if the configuration is valid
func isConfigValid(conf *Config) *Config {
	if conf != nil {
		if len(strings.TrimSpace(conf.Address)) == 0 {
			conf.Address = defaultHttpListemAddress
		}
		if conf.Port <= 0 {
			conf.Port = defaultHttpListemPort
		}
		if conf.HttpReadTimeout <= 0 {
			conf.HttpReadTimeout = defaultIdleTimeout
		}
		if conf.HttpWriteTimeout <= 0 {
			conf.HttpWriteTimeout = defaultIdleTimeout
		}
		if conf.HttpReadHeaderTimeout <= 0 {
			conf.HttpReadHeaderTimeout = defaultIdleTimeout
		}
		if conf.Logger == nil {
			conf.Logger = defaultConsoleLogger.S().Named(defaultLoggerName)
		}
		if conf.AccessLogEventFunc == nil {
			conf.AccessLogEventFunc = m.DefaultAcceseEventFunc
		}
		if conf.RecoveryLogEventFunc == nil {
			conf.RecoveryLogEventFunc = m.DefaultRecoveryEventFunc
		}
	} else {
		conf = NewConfig()
	}

	return conf
}
