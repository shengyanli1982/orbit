package orbit

import (
	"strings"
	"time"

	com "github.com/shengyanli1982/orbit/common"
	bp "github.com/shengyanli1982/orbit/internal/pool"
	"go.uber.org/zap"
)

var (
	defaultConsoleLogger     = NewLogger(nil)                            // default console logger
	defaultHttpListenAddress = "127.0.0.0"                               // default http listen address
	defaultHttpListenPort    = uint16(8080)                              // default http listen port
	defaultIdleTimeout       = uint32((15 * time.Second).Milliseconds()) // http idle timeout
)

// Configuration
type Config struct {
	Address               string             `json:"address,omitempty" yaml:"address,omitempty"`
	Port                  uint16             `json:"port,omitempty" yaml:"port,omitempty"`
	ReleaseMode           bool               `json:"releaseMode,omitempty" yaml:"releaseMode,omitempty"`
	HttpReadTimeout       uint32             `json:"httpReadTimeout,omitempty" yaml:"httpReadTimeout,omitempty"`
	HttpWriteTimeout      uint32             `json:"httpWriteTimeout,omitempty" yaml:"httpWriteTimeout,omitempty"`
	HttpReadHeaderTimeout uint32             `json:"httpReadHeaderTimeout,omitempty" yaml:"httpReadHeaderTimeout,omitempty"`
	Logger                *zap.SugaredLogger `json:"-" yaml:"-"`
	AccessLogEventFunc    com.LogEventFunc   `json:"-" yaml:"-"`
	RecoveryLogEventFunc  com.LogEventFunc   `json:"-" yaml:"-"`
}

// NewConfig creates a new config
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

// WithSugaredLogger sets a new sugared logger
func (c *Config) WithSugaredLogger(logger *zap.SugaredLogger) *Config {
	c.Logger = logger
	return c
}

// WithLogger sets a new logger
func (c *Config) WithLogger(logger *zap.Logger) *Config {
	c.Logger = logger.Sugar()
	return c
}

// WithAddress sets a new address
func (c *Config) WithAddress(address string) *Config {
	c.Address = address
	return c
}

// WithPort sets a new port
func (c *Config) WithPort(port uint16) *Config {
	c.Port = port
	return c
}

// WithRelease sets to Release mode
func (c *Config) WithRelease() *Config {
	c.ReleaseMode = true
	return c
}

// WithHttpReadTimeout sets a new Http read timeout
func (c *Config) WithHttpReadTimeout(timeout uint32) *Config {
	c.HttpReadTimeout = timeout
	return c
}

// WithHttpWriteTimeout sets a new Http write timeout
func (c *Config) WithHttpWriteTimeout(timeout uint32) *Config {
	c.HttpWriteTimeout = timeout
	return c
}

// WithHttpReadHeaderTimeout sets a new Http read header timeout
func (c *Config) WithHttpReadHeaderTimeout(timeout uint32) *Config {
	c.HttpReadHeaderTimeout = timeout
	return c
}

func (c *Config) WithAccessLogEventFunc(fn com.LogEventFunc) *Config {
	c.AccessLogEventFunc = fn
	return c
}

func (c *Config) WithRecoveryLogEventFunc(fn com.LogEventFunc) *Config {
	c.RecoveryLogEventFunc = fn
	return c
}

// DefaultConfig returns a default config
func DefaultConfig() *Config {
	return NewConfig()
}

// isConfigValid checks if the configuration is valid
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
