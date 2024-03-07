package common

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/shengyanli1982/orbit/utils/log"
	"go.uber.org/zap"
)

// OrbitName 是 "orbit" 的常量定义。
// OrbitName is a constant definition of "orbit".

const OrbitName = "orbit"

// HttpHeaderContentType 表示 Content-Type 的 HTTP 头部键。
// HttpHeaderContentType represents the HTTP header key for Content-Type.
const HttpHeaderContentType = "Content-Type"

// HttpHeaderJSONContentTypeValue 表示 JSON Content-Type 的值。
// HttpHeaderJSONContentTypeValue represents the value for JSON Content-Type.
const HttpHeaderJSONContentTypeValue = binding.MIMEJSON

// HttpHeaderXMLContentTypeValue 表示 XML Content-Type 的值。
// HttpHeaderXMLContentTypeValue represents the value for XML Content-Type.
const HttpHeaderXMLContentTypeValue = binding.MIMEXML

// HttpHeaderPXMLContentTypeValue 表示 Test XML Content-Type 的值。
// HttpHeaderPXMLContentTypeValue represents the value for Test XML Content-Type.
const HttpHeaderPXMLContentTypeValue = binding.MIMEXML2

// HttpHeaderYAMLContentTypeValue 表示 YAML Content-Type 的值。
// HttpHeaderYAMLContentTypeValue represents the value for YAML Content-Type.
const HttpHeaderYAMLContentTypeValue = binding.MIMEYAML

// HttpHeaderTOMLContentTypeValue 表示 TOML Content-Type 的值。
// HttpHeaderTOMLContentTypeValue represents the value for TOML Content-Type.
const HttpHeaderTOMLContentTypeValue = binding.MIMETOML

// HttpHeaderTextContentTypeValue 表示 Plain Text Content-Type 的值。
// HttpHeaderTextContentTypeValue represents the value for Plain Text Content-Type.
const HttpHeaderTextContentTypeValue = binding.MIMEPlain

// HttpHeaderJavascriptContentTypeValue 表示 JavaScript Content-Type 的值。
// HttpHeaderJavascriptContentTypeValue represents the value for JavaScript Content-Type.
const HttpHeaderJavascriptContentTypeValue = "application/javascript"

// HttpHeaderRequestID 表示 Request ID 的 HTTP 头部键。
// HttpHeaderRequestID represents the HTTP header key for Request ID.
const HttpHeaderRequestID = "X-Request-Id"

// EmptyURLPath 表示空的 URL 路径。
// EmptyURLPath represents the empty URL path.
const EmptyURLPath = ""

// PromMetricURLPath 表示 Prometheus metrics 的 URL 路径。
// PromMetricURLPath represents the URL path for Prometheus metrics.
const PromMetricURLPath = "/metrics"

// HealthCheckURLPath 表示 health check 的 URL 路径。
// HealthCheckURLPath represents the URL path for health check.
const HealthCheckURLPath = "/ping"

// RootURLPath 表示 root URL 路径。
// RootURLPath represents the root URL path.
const RootURLPath = "/"

// SwaggerURLPath 表示 Swagger documentation 的 URL 路径。
// SwaggerURLPath represents the URL path for Swagger documentation.
const SwaggerURLPath = "/docs"

// PprofURLPath 表示 pprof debugging 的 URL 路径。
// PprofURLPath represents the URL path for pprof debugging.
const PprofURLPath = "/debug/pprof"

// RequestBodyBufferKey 表示 request body buffer 的键。
// RequestBodyBufferKey represents the key for request body buffer.
const RequestBodyBufferKey = "REQUEST_BODY_zdiT5HaFaMF7ZfO556rZRYqn"

// ResponseBodyBufferKey 表示 response body buffer 的键。
// ResponseBodyBufferKey represents the key for response body buffer.
const ResponseBodyBufferKey = "RESPONSE_BODY_DT6IKLsNULVD3bTgnz1QJbeN"

// RequestLoggerKey 表示 request logger 的键。
// RequestLoggerKey represents the key for request logger.
const RequestLoggerKey = "REQUEST_LOGGER_3Z3opcTKBSe2O5yZQnSGD"

// RequestOKCode 表示请求成功的代码。
// RequestOKCode represents the success code for a request.
const RequestOKCode int64 = 0

// RequestErrorCode 表示请求错误的代码。
// RequestErrorCode represents the error code for a request.
const RequestErrorCode int64 = 10

// RequestOK 表示请求成功的消息。
// RequestOK represents the success message for a request.
const RequestOK = "success"

// LogEventFunc 表示一个用于记录事件的函数。
// LogEventFunc represents a function for logging events.
type LogEventFunc func(logger *zap.SugaredLogger, event *log.LogEvent)
