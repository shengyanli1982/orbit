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

// Cors 是一个中间件，用于向响应添加 CORS 头。
// Cors is a middleware that adds CORS headers to the response.
func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 设置跨域请求头
		// Add cross-origin request headers
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		context.Header("Access-Control-Allow-Headers", "*")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		context.Header("Access-Control-Allow-Credentials", "true")
		context.Header("Access-Control-Max-Age", "172800")

		// 处理 OPTIONS 请求
		// Handle OPTIONS request
		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 执行下一个中间件
		// Execute the next middleware
		context.Next()
	}
}

// AccessLogger 是一个中间件，用于记录 HTTP 服务器访问信息。
// AccessLogger is a middleware that logs HTTP server access information.
func AccessLogger(logger *zap.SugaredLogger, logEventFunc com.LogEventFunc, record bool) gin.HandlerFunc {
	return func(context *gin.Context) {
		// 设置请求日志器
		// Set request logger
		context.Set(com.RequestLoggerKey, logger)

		// 设置请求开始时间
		// Start time
		start := time.Now()

		// 生成请求信息
		// Generate request info
		path := httptool.GenerateRequestPath(context)
		requestContentType := httptool.StringFilterFlags(context.Request.Header.Get(com.HttpHeaderContentType))

		// 生成请求体
		// Generate request body
		var requestBody []byte
		if record && httptool.CanRecordContextBody(context.Request.Header) {
			requestBody, _ = httptool.GenerateRequestBody(context)
		}

		// 执行下一个中间件
		// Execute the next middleware
		context.Next()

		// 处理 response
		// Handle response
		if len(context.Errors) > 0 {
			// 记录错误
			// Log error
			for _, err := range context.Errors.Errors() {
				logger.Error(err)
			}
		} else {
			// Response 响应延迟
			// Response latency
			latency := time.Since(start)

			// 记录日志事件
			// Get log event object from pool
			event := com.LogEventPool.Get()

			// 请求的元数据
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

			// 记录事件
			// Log event
			logEventFunc(logger, event)

			// 将事件对象放回池中
			// Put event object back to pool
			com.LogEventPool.Put(event)
		}

		// 清除请求日志器
		// Clear request logger
		context.Set(com.RequestLoggerKey, nil)
	}
}

// Recovery 是一个中间件，用于从 panic 中恢复并记录错误。
// Recovery is a middleware that recovers from panics and logs the error.
func Recovery(logger *zap.SugaredLogger, logEventFunc com.LogEventFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		// 恢复 panic
		// Recover from panic
		defer func() {
			// 获取 panic 错误
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

				// 记录错误
				// Log error
				if brokenPipe {
					logger.Error("broken connection", zap.Any("error", err))
				} else {
					// 生成请求体
					// Generate request body
					var body []byte
					path := httptool.GenerateRequestPath(context)
					body, _ = httptool.GenerateRequestBody(context)
					requestContentType := httptool.StringFilterFlags(context.Request.Header.Get(com.HttpHeaderContentType))

					// 记录日志事件
					// Get log event object from pool
					event := com.LogEventPool.Get()

					// 请求的元数据
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

					// 记录事件
					// Log event
					logEventFunc(logger, event)

					// 将事件对象放回池中
					// Put event object back to pool
					com.LogEventPool.Put(event)
				}

				// 中断请求
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

		// 执行下一个中间件
		// Execute the next middleware
		context.Next()
	}
}
