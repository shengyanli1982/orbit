package orbit

import (
	"strings"

	com "github.com/shengyanli1982/orbit/common"
	bp "github.com/shengyanli1982/orbit/internal/pool"
	"go.uber.org/zap"
)

var (
	defaultConsoleLogger     = NewLogger(nil) // default console logger
	defaultHttpListenAddress = "127.0.0.0"    // default http listen address
	defaultHttpListenPort    = uint16(8080)   // default http listen port
	defaultIdleTimeout       = uint32(15000)  // http idle timeout
)

// Configuration represents the configuration for the Orbit framework.
type Config struct {
	Address               string             `json:"address,omitempty" yaml:"address,omitempty"`                             // Address to listen on
	Port                  uint16             `json:"port,omitempty" yaml:"port,omitempty"`                                   // Port to listen on
	ReleaseMode           bool               `json:"releaseMode,omitempty" yaml:"releaseMode,omitempty"`                     // Release mode flag
	HttpReadTimeout       uint32             `json:"httpReadTimeout,omitempty" yaml:"httpReadTimeout,omitempty"`             // HTTP read timeout
	HttpWriteTimeout      uint32             `json:"httpWriteTimeout,omitempty" yaml:"httpWriteTimeout,omitempty"`           // HTTP write timeout
	HttpReadHeaderTimeout uint32             `json:"httpReadHeaderTimeout,omitempty" yaml:"httpReadHeaderTimeout,omitempty"` // HTTP read header timeout
	Logger                *zap.SugaredLogger `json:"-" yaml:"-"`                                                             // Logger instance
	AccessLogEventFunc    com.LogEventFunc   `json:"-" yaml:"-"`                                                             // Access log event function
	RecoveryLogEventFunc  com.LogEventFunc   `json:"-" yaml:"-"`                                                             // Recovery log event function
}

// NewConfig creates a new Config instance with default values.
func NewConfig() *Config {
	return &Config{
		Address:               defaultHttpListenAddress,
		Port:                  defaultHttpListenPort,
		ReleaseMode:           false,
		HttpReadTimeout:       defaultIdleTimeout,
		HttpWriteTimeout:      defaultIdleTimeout,
		HttpReadHeaderTimeout: defaultIdleTimeout,
		Logger:                defaultConsoleLogger.GetZapSugaredLogger().Named(defaultLoggerName),
		AccessLogEventFunc:    DefaultAccessEventFunc,
		RecoveryLogEventFunc:  DefaultRecoveryEventFunc,
	}
}

// WithSugaredLogger sets a new sugared logger for the Config instance.
func (c *Config) WithSugaredLogger(logger *zap.SugaredLogger) *Config {
	c.Logger = logger
	return c
}

// WithLogger sets a new logger for the Config instance.
func (c *Config) WithLogger(logger *zap.Logger) *Config {
	c.Logger = logger.Sugar()
	return c
}

// WithAddress sets a new address for the Config instance.
func (c *Config) WithAddress(address string) *Config {
	c.Address = address
	return c
}

// WithPort sets a new port for the Config instance.
func (c *Config) WithPort(port uint16) *Config {
	c.Port = port
	return c
}

// WithRelease sets the Config instance to release mode.
func (c *Config) WithRelease() *Config {
	c.ReleaseMode = true
	return c
}

// WithHttpReadTimeout sets a new HTTP read timeout for the Config instance.
func (c *Config) WithHttpReadTimeout(timeout uint32) *Config {
	c.HttpReadTimeout = timeout
	return c
}

// WithHttpWriteTimeout sets a new HTTP write timeout for the Config instance.
func (c *Config) WithHttpWriteTimeout(timeout uint32) *Config {
	c.HttpWriteTimeout = timeout
	return c
}

// WithHttpReadHeaderTimeout sets a new HTTP read header timeout for the Config instance.
func (c *Config) WithHttpReadHeaderTimeout(timeout uint32) *Config {
	c.HttpReadHeaderTimeout = timeout
	return c
}

// WithAccessLogEventFunc sets a new access log event function for the Config instance.
func (c *Config) WithAccessLogEventFunc(fn com.LogEventFunc) *Config {
	c.AccessLogEventFunc = fn
	return c
}

// WithRecoveryLogEventFunc sets a new recovery log event function for the Config instance.
func (c *Config) WithRecoveryLogEventFunc(fn com.LogEventFunc) *Config {
	c.RecoveryLogEventFunc = fn
	return c
}

// DefaultConfig returns a new Config instance with default values.
func DefaultConfig() *Config {
	return NewConfig()
}

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
		if conf.Logger == nil {
			conf.Logger = defaultConsoleLogger.GetZapSugaredLogger().Named(defaultLoggerName)
		}
		if conf.AccessLogEventFunc == nil {
			conf.AccessLogEventFunc = DefaultAccessEventFunc
		}
		if conf.RecoveryLogEventFunc == nil {
			conf.RecoveryLogEventFunc = DefaultRecoveryEventFunc
		}
	} else {
		conf = NewConfig()
	}

	return conf
}

// DefaultAccessEventFunc is the default access log event function.
func DefaultAccessEventFunc(logger *zap.SugaredLogger, event *bp.LogEvent) {
	logger.Infow(
		event.Message,
		"id", event.ID,
		"ip", event.IP,
		"endpoint", event.EndPoint,
		"path", event.Path,
		"method", event.Method,
		"code", event.Code,
		"status", event.Status,
		"latency", event.Latency,
		"agent", event.Agent,
		"query", event.ReqQuery,
		"reqContentType", event.ReqContentType,
		"reqBody", event.ReqBody,
	)
}

// DefaultRecoveryEventFunc is the default recovery log event function.
func DefaultRecoveryEventFunc(logger *zap.SugaredLogger, event *bp.LogEvent) {
	logger.Errorw(
		event.Message,
		"id", event.ID,
		"ip", event.IP,
		"endpoint", event.EndPoint,
		"path", event.Path,
		"method", event.Method,
		"code", event.Code,
		"status", event.Status,
		"latency", event.Latency,
		"agent", event.Agent,
		"query", event.ReqQuery,
		"reqContentType", event.ReqContentType,
		"reqBody", event.ReqBody,
		"error", event.Error,
		"errorStack", event.ErrorStack,
	)
}
