package common

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-logr/logr"
	"github.com/shengyanli1982/orbit/utils/log"
)

// OrbitName 是框架的名称
const OrbitName = "orbit"

// HTTP 头部相关常量
const (
	// HTTP 头部键
	HttpHeaderContentType = "Content-Type"
	HttpHeaderRequestID   = "X-Request-Id"

	// Content-Type 值
	HttpHeaderJSONContentTypeValue       = binding.MIMEJSON
	HttpHeaderXMLContentTypeValue        = binding.MIMEXML
	HttpHeaderPXMLContentTypeValue       = binding.MIMEXML2
	HttpHeaderYAMLContentTypeValue       = binding.MIMEYAML
	HttpHeaderTOMLContentTypeValue       = binding.MIMETOML
	HttpHeaderTextContentTypeValue       = binding.MIMEPlain
	HttpHeaderJavascriptContentTypeValue = "application/javascript"
)

// URL 路径相关常量
const (
	EmptyURLPath       = ""
	PromMetricURLPath  = "/metrics"
	HealthCheckURLPath = "/ping"
	RootURLPath        = "/"
	SwaggerURLPath     = "/docs"
	PprofURLPath       = "/debug/pprof"
)

// 请求相关常量
const (
	// 请求和响应的缓冲区键
	RequestBodyBufferKey  = "REQUEST_BODY_zdiT5HaFaMF7ZfO556rZRYqn"
	ResponseBodyBufferKey = "RESPONSE_BODY_DT6IKLsNULVD3bTgnz1QJbeN"
	RequestLoggerKey      = "REQUEST_LOGGER_3Z3opcTKBSe2O5yZQnSGD"

	// 请求状态码和消息
	RequestOKCode    int64 = 0
	RequestErrorCode int64 = 10
	RequestOK              = "success"
)

// HTTP 服务器默认配置常量
const (
	// 服务器关闭的默认超时时间（秒）
	DefaultShutdownTimeoutSeconds = 10

	// HTTP 请求头的默认最大字节数 (2MB)
	DefaultMaxHeaderBytes int = 1 << 21

	// HTTP 连接的默认空闲超时时间（毫秒）
	DefaultHttpIdleTimeoutMillis uint32 = 15000

	// 默认的 HTTP 监听地址和端口
	DefaultHttpListenAddress        = "127.0.0.1"
	DefaultHttpListenPort    uint16 = 8080
)

// LogEventFunc 是用于记录事件的函数类型
type LogEventFunc func(logger *logr.Logger, event *log.LogEvent)
