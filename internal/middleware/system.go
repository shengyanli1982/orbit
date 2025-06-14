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
		// 预先获取所有需要的值，避免重复获取
		req := context.Request
		header := req.Header
		clientIP := context.ClientIP()
		method := req.Method
		path := httptool.GenerateRequestPath(context)
		requestContentType := httptool.StringFilterFlags(header.Get(com.HttpHeaderContentType))
		requestID := header.Get(com.HttpHeaderRequestID)
		userAgent := req.UserAgent()
		remoteAddr := req.RemoteAddr
		rawQuery := req.URL.RawQuery

		// 设置请求日志记录器
		context.Set(com.RequestLoggerKey, logger)
		start := time.Now()

		// 只在需要时才记录请求体
		var requestBody []byte
		if record && httptool.CanRecordContextBody(header) {
			requestBody, _ = httptool.GenerateRequestBody(context)
		}

		context.Next()

		// 使用 defer 确保资源清理
		defer context.Set(com.RequestLoggerKey, nil)

		// 错误处理优化
		if errs := context.Errors; len(errs) > 0 {
			for _, err := range errs {
				logger.Error(err, "Error occurred")
			}
			return
		}

		// 从对象池获取事件对象
		event := com.LogEventPool.Get()
		defer com.LogEventPool.Put(event)

		// 一次性设置所有字段
		event.Message = "http server access log"
		event.ID = requestID
		event.IP = clientIP
		event.EndPoint = remoteAddr
		event.Path = path
		event.Method = method
		event.Code = context.Writer.Status()
		event.Status = http.StatusText(event.Code)
		event.Latency = time.Since(start).String()
		event.Agent = userAgent
		event.ReqContentType = requestContentType
		event.ReqQuery = rawQuery
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
				// 预先获取所有需要的值，避免重复获取
				req := context.Request
				clientIP := context.ClientIP()
				method := req.Method
				path := httptool.GenerateRequestPath(context)
				requestID := context.GetHeader(com.HttpHeaderRequestID)
				userAgent := req.UserAgent()
				remoteAddr := req.RemoteAddr
				requestContentType := httptool.StringFilterFlags(
					req.Header.Get(com.HttpHeaderContentType),
				)
				rawQuery := req.URL.RawQuery
				status := context.Writer.Status()

				var brokenPipe bool

				// 使用类型断言优化
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						errMsg := strings.ToLower(se.Error())
						if strings.Contains(errMsg, "broken pipe") ||
							strings.Contains(errMsg, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// 优化错误处理
				errObj, ok := err.(error)
				if !ok {
					errObj = fmt.Errorf("%v", err)
				}

				if brokenPipe {
					logger.Error(errObj, "broken connection")
					_ = context.Error(errObj)
					context.Abort()
					return
				}

				// 从对象池获取事件对象
				event := com.LogEventPool.Get()
				defer com.LogEventPool.Put(event)

				// 一次性设置所有字段
				event.Message = "http server recovery from panic"
				event.ID = requestID
				event.IP = clientIP
				event.EndPoint = remoteAddr
				event.Path = path
				event.Method = method
				event.Code = status
				event.Status = http.StatusText(status)
				event.Agent = userAgent
				event.ReqContentType = requestContentType
				event.ReqQuery = rawQuery

				// 只在需要时才生成请求体
				if body, err := httptool.GenerateRequestBody(context); err == nil {
					event.ReqBody = conver.BytesToString(body)
				}

				event.Error = errObj
				event.ErrorStack = conver.BytesToString(debug.Stack())

				logEventFunc(logger, event)

				// 使用 strings.Builder 替代 + 运算符进行字符串拼接
				var sb strings.Builder
				sb.WriteString("[500] http server internal error, method: ")
				sb.WriteString(method)
				sb.WriteString(", path: ")
				sb.WriteString(req.URL.Path)

				context.AbortWithStatus(http.StatusInternalServerError)
				context.String(http.StatusInternalServerError, sb.String())
			}
		}()

		context.Next()
	}
}
