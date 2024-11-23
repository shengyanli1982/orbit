package orbit

import (
	"strings"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/utils/log"
)

var (
	// defaultHttpListenAddress 是默认的 HTTP 监听地址，设置为 "127.0.0.1"。
	// defaultHttpListenAddress is the default HTTP listen address, set to "127.0.0.1".
	defaultHttpListenAddress = "127.0.0.1"

	// defaultHttpListenPort 是默认的 HTTP 监听端口，设置为 8080。
	// defaultHttpListenPort is the default HTTP listen port, set to 8080.
	defaultHttpListenPort = uint16(8080)

	// defaultIdleTimeout 是 HTTP 空闲超时时间，设置为 15000 毫秒。
	// defaultIdleTimeout is the HTTP idle timeout, set to 15000 milliseconds.
	defaultIdleTimeout = uint32(15000)
)

// Config 结构体表示 Orbit 框架的配置。
// The Config struct represents the configuration for the Orbit framework.
type Config struct {
	// Address 是监听的地址。
	// Address is the address to listen on.
	Address string `json:"address,omitempty" yaml:"address,omitempty"`

	// Port 是监听的端口。
	// Port is the port to listen on.
	Port uint16 `json:"port,omitempty" yaml:"port,omitempty"`

	// ReleaseMode 是发布模式标志。
	// ReleaseMode is the release mode flag.
	ReleaseMode bool `json:"releaseMode,omitempty" yaml:"releaseMode,omitempty"`

	// HttpReadTimeout 是 HTTP 读取超时时间。
	// HttpReadTimeout is the HTTP read timeout.
	HttpReadTimeout uint32 `json:"httpReadTimeout,omitempty" yaml:"httpReadTimeout,omitempty"`

	// HttpWriteTimeout 是 HTTP 写入超时时间。
	// HttpWriteTimeout is the HTTP write timeout.
	HttpWriteTimeout uint32 `json:"httpWriteTimeout,omitempty" yaml:"httpWriteTimeout,omitempty"`

	// HttpReadHeaderTimeout 是 HTTP 读取头部超时时间。
	// HttpReadHeaderTimeout is the HTTP read header timeout.
	HttpReadHeaderTimeout uint32 `json:"httpReadHeaderTimeout,omitempty" yaml:"httpReadHeaderTimeout,omitempty"`

	// logger 是日志实例。
	// logger is the logger instance.
	logger *logr.Logger `json:"-" yaml:"-"`

	// accessLogEventFunc 是访问日志事件函数。
	// accessLogEventFunc is the access log event function.
	accessLogEventFunc com.LogEventFunc `json:"-" yaml:"-"`

	// recoveryLogEventFunc 是恢复日志事件函数。
	// recoveryLogEventFunc is the recovery log event function.
	recoveryLogEventFunc com.LogEventFunc `json:"-" yaml:"-"`

	// prometheusRegistry 是 Prometheus 注册表。
	// prometheusRegistry is the Prometheus registry.
	prometheusRegistry *prometheus.Registry `json:"-" yaml:"-"`
}

// NewConfig 创建一个新的 Config 实例，并使用默认值。
// NewConfig creates a new Config instance with default values.
func NewConfig() *Config {
	return &Config{
		// Address 是默认的监听地址。
		// Address is the default address to listen on.
		Address: defaultHttpListenAddress,

		// Port 是默认的监听端口。
		// Port is the default port to listen on.
		Port: defaultHttpListenPort,

		// ReleaseMode 是默认的发布模式标志，设置为 false。
		// ReleaseMode is the default release mode flag, set to false.
		ReleaseMode: false,

		// HttpReadTimeout 是默认的 HTTP 读取超时时间。
		// HttpReadTimeout is the default HTTP read timeout.
		HttpReadTimeout: defaultIdleTimeout,

		// HttpWriteTimeout 是默认的 HTTP 写入超时时间。
		// HttpWriteTimeout is the default HTTP write timeout.
		HttpWriteTimeout: defaultIdleTimeout,

		// HttpReadHeaderTimeout 是默认的 HTTP 读取头部超时时间。
		// HttpReadHeaderTimeout is the default HTTP read header timeout.
		HttpReadHeaderTimeout: defaultIdleTimeout,

		// logger 是默认的日志实例。
		// logger is the default logger instance.
		logger: &com.DefaultLogrLogger,

		// accessLogEventFunc 是默认的访问日志事件函数。
		// accessLogEventFunc is the default access log event function.
		accessLogEventFunc: log.DefaultAccessEventFunc,

		// recoveryLogEventFunc 是默认的恢复日志事件函数。
		// recoveryLogEventFunc is the default recovery log event function.
		recoveryLogEventFunc: log.DefaultRecoveryEventFunc,

		// prometheusRegistry 是默认的 Prometheus 注册表。
		// prometheusRegistry is the default Prometheus registry.
		prometheusRegistry: prometheus.DefaultRegisterer.(*prometheus.Registry),
	}
}

func (c *Config) WithLogger(logger *logr.Logger) *Config {
	c.logger = logger
	return c
}

// WithAddress 为 Config 实例设置一个新的监听地址。
// WithAddress sets a new address for the Config instance.
func (c *Config) WithAddress(address string) *Config {
	c.Address = address
	return c
}

// WithPort 为 Config 实例设置一个新的监听端口。
// WithPort sets a new port for the Config instance.
func (c *Config) WithPort(port uint16) *Config {
	c.Port = port
	return c
}

// WithRelease 将 Config 实例设置为发布模式。
// WithRelease sets the Config instance to release mode.
func (c *Config) WithRelease() *Config {
	c.ReleaseMode = true
	return c
}

// WithHttpReadTimeout 为 Config 实例设置一个新的 HTTP 读取超时时间。
// WithHttpReadTimeout sets a new HTTP read timeout for the Config instance.
func (c *Config) WithHttpReadTimeout(timeout uint32) *Config {
	c.HttpReadTimeout = timeout
	return c
}

// WithHttpWriteTimeout 为 Config 实例设置一个新的 HTTP 写入超时时间。
// WithHttpWriteTimeout sets a new HTTP write timeout for the Config instance.
func (c *Config) WithHttpWriteTimeout(timeout uint32) *Config {
	c.HttpWriteTimeout = timeout
	return c
}

// WithHttpReadHeaderTimeout 为 Config 实例设置一个新的 HTTP 读取头部超时时间。
// WithHttpReadHeaderTimeout sets a new HTTP read header timeout for the Config instance.
func (c *Config) WithHttpReadHeaderTimeout(timeout uint32) *Config {
	c.HttpReadHeaderTimeout = timeout
	return c
}

// WithAccessLogEventFunc 为 Config 实例设置一个新的访问日志事件函数。
// WithAccessLogEventFunc sets a new access log event function for the Config instance.
func (c *Config) WithAccessLogEventFunc(fn com.LogEventFunc) *Config {
	c.accessLogEventFunc = fn
	return c
}

// WithRecoveryLogEventFunc 为 Config 实例设置一个新的恢复日志事件函数。
// WithRecoveryLogEventFunc sets a new recovery log event function for the Config instance.
func (c *Config) WithRecoveryLogEventFunc(fn com.LogEventFunc) *Config {
	c.recoveryLogEventFunc = fn
	return c
}

// WithPrometheusRegistry 为 Config 实例设置一个新的 Prometheus 注册表。
// WithPrometheusRegistry sets a new Prometheus registry for the Config instance.
func (c *Config) WithPrometheusRegistry(registry *prometheus.Registry) *Config {
	c.prometheusRegistry = registry
	return c
}

// DefaultConfig 返回一个带有默认值的新 Config 实例。
// DefaultConfig returns a new Config instance with default values.
func DefaultConfig() *Config {
	return NewConfig()
}

// isConfigValid 检查配置是否有效，如果需要，应用默认值。
// isConfigValid checks if the configuration is valid and applies default values if necessary.
func isConfigValid(conf *Config) *Config {
	if conf != nil {
		// 如果 Address 为空或者只包含空格，设置为默认的 HTTP 监听地址。
		// If Address is empty or only contains spaces, set it to the default HTTP listen address.
		if len(strings.TrimSpace(conf.Address)) == 0 {
			conf.Address = defaultHttpListenAddress
		}

		// 如果 Port 小于或等于 0，设置为默认的 HTTP 监听端口。
		// If Port is less than or equal to 0, set it to the default HTTP listen port.
		if conf.Port <= 0 {
			conf.Port = defaultHttpListenPort
		}

		// 如果 HttpReadTimeout 小于或等于 0，设置为默认的空闲超时时间。
		// If HttpReadTimeout is less than or equal to 0, set it to the default idle timeout.
		if conf.HttpReadTimeout <= 0 {
			conf.HttpReadTimeout = defaultIdleTimeout
		}

		// 如果 HttpWriteTimeout 小于或等于 0，设置为默认的空闲超时时间。
		// If HttpWriteTimeout is less than or equal to 0, set it to the default idle timeout.
		if conf.HttpWriteTimeout <= 0 {
			conf.HttpWriteTimeout = defaultIdleTimeout
		}

		// 如果 HttpReadHeaderTimeout 小于或等于 0，设置为默认的空闲超时时间。
		// If HttpReadHeaderTimeout is less than or equal to 0, set it to the default idle timeout.
		if conf.HttpReadHeaderTimeout <= 0 {
			conf.HttpReadHeaderTimeout = defaultIdleTimeout
		}

		// 如果 logger 为 nil，设置为默认的 DefaultLogrLogger
		// If logger is nil, set it to the default DefaultLogrLogger.
		if conf.logger == nil {
			conf.logger = &com.DefaultLogrLogger
		}

		// 如果 accessLogEventFunc 为 nil，设置为默认的访问日志事件函数。
		// If accessLogEventFunc is nil, set it to the default access log event function.
		if conf.accessLogEventFunc == nil {
			conf.accessLogEventFunc = log.DefaultAccessEventFunc
		}

		// 如果 recoveryLogEventFunc 为 nil，设置为默认的恢复日志事件函数。
		// If recoveryLogEventFunc is nil, set it to the default recovery log event function.
		if conf.recoveryLogEventFunc == nil {
			conf.recoveryLogEventFunc = log.DefaultRecoveryEventFunc
		}

		// 如果 prometheusRegistry 为 nil，设置为默认的 Prometheus 注册表。
		// If prometheusRegistry is nil, set it to the default Prometheus registry.
		if conf.prometheusRegistry == nil {
			conf.prometheusRegistry = prometheus.DefaultRegisterer.(*prometheus.Registry)
		}
	} else {
		// 如果 conf 为 nil，创建一个新的 Config 实例。
		// If conf is nil, create a new Config instance.
		conf = DefaultConfig()
	}

	// 返回验证和可能修改过的配置。
	// Return the validated and possibly modified configuration.
	return conf
}
