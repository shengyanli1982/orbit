package middleware

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/internal/conver"
	"github.com/shengyanli1982/orbit/utils/httptool"
)

// Cors 函数返回一个处理跨域请求的 Gin 中间件。
// The Cors function returns a Gin middleware that handles CORS requests.
func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 设置允许跨域的各种 Header
		// Set various CORS headers
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		context.Header("Access-Control-Allow-Headers", "*")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		context.Header("Access-Control-Allow-Credentials", "true")
		context.Header("Access-Control-Max-Age", "172800")

		// 处理 OPTIONS 预检请求
		// Handle OPTIONS preflight requests
		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
			return
		}

		context.Next()
	}
}

// AccessLogger 函数返回一个用于记录访问日志的 Gin 中间件。
// The AccessLogger function returns a Gin middleware that logs access information.
func AccessLogger(logger *logr.Logger, logEventFunc com.LogEventFunc, record bool) gin.HandlerFunc {
	return func(context *gin.Context) {
		var (
			requestBody []byte
			req         = context.Request
			header      = req.Header
		)

		// 确保请求结束时清理 logger
		// Ensure logger is cleaned up when request ends
		defer context.Set(com.RequestLoggerKey, nil)

		// 设置请求上下文中的 logger
		// Set logger in request context
		context.Set(com.RequestLoggerKey, logger)
		start := time.Now()

		// 获取请求信息
		// Get request information
		path := httptool.GenerateRequestPath(context)
		requestContentType := httptool.StringFilterFlags(header.Get(com.HttpHeaderContentType))

		// 如果需要记录请求体且内容类型允许
		// Record request body if needed and content type allows
		if record && httptool.CanRecordContextBody(header) {
			requestBody, _ = httptool.GenerateRequestBody(context)
		}

		context.Next()

		// 处理错误情况
		// Handle errors if any
		if errors := context.Errors; len(errors) > 0 {
			for _, err := range errors {
				logger.Error(err, "Error occurred")
			}
			return
		}

		// 从对象池获取日志事件对象
		// Get log event object from pool
		event := com.LogEventPool.Get()
		defer com.LogEventPool.Put(event)

		latency := time.Since(start)

		// 填充日志事件信息
		// Fill log event information
		event.Message = "http server access log"
		event.ID = header.Get(com.HttpHeaderRequestID)
		event.IP = context.ClientIP()
		event.EndPoint = req.RemoteAddr
		event.Path = path
		event.Method = req.Method
		event.Code = context.Writer.Status()
		event.Status = http.StatusText(event.Code)
		event.Latency = latency.String()
		event.Agent = req.UserAgent()
		event.ReqContentType = requestContentType
		event.ReqQuery = req.URL.RawQuery
		event.ReqBody = conver.BytesToString(requestBody)

		logEventFunc(logger, event)
	}
}

// Recovery 函数返回一个用于处理 panic 恢复的 Gin 中间件。
// The Recovery function returns a Gin middleware that recovers from panics.
func Recovery(logger *logr.Logger, logEventFunc com.LogEventFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var (
					brokenPipe bool
					errObj     error
					method     = context.Request.Method
					path       = context.Request.URL.Path
				)

				// 检查是否为断开连接错误
				// Check if it's a broken pipe error
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						errMsg := strings.ToLower(se.Error())
						brokenPipe = strings.Contains(errMsg, "broken pipe") ||
							strings.Contains(errMsg, "connection reset by peer")
					}
				}

				// 转换错误类型
				// Convert error type
				switch e := err.(type) {
				case error:
					errObj = e
				default:
					errObj = fmt.Errorf("%v", e)
				}

				// 处理断开连接的情况
				// Handle broken pipe case
				if brokenPipe {
					logger.Error(errObj, "broken connection")
					_ = context.Error(errObj)
					context.Abort()
					return
				}

				// 收集请求信息
				// Collect request information
				reqPath := httptool.GenerateRequestPath(context)
				body, _ := httptool.GenerateRequestBody(context)

				// 记录错误日志
				// Log error information
				event := com.LogEventPool.Get()
				event.Message = "http server recovery from panic"
				event.ID = context.GetHeader(com.HttpHeaderRequestID)
				event.IP = context.ClientIP()
				event.EndPoint = context.Request.RemoteAddr
				event.Path = reqPath
				event.Method = method
				event.Code = context.Writer.Status()
				event.Status = http.StatusText(event.Code)
				event.Agent = context.Request.UserAgent()
				event.ReqContentType = httptool.StringFilterFlags(
					context.Request.Header.Get(com.HttpHeaderContentType),
				)
				event.ReqQuery = context.Request.URL.RawQuery
				event.ReqBody = conver.BytesToString(body)
				event.Error = errObj
				event.ErrorStack = conver.BytesToString(debug.Stack())

				logEventFunc(logger, event)
				com.LogEventPool.Put(event)

				// 返回 500 错误
				// Return 500 error
				context.AbortWithStatus(http.StatusInternalServerError)

				// 构建错误响应
				// Build error response
				var b strings.Builder
				b.WriteString("[500] http server internal error, method: ")
				b.WriteString(method)
				b.WriteString(", path: ")
				b.WriteString(path)
				context.String(http.StatusInternalServerError, b.String())
			}
		}()

		context.Next()
	}
}
