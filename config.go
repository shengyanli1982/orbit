package orbit

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	ilog "github.com/shengyanli1982/orbit/internal/log"
	"go.uber.org/zap"
)

var (
	// defaultHttpListenAddress 是默认的 http 监听地址
	// defaultHttpListenAddress is the default http listen address
	defaultHttpListenAddress = "127.0.0.1"

	// defaultHttpListenPort 是默认的 http 监听端口
	// defaultHttpListenPort is the default http listen port
	defaultHttpListenPort = uint16(8080)

	// defaultIdleTimeout 是默认的 http 空闲超时时间
	// defaultIdleTimeout is the default http idle timeout
	defaultIdleTimeout = uint32(15000)
)

// Configuration 表示 Orbit 框架的配置。
// Configuration represents the configuration for the Orbit framework.
type Config struct {
	Address               string               `json:"address,omitempty" yaml:"address,omitempty"`                             // Address to listen on
	Port                  uint16               `json:"port,omitempty" yaml:"port,omitempty"`                                   // Port to listen on
	ReleaseMode           bool                 `json:"releaseMode,omitempty" yaml:"releaseMode,omitempty"`                     // Release mode flag
	HttpReadTimeout       uint32               `json:"httpReadTimeout,omitempty" yaml:"httpReadTimeout,omitempty"`             // HTTP read timeout
	HttpWriteTimeout      uint32               `json:"httpWriteTimeout,omitempty" yaml:"httpWriteTimeout,omitempty"`           // HTTP write timeout
	HttpReadHeaderTimeout uint32               `json:"httpReadHeaderTimeout,omitempty" yaml:"httpReadHeaderTimeout,omitempty"` // HTTP read header timeout
	logger                *zap.SugaredLogger   `json:"-" yaml:"-"`                                                             // Logger instance
	accessLogEventFunc    com.LogEventFunc     `json:"-" yaml:"-"`                                                             // Access log event function
	recoveryLogEventFunc  com.LogEventFunc     `json:"-" yaml:"-"`                                                             // Recovery log event function
	prometheusRegistry    *prometheus.Registry `json:"-" yaml:"-"`                                                             // Prometheus registry
}

// NewConfig 创建一个带有默认值 Config 实例。
// NewConfig creates a new Config instance with default values.
func NewConfig() *Config {
	return &Config{
		Address:               defaultHttpListenAddress,                            // Default address to listen on
		Port:                  defaultHttpListenPort,                               // Default port to listen on
		ReleaseMode:           false,                                               // Default release mode flag
		HttpReadTimeout:       defaultIdleTimeout,                                  // Default HTTP read timeout
		HttpWriteTimeout:      defaultIdleTimeout,                                  // Default HTTP write timeout
		HttpReadHeaderTimeout: defaultIdleTimeout,                                  // Default HTTP read header timeout
		logger:                com.DefaultSugeredLogger,                            // Default logger instance
		accessLogEventFunc:    ilog.DefaultAccessEventFunc,                         // Default access log event function
		recoveryLogEventFunc:  ilog.DefaultRecoveryEventFunc,                       // Default recovery log event function
		prometheusRegistry:    prometheus.DefaultRegisterer.(*prometheus.Registry), // Default Prometheus registry
	}
}

// WithSugaredLogger 设置 Config 实例 sugared logger。
// WithSugaredLogger sets a new sugared logger for the Config instance.
func (c *Config) WithSugaredLogger(logger *zap.SugaredLogger) *Config {
	c.logger = logger
	return c
}

// WithLogger 设置 Config 实例 logger。
// WithLogger sets a new logger for the Config instance.
func (c *Config) WithLogger(logger *zap.Logger) *Config {
	c.logger = logger.Sugar()
	return c
}

// WithAddress 设置 Config 实例监听地址。
// WithAddress sets a new address for the Config instance.
func (c *Config) WithAddress(address string) *Config {
	c.Address = address
	return c
}

// WithPort 设置 Config 实例监听端口。
// WithPort sets a new port for the Config instance.
func (c *Config) WithPort(port uint16) *Config {
	c.Port = port
	return c
}

// WithRelease 设置 Config 实例为 release 模式。
// WithRelease sets the Config instance to release mode.
func (c *Config) WithRelease() *Config {
	c.ReleaseMode = true
	return c
}

// WithHttpReadTimeout 设置 Config 实例 HTTP 读取超时时间。
// WithHttpReadTimeout sets a new HTTP read timeout for the Config instance.
func (c *Config) WithHttpReadTimeout(timeout uint32) *Config {
	c.HttpReadTimeout = timeout
	return c
}

// WithHttpWriteTimeout 设置 Config 实例 HTTP 写入超时时间。
// WithHttpWriteTimeout sets a new HTTP write timeout for the Config instance.
func (c *Config) WithHttpWriteTimeout(timeout uint32) *Config {
	c.HttpWriteTimeout = timeout
	return c
}

// WithHttpReadHeaderTimeout 设置 Config 实例 HTTP 读取头部超时时间。
// WithHttpReadHeaderTimeout sets a new HTTP read header timeout for the Config instance.
func (c *Config) WithHttpReadHeaderTimeout(timeout uint32) *Config {
	c.HttpReadHeaderTimeout = timeout
	return c
}

// WithAccessLogEventFunc 设置 Config 实例访问日志事件函数。
// WithAccessLogEventFunc sets a new access log event function for the Config instance.
func (c *Config) WithAccessLogEventFunc(fn com.LogEventFunc) *Config {
	c.accessLogEventFunc = fn
	return c
}

// WithRecoveryLogEventFunc 设置 Config 实例恢复日志事件函数。
// WithRecoveryLogEventFunc sets a new recovery log event function for the Config instance.
func (c *Config) WithRecoveryLogEventFunc(fn com.LogEventFunc) *Config {
	c.recoveryLogEventFunc = fn
	return c
}

// WithPrometheusRegistry 设置 Config 实例 Prometheus 注册器。
// WithPrometheusRegistry sets a new Prometheus registry for the Config instance.
func (c *Config) WithPrometheusRegistry(registry *prometheus.Registry) *Config {
	c.prometheusRegistry = registry
	return c
}

// WithDefaultPrometheusRegistry 设置 Config 实例默认 Prometheus 注册器。
// DefaultConfig returns a new Config instance with default values.
func DefaultConfig() *Config {
	return NewConfig()
}

// WithDefaultPrometheusRegistry 设置 Config 实例默认 Prometheus 注册器。
// isConfigValid checks if the configuration is valid and applies default values if necessary.
func isConfigValid(conf *Config) *Config {
	if conf != nil {
		if len(strings.TrimSpace(conf.Address)) == 0 {
			conf.Address = defaultHttpListenAddress
		}
		if conf.Port <= 0 {
			conf.Port = defaultHttpListenPort
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
		if conf.logger == nil {
			conf.logger = com.DefaultSugeredLogger
		}
		if conf.accessLogEventFunc == nil {
			conf.accessLogEventFunc = ilog.DefaultAccessEventFunc
		}
		if conf.recoveryLogEventFunc == nil {
			conf.recoveryLogEventFunc = ilog.DefaultRecoveryEventFunc
		}
		if conf.prometheusRegistry == nil {
			conf.prometheusRegistry = prometheus.DefaultRegisterer.(*prometheus.Registry)
		}
	} else {
		conf = DefaultConfig()
	}

	return conf
}
