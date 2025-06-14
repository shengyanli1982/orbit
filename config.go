package orbit

import (
	"strings"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/utils/log"
)

// 默认配置值
var (
	defaultHttpListenAddress = com.DefaultHttpListenAddress     // 默认HTTP监听地址
	defaultHttpListenPort    = com.DefaultHttpListenPort        // 默认HTTP监听端口
	defaultIdleTimeout       = com.DefaultHttpIdleTimeoutMillis // 默认空闲超时时间（毫秒）
)

// Config 结构体定义了服务器的配置选项
type Config struct {
	Address               string               `json:"address,omitempty" yaml:"address,omitempty"`                             // HTTP服务器监听地址
	Port                  uint16               `json:"port,omitempty" yaml:"port,omitempty"`                                   // HTTP服务器监听端口
	ReleaseMode           bool                 `json:"releaseMode,omitempty" yaml:"releaseMode,omitempty"`                     // 是否为发布模式
	HttpReadTimeout       uint32               `json:"httpReadTimeout,omitempty" yaml:"httpReadTimeout,omitempty"`             // HTTP读取超时时间
	HttpWriteTimeout      uint32               `json:"httpWriteTimeout,omitempty" yaml:"httpWriteTimeout,omitempty"`           // HTTP写入超时时间
	HttpReadHeaderTimeout uint32               `json:"httpReadHeaderTimeout,omitempty" yaml:"httpReadHeaderTimeout,omitempty"` // HTTP读取头部超时时间
	HttpIdleTimeout       uint32               `json:"httpIdleTimeout,omitempty" yaml:"httpIdleTimeout,omitempty"`             // HTTP空闲超时时间
	MaxHeaderBytes        uint32               `json:"maxHeaderBytes,omitempty" yaml:"maxHeaderBytes,omitempty"`               // HTTP最大头部字节数
	logger                *logr.Logger         `json:"-" yaml:"-"`                                                             // 日志记录器
	accessLogEventFunc    com.LogEventFunc     `json:"-" yaml:"-"`                                                             // 访问日志事件处理函数
	recoveryLogEventFunc  com.LogEventFunc     `json:"-" yaml:"-"`                                                             // 恢复日志事件处理函数
	prometheusRegistry    *prometheus.Registry `json:"-" yaml:"-"`                                                             // Prometheus注册表
}

// 创建并返回一个新的默认配置实例
func NewConfig() *Config {
	return &Config{
		Address:               defaultHttpListenAddress,
		Port:                  defaultHttpListenPort,
		ReleaseMode:           false,
		HttpReadTimeout:       defaultIdleTimeout,
		HttpWriteTimeout:      defaultIdleTimeout,
		HttpReadHeaderTimeout: defaultIdleTimeout,
		HttpIdleTimeout:       defaultIdleTimeout,
		MaxHeaderBytes:        defaultIdleTimeout,
		logger:                &com.DefaultLogrLogger,
		accessLogEventFunc:    log.DefaultAccessEventFunc,
		recoveryLogEventFunc:  log.DefaultRecoveryEventFunc,
		prometheusRegistry:    prometheus.DefaultRegisterer.(*prometheus.Registry),
	}
}

// 设置日志记录器
func (c *Config) WithLogger(logger *logr.Logger) *Config {
	c.logger = logger
	return c
}

// 设置HTTP监听地址
func (c *Config) WithAddress(address string) *Config {
	c.Address = address
	return c
}

// 设置HTTP监听端口
func (c *Config) WithPort(port uint16) *Config {
	c.Port = port
	return c
}

// 启用发布模式
func (c *Config) WithRelease() *Config {
	c.ReleaseMode = true
	return c
}

// 设置HTTP读取超时时间
func (c *Config) WithHttpReadTimeout(timeout uint32) *Config {
	c.HttpReadTimeout = timeout
	return c
}

// 设置HTTP写入超时时间
func (c *Config) WithHttpWriteTimeout(timeout uint32) *Config {
	c.HttpWriteTimeout = timeout
	return c
}

// 设置HTTP读取头部超时时间
func (c *Config) WithHttpReadHeaderTimeout(timeout uint32) *Config {
	c.HttpReadHeaderTimeout = timeout
	return c
}

// 设置HTTP空闲超时时间
func (c *Config) WithHttpIdleTimeout(timeout uint32) *Config {
	c.HttpIdleTimeout = timeout
	return c
}

// 设置HTTP最大头部字节数
func (c *Config) WithMaxHeaderBytes(bytes uint32) *Config {
	c.MaxHeaderBytes = bytes
	return c
}

// 设置访问日志事件处理函数
func (c *Config) WithAccessLogEventFunc(fn com.LogEventFunc) *Config {
	c.accessLogEventFunc = fn
	return c
}

// 设置恢复日志事件处理函数
func (c *Config) WithRecoveryLogEventFunc(fn com.LogEventFunc) *Config {
	c.recoveryLogEventFunc = fn
	return c
}

// 设置Prometheus注册表
func (c *Config) WithPrometheusRegistry(registry *prometheus.Registry) *Config {
	c.prometheusRegistry = registry
	return c
}

// 返回默认配置实例
func DefaultConfig() *Config {
	return NewConfig()
}

// 验证配置的有效性，并设置默认值
func isConfigValid(conf *Config) *Config {
	// 如果配置为空，返回默认配置
	if conf == nil {
		return DefaultConfig()
	}

	// 使用默认配置作为基准进行比较和设置
	defaultConf := DefaultConfig()

	// 验证并设置基本网络配置
	if strings.TrimSpace(conf.Address) == "" {
		conf.Address = defaultConf.Address
	}
	if conf.Port == 0 {
		conf.Port = defaultConf.Port
	}

	// 验证并设置超时配置
	if conf.HttpReadTimeout == 0 {
		conf.HttpReadTimeout = defaultConf.HttpReadTimeout
	}
	if conf.HttpWriteTimeout == 0 {
		conf.HttpWriteTimeout = defaultConf.HttpWriteTimeout
	}
	if conf.HttpReadHeaderTimeout == 0 {
		conf.HttpReadHeaderTimeout = defaultConf.HttpReadHeaderTimeout
	}
	if conf.HttpIdleTimeout == 0 {
		conf.HttpIdleTimeout = defaultConf.HttpIdleTimeout
	}
	if conf.MaxHeaderBytes == 0 {
		conf.MaxHeaderBytes = defaultConf.MaxHeaderBytes
	}

	// 验证并设置日志和事件处理配置
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
	if conf.prometheusRegistry == nil {
		conf.prometheusRegistry = defaultConf.prometheusRegistry
	}

	return conf
}
