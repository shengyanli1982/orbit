package common

import (
	"github.com/gin-gonic/gin/binding"
	bp "github.com/shengyanli1982/orbit/internal/pool"
	"go.uber.org/zap"
)

const (
	HttpHeaderContentType                = "Content-Type"
	HttpHeaderJSONContentTypeValue       = binding.MIMEJSON
	HttpHeaderXMLContentTypeValue        = binding.MIMEXML
	HttpHeaderXML2ContentTypeValue       = binding.MIMEXML2
	HttpHeaderYAMLContentTypeValue       = binding.MIMEYAML
	HttpHeaderTOMLContentTypeValue       = binding.MIMETOML
	HttpHeaderTextContentTypeValue       = binding.MIMEPlain
	HttpHeaderJavascriptContentTypeValue = "application/javascript"
)

const (
	HttpHeaderRequestID = "X-Request-Id"
)

const (
	PromMetricURLPath  = "/metrics"
	HealthCheckURLPath = "/ping"
	RootURLPath        = "/"
	SwaggerURLPath     = "/docs"
	PprofURLPath       = "/debug/pprof"
)

const (
	RequestBodyBufferKey  = "REQUEST_BODY_zdiT5HaFaMF7ZfO556rZRYqn"
	ResponseBodyBufferKey = "RESPONSE_BODY_DT6IKLsNULVD3bTgnz1QJbeN"
	RequestLoggerKey      = "REQUEST_LOGGER_3Z3opcTKBSe2O5yZQnSGD"
)

const (
	RequestOKCode    int64 = 0
	RequestErrorCode int64 = iota + 10
)

const (
	RequestOK = "success"
)

type LogEventFunc func(logger *zap.SugaredLogger, event *bp.LogEvent)
