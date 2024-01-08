package common

import "github.com/gin-gonic/gin/binding"

const (
	HttpHeaderContentType                = "Content-Type"
	HttpHeaderJsonContentTypeValue       = binding.MIMEJSON
	HttpHeaderXmlContentTypeValue        = binding.MIMEXML
	HttpHeaderXml2ContentTypeValue       = binding.MIMEXML2
	HttpHeaderYamlContentTypeValue       = binding.MIMEYAML
	HttpHeaderTomlContentTypeValue       = binding.MIMETOML
	HttpHeaderTextContentTypeValue       = binding.MIMEPlain
	HttpHeaderJavascriptContentTypeValue = "application/javascript"
)

const (
	HttpRequestID = "X-Request-Id"
)

const (
	PromMetricUrlPath      = "/metrics"
	HttpHealthCheckUrlPath = "/ping"
	RootUrlPath            = "/"
	HttpSwaggerUrlPath     = "/docs"
	HttpPprofUrlPath       = "/debug/pprof"
)

const (
	RequestBodyBufferKey  = "REQUEST_BODY_zdiT5HaFaMF7ZfO556rZRYqn"
	ResponseBodyBufferKey = "RESPONSE_BODY_DT6IKLsNULVD3bTgnz1QJbeN"
	RequestLoggerKey      = "REQUEST_LOGGER_3Z3opcTKBSe2O5yZQnSGD"
)

const (
	HttpRequestOkCode    int64 = 0
	HttpRequestErrorCode int64 = iota + 10
)

const (
	HttpRequestOk = "success"
)
