package common

import (
	"github.com/gin-gonic/gin/binding"
	bp "github.com/shengyanli1982/orbit/internal/pool"
	"go.uber.org/zap"
)

const OrbitName = "orbit"

// HttpHeaderContentType represents the HTTP header key for Content-Type.
const HttpHeaderContentType = "Content-Type"

// HttpHeaderJSONContentTypeValue represents the value for JSON Content-Type.
const HttpHeaderJSONContentTypeValue = binding.MIMEJSON

// HttpHeaderXMLContentTypeValue represents the value for XML Content-Type.
const HttpHeaderXMLContentTypeValue = binding.MIMEXML

// HttpHeaderPXMLContentTypeValue represents the value for Test XML Content-Type.
const HttpHeaderPXMLContentTypeValue = binding.MIMEXML2

// HttpHeaderYAMLContentTypeValue represents the value for YAML Content-Type.
const HttpHeaderYAMLContentTypeValue = binding.MIMEYAML

// HttpHeaderTOMLContentTypeValue represents the value for TOML Content-Type.
const HttpHeaderTOMLContentTypeValue = binding.MIMETOML

// HttpHeaderTextContentTypeValue represents the value for Plain Text Content-Type.
const HttpHeaderTextContentTypeValue = binding.MIMEPlain

// HttpHeaderJavascriptContentTypeValue represents the value for JavaScript Content-Type.
const HttpHeaderJavascriptContentTypeValue = "application/javascript"

// HttpHeaderRequestID represents the HTTP header key for Request ID.
const HttpHeaderRequestID = "X-Request-Id"

// EmptyURLPath represents the empty URL path.
const EmptyURLPath = ""

// PromMetricURLPath represents the URL path for Prometheus metrics.
const PromMetricURLPath = "/metrics"

// HealthCheckURLPath represents the URL path for health check.
const HealthCheckURLPath = "/ping"

// RootURLPath represents the root URL path.
const RootURLPath = "/"

// SwaggerURLPath represents the URL path for Swagger documentation.
const SwaggerURLPath = "/docs"

// PprofURLPath represents the URL path for pprof debugging.
const PprofURLPath = "/debug/pprof"

// RequestBodyBufferKey represents the key for request body buffer.
const RequestBodyBufferKey = "REQUEST_BODY_zdiT5HaFaMF7ZfO556rZRYqn"

// ResponseBodyBufferKey represents the key for response body buffer.
const ResponseBodyBufferKey = "RESPONSE_BODY_DT6IKLsNULVD3bTgnz1QJbeN"

// RequestLoggerKey represents the key for request logger.
const RequestLoggerKey = "REQUEST_LOGGER_3Z3opcTKBSe2O5yZQnSGD"

// RequestOKCode represents the success code for a request.
const RequestOKCode int64 = 0

// RequestErrorCode represents the error code for a request.
const RequestErrorCode int64 = 10

// RequestOK represents the success message for a request.
const RequestOK = "success"

// LogEventFunc represents a function for logging events.
type LogEventFunc func(logger *zap.SugaredLogger, event *bp.LogEvent)
