package middleware

import (
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/internal/conver"
	"github.com/shengyanli1982/orbit/utils/httptool"
	"go.uber.org/zap"
)

// Cors is a middleware that adds CORS headers to the response.
func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Add cross-origin request headers
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		context.Header("Access-Control-Allow-Headers", "*")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		context.Header("Access-Control-Allow-Credentials", "true")
		context.Header("Access-Control-Max-Age", "172800")

		// Handle OPTIONS request
		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}

		// Handle the request
		context.Next()
	}
}

// AccessLogger is a middleware that logs HTTP server access information.
func AccessLogger(logger *zap.SugaredLogger, logEventFunc com.LogEventFunc, record bool) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Set request logger
		context.Set(com.RequestLoggerKey, logger)

		// Start time
		start := time.Now()

		// Generate request info
		path := httptool.GenerateRequestPath(context)
		requestContentType := httptool.StringFilterFlags(context.Request.Header.Get(com.HttpHeaderContentType))

		// Generate request body
		var requestBody []byte
		if record && httptool.CanRecordContextBody(context.Request.Header) {
			requestBody, _ = httptool.GenerateRequestBody(context)
		}

		// To next middleware
		context.Next()

		// Handle response
		if len(context.Errors) > 0 {
			// Log error
			for _, err := range context.Errors.Errors() {
				logger.Error(err)
			}
		} else {
			// Response latency
			latency := time.Since(start)

			// Get log event object from pool
			event := com.LogEventPool.Get()

			// Request metadata
			event.Message = "http server access log"
			event.ID = context.GetHeader(com.HttpHeaderRequestID)
			event.IP = context.ClientIP()
			event.EndPoint = context.Request.RemoteAddr
			event.Path = path
			event.Method = context.Request.Method
			event.Code = context.Writer.Status()
			event.Status = http.StatusText(event.Code)
			event.Latency = latency.String()
			event.Agent = context.Request.UserAgent()
			event.ReqContentType = requestContentType
			event.ReqQuery = context.Request.URL.RawQuery
			event.ReqBody = conver.BytesToString(requestBody)

			// Log event
			logEventFunc(logger, event)

			// Put event object back to pool
			com.LogEventPool.Put(event)
		}

		// Clear request logger
		context.Set(com.RequestLoggerKey, nil)
	}
}

// Recovery is a middleware that recovers from panics and logs the error.
func Recovery(logger *zap.SugaredLogger, logEventFunc com.LogEventFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Recover from panic
		defer func() {
			// Get panic error
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						errMessage := strings.ToLower(se.Error())
						if strings.Contains(errMessage, "broken pipe") || strings.Contains(errMessage, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// Log error
				if brokenPipe {
					logger.Error("broken connection", zap.Any("error", err))
				} else {
					// Generate request body
					var body []byte
					path := httptool.GenerateRequestPath(context)
					body, _ = httptool.GenerateRequestBody(context)
					requestContentType := httptool.StringFilterFlags(context.Request.Header.Get(com.HttpHeaderContentType))

					// Get log event object from pool
					event := com.LogEventPool.Get()

					// Request metadata
					event.Message = "http server recovery from panic"
					event.ID = context.GetHeader(com.HttpHeaderRequestID)
					event.IP = context.ClientIP()
					event.EndPoint = context.Request.RemoteAddr
					event.Path = path
					event.Method = context.Request.Method
					event.Code = context.Writer.Status()
					event.Status = http.StatusText(event.Code)
					event.Agent = context.Request.UserAgent()
					event.ReqContentType = requestContentType
					event.ReqQuery = context.Request.URL.RawQuery
					event.ReqBody = conver.BytesToString(body)
					event.Error = err
					event.ErrorStack = conver.BytesToString(debug.Stack())

					// Log event
					logEventFunc(logger, event)

					// Put event object back to pool
					com.LogEventPool.Put(event)
				}

				// Abort request
				if brokenPipe {
					_ = context.Error(err.(error))
					context.Abort()
				} else {
					context.AbortWithStatus(http.StatusInternalServerError)
					context.String(http.StatusInternalServerError, "[500] http server internal error, method: "+context.Request.Method+", path: "+context.Request.URL.Path)
				}
			}
		}()

		// To next middleware
		context.Next()
	}
}
