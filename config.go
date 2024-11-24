package orbit

import (
	"strings"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/utils/log"
)

// 默认配置值
// Default configuration values
var (
	defaultHttpListenAddress = "127.0.0.1"   // 默认HTTP监听地址 (Default HTTP listen address)
	defaultHttpListenPort    = uint16(8080)  // 默认HTTP监听端口 (Default HTTP listen port)
	defaultIdleTimeout       = uint32(15000) // 默认空闲超时时间（毫秒） (Default idle timeout in milliseconds)
)

// Config 结构体定义了服务器的配置选项
// The Config struct defines the server configuration options
type Config struct {
	Address               string               `json:"address,omitempty" yaml:"address,omitempty"`                             // HTTP服务器监听地址 (HTTP server listen address)
	Port                  uint16               `json:"port,omitempty" yaml:"port,omitempty"`                                   // HTTP服务器监听端口 (HTTP server listen port)
	ReleaseMode           bool                 `json:"releaseMode,omitempty" yaml:"releaseMode,omitempty"`                     // 是否为发布模式 (Whether in release mode)
	HttpReadTimeout       uint32               `json:"httpReadTimeout,omitempty" yaml:"httpReadTimeout,omitempty"`             // HTTP读取超时时间 (HTTP read timeout)
	HttpWriteTimeout      uint32               `json:"httpWriteTimeout,omitempty" yaml:"httpWriteTimeout,omitempty"`           // HTTP写入超时时间 (HTTP write timeout)
	HttpReadHeaderTimeout uint32               `json:"httpReadHeaderTimeout,omitempty" yaml:"httpReadHeaderTimeout,omitempty"` // HTTP读取头部超时时间 (HTTP read header timeout)
	logger                *logr.Logger         `json:"-" yaml:"-"`                                                             // 日志记录器 (Logger instance)
	accessLogEventFunc    com.LogEventFunc     `json:"-" yaml:"-"`                                                             // 访问日志事件处理函数 (Access log event handler function)
	recoveryLogEventFunc  com.LogEventFunc     `json:"-" yaml:"-"`                                                             // 恢复日志事件处理函数 (Recovery log event handler function)
	prometheusRegistry    *prometheus.Registry `json:"-" yaml:"-"`                                                             // Prometheus注册表 (Prometheus registry)
}

// NewConfig 函数创建并返回一个新的默认配置实例
// The NewConfig function creates and returns a new default configuration instance
func NewConfig() *Config {
	return &Config{
		Address:               defaultHttpListenAddress,
		Port:                  defaultHttpListenPort,
		ReleaseMode:           false,
		HttpReadTimeout:       defaultIdleTimeout,
		HttpWriteTimeout:      defaultIdleTimeout,
		HttpReadHeaderTimeout: defaultIdleTimeout,
		logger:                &com.DefaultLogrLogger,
		accessLogEventFunc:    log.DefaultAccessEventFunc,
		recoveryLogEventFunc:  log.DefaultRecoveryEventFunc,
		prometheusRegistry:    prometheus.DefaultRegisterer.(*prometheus.Registry),
	}
}

// WithLogger 方法设置日志记录器
// The WithLogger method sets the logger
func (c *Config) WithLogger(logger *logr.Logger) *Config {
	c.logger = logger
	return c
}

// WithAddress 方法设置HTTP监听地址
// The WithAddress method sets the HTTP listen address
func (c *Config) WithAddress(address string) *Config {
	c.Address = address
	return c
}

// WithPort 方法设置HTTP监听端口
// The WithPort method sets the HTTP listen port
func (c *Config) WithPort(port uint16) *Config {
	c.Port = port
	return c
}

// WithRelease 方法启用发布模式
// The WithRelease method enables release mode
func (c *Config) WithRelease() *Config {
	c.ReleaseMode = true
	return c
}

// WithHttpReadTimeout 方法设置HTTP读取超时时间
// The WithHttpReadTimeout method sets the HTTP read timeout
func (c *Config) WithHttpReadTimeout(timeout uint32) *Config {
	c.HttpReadTimeout = timeout
	return c
}

// WithHttpWriteTimeout 方法设置HTTP写入超时时间
// The WithHttpWriteTimeout method sets the HTTP write timeout
func (c *Config) WithHttpWriteTimeout(timeout uint32) *Config {
	c.HttpWriteTimeout = timeout
	return c
}

// WithHttpReadHeaderTimeout 方法设置HTTP读取头部超时时间
// The WithHttpReadHeaderTimeout method sets the HTTP read header timeout
func (c *Config) WithHttpReadHeaderTimeout(timeout uint32) *Config {
	c.HttpReadHeaderTimeout = timeout
	return c
}

// WithAccessLogEventFunc 方法设置访问日志事件处理函数
// The WithAccessLogEventFunc method sets the access log event handler function
func (c *Config) WithAccessLogEventFunc(fn com.LogEventFunc) *Config {
	c.accessLogEventFunc = fn
	return c
}

// WithRecoveryLogEventFunc 方法设置恢复日志事件处理函数
// The WithRecoveryLogEventFunc method sets the recovery log event handler function
func (c *Config) WithRecoveryLogEventFunc(fn com.LogEventFunc) *Config {
	c.recoveryLogEventFunc = fn
	return c
}

// WithPrometheusRegistry 方法设置Prometheus注册表
// The WithPrometheusRegistry method sets the Prometheus registry
func (c *Config) WithPrometheusRegistry(registry *prometheus.Registry) *Config {
	c.prometheusRegistry = registry
	return c
}

// DefaultConfig 函数返回默认配置实例
// The DefaultConfig function returns a default configuration instance
func DefaultConfig() *Config {
	return NewConfig()
}

// isConfigValid 函数验证配置的有效性，并设置默认值
// The isConfigValid function validates the configuration and sets default values
func isConfigValid(conf *Config) *Config {
	if conf != nil {
		// 检查并设置默认地址
		// Check and set default address
		if len(strings.TrimSpace(conf.Address)) == 0 {
			conf.Address = defaultHttpListenAddress
		}

		// 检查并设置默认端口
		// Check and set default port
		if conf.Port <= 0 {
			conf.Port = defaultHttpListenPort
		}

		// 检查并设置默认超时时间
		// Check and set default timeout values
		if conf.HttpReadTimeout <= 0 {
			conf.HttpReadTimeout = defaultIdleTimeout
		}
		if conf.HttpWriteTimeout <= 0 {
			conf.HttpWriteTimeout = defaultIdleTimeout
		}
		if conf.HttpReadHeaderTimeout <= 0 {
			conf.HttpReadHeaderTimeout = defaultIdleTimeout
		}

		// 检查并设置默认日志记录器和事件处理函数
		// Check and set default logger and event handlers
		if conf.logger == nil {
			conf.logger = &com.DefaultLogrLogger
		}
		if conf.accessLogEventFunc == nil {
			conf.accessLogEventFunc = log.DefaultAccessEventFunc
		}
		if conf.recoveryLogEventFunc == nil {
			conf.recoveryLogEventFunc = log.DefaultRecoveryEventFunc
		}

		// 检查并设置默认Prometheus注册表
		// Check and set default Prometheus registry
		if conf.prometheusRegistry == nil {
			conf.prometheusRegistry = prometheus.DefaultRegisterer.(*prometheus.Registry)
		}
	} else {
		// 如果配置为空，则创建默认配置
		// If configuration is nil, create default configuration
		conf = DefaultConfig()
	}

	return conf
}
