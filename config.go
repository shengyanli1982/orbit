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

// WithRelease 方法启用发布���式
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
	// 如果配置为空，返回默认配置
	// If configuration is nil, return default configuration
	if conf == nil {
		return DefaultConfig()
	}

	// 使用默认配置作为基准进行比较和设置
	// Use default configuration as baseline for comparison and setting
	defaultConf := DefaultConfig()

	// 验证并设置基本网络配置
	// Validate and set basic network configuration
	if strings.TrimSpace(conf.Address) == "" {
		conf.Address = defaultConf.Address
	}
	if conf.Port == 0 {
		conf.Port = defaultConf.Port
	}

	// 验证并设置超时配置
	// Validate and set timeout configuration
	if conf.HttpReadTimeout == 0 {
		conf.HttpReadTimeout = defaultConf.HttpReadTimeout
	}
	if conf.HttpWriteTimeout == 0 {
		conf.HttpWriteTimeout = defaultConf.HttpWriteTimeout
	}
	if conf.HttpReadHeaderTimeout == 0 {
		conf.HttpReadHeaderTimeout = defaultConf.HttpReadHeaderTimeout
	}

	// 验证并设置日志和事件处理配置
	// Validate and set logging and event handling configuration
	if conf.logger == nil {
		conf.logger = defaultConf.logger
	}
	if conf.accessLogEventFunc == nil {
		conf.accessLogEventFunc = defaultConf.accessLogEventFunc
	}
	if conf.recoveryLogEventFunc == nil {
		conf.recoveryLogEventFunc = defaultConf.recoveryLogEventFunc
	}

	// 验证并设置监控配置
	// Validate and set monitoring configuration
	if conf.prometheusRegistry == nil {
		conf.prometheusRegistry = defaultConf.prometheusRegistry
	}

	return conf
}
