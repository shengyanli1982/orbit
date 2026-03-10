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
	defaultMaxHeaderBytes    = com.DefaultMaxHeaderBytes        // 默认最大头部字节数
	// 默认与 Gin 保持一致：信任所有代理，按需通过 WithTrustedProxies 显式收紧
	defaultTrustedProxies = []string{"0.0.0.0/0", "::/0"}
	// 默认按标准代理头顺序解析真实客户端IP
	defaultRemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}
	// 默认 CORS 策略：启用但保守，不放开所有来源
	defaultCORSPolicy = com.CORSPolicy{
		Enabled:          true,
		AllowAllOrigins:  false,
		AllowedOrigins:   []string{},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-Id"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAgeSeconds:    600,
	}
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
	TrustedProxies        []string             `json:"trustedProxies,omitempty" yaml:"trustedProxies,omitempty"`               // 可信代理CIDR列表
	RemoteIPHeaders       []string             `json:"remoteIPHeaders,omitempty" yaml:"remoteIPHeaders,omitempty"`             // 真实客户端IP解析头
	CORSPolicy            *com.CORSPolicy      `json:"corsPolicy,omitempty" yaml:"corsPolicy,omitempty"`                       // CORS 策略（nil 表示使用默认策略）
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
		MaxHeaderBytes:        uint32(defaultMaxHeaderBytes),
		TrustedProxies:        cloneStringSlice(defaultTrustedProxies),
		RemoteIPHeaders:       cloneStringSlice(defaultRemoteIPHeaders),
		CORSPolicy:            cloneCORSPolicyPtr(&defaultCORSPolicy),
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

// 设置信任的代理CIDR列表
func (c *Config) WithTrustedProxies(proxies []string) *Config {
	c.TrustedProxies = cloneStringSlice(proxies)
	return c
}

// 设置用于解析真实客户端IP的HTTP头部列表
func (c *Config) WithRemoteIPHeaders(headers []string) *Config {
	c.RemoteIPHeaders = cloneStringSlice(headers)
	return c
}

// 设置 CORS 策略
func (c *Config) WithCORSPolicy(policy com.CORSPolicy) *Config {
	c.CORSPolicy = cloneCORSPolicyPtr(&policy)
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
	if conf.TrustedProxies == nil {
		conf.TrustedProxies = cloneStringSlice(defaultConf.TrustedProxies)
	} else {
		conf.TrustedProxies = cloneStringSlice(conf.TrustedProxies)
	}
	if conf.RemoteIPHeaders == nil {
		conf.RemoteIPHeaders = cloneStringSlice(defaultConf.RemoteIPHeaders)
	} else {
		conf.RemoteIPHeaders = cloneStringSlice(conf.RemoteIPHeaders)
	}
	conf.CORSPolicy = normalizeCORSPolicy(conf.CORSPolicy, defaultConf.CORSPolicy)

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

func cloneStringSlice(values []string) []string {
	if values == nil {
		return nil
	}

	result := make([]string, len(values))
	copy(result, values)
	return result
}

func isStringSliceEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	for i := 0; i < len(left); i++ {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func cloneCORSPolicy(policy com.CORSPolicy) com.CORSPolicy {
	policy.AllowedOrigins = cloneStringSlice(policy.AllowedOrigins)
	policy.AllowedMethods = cloneStringSlice(policy.AllowedMethods)
	policy.AllowedHeaders = cloneStringSlice(policy.AllowedHeaders)
	policy.ExposeHeaders = cloneStringSlice(policy.ExposeHeaders)
	return policy
}

func cloneCORSPolicyPtr(policy *com.CORSPolicy) *com.CORSPolicy {
	if policy == nil {
		return nil
	}
	cp := cloneCORSPolicy(*policy)
	return &cp
}

func normalizeCORSPolicy(current, fallback *com.CORSPolicy) *com.CORSPolicy {
	if fallback == nil {
		return cloneCORSPolicyPtr(current)
	}
	if current == nil {
		return cloneCORSPolicyPtr(fallback)
	}

	merged := cloneCORSPolicy(*current)
	def := cloneCORSPolicy(*fallback)

	if len(merged.AllowedMethods) == 0 {
		merged.AllowedMethods = def.AllowedMethods
	}
	if len(merged.AllowedHeaders) == 0 {
		merged.AllowedHeaders = def.AllowedHeaders
	}
	if len(merged.ExposeHeaders) == 0 {
		merged.ExposeHeaders = def.ExposeHeaders
	}
	if merged.MaxAgeSeconds <= 0 {
		merged.MaxAgeSeconds = def.MaxAgeSeconds
	}

	return &merged
}
