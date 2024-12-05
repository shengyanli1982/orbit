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
		var requestBody []byte
		req := context.Request
		header := req.Header

		// 使用 defer 确保资源清理
		defer context.Set(com.RequestLoggerKey, nil)

		context.Set(com.RequestLoggerKey, logger)
		start := time.Now()

		// 提前获取常用值，避免重复获取
		path := httptool.GenerateRequestPath(context)
		requestContentType := httptool.StringFilterFlags(header.Get(com.HttpHeaderContentType))

		// 只在需要时才记录请求体
		if record && httptool.CanRecordContextBody(header) {
			requestBody, _ = httptool.GenerateRequestBody(context)
		}

		context.Next()

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

		// 使用指针方式设置字段，减少值拷贝
		event.Message = "http server access log"
		event.ID = header.Get(com.HttpHeaderRequestID)
		event.IP = context.ClientIP()
		event.EndPoint = req.RemoteAddr
		event.Path = path
		event.Method = req.Method
		event.Code = context.Writer.Status()
		event.Status = http.StatusText(event.Code)
		event.Latency = time.Since(start).String()
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

				// 使用指针方式设置字段，减少值拷贝
				event.Message = "http server recovery from panic"
				event.ID = context.GetHeader(com.HttpHeaderRequestID)
				event.IP = context.ClientIP()
				event.EndPoint = context.Request.RemoteAddr
				event.Path = httptool.GenerateRequestPath(context)
				event.Method = context.Request.Method
				event.Code = context.Writer.Status()
				event.Status = http.StatusText(event.Code)
				event.Agent = context.Request.UserAgent()
				event.ReqContentType = httptool.StringFilterFlags(
					context.Request.Header.Get(com.HttpHeaderContentType),
				)
				event.ReqQuery = context.Request.URL.RawQuery

				// 只在需要时才生成请求体
				if body, err := httptool.GenerateRequestBody(context); err == nil {
					event.ReqBody = conver.BytesToString(body)
				}

				event.Error = errObj
				event.ErrorStack = conver.BytesToString(debug.Stack())

				logEventFunc(logger, event)

				// 使用预定义的错误信息和 + 运算符优化字符串拼接
				context.AbortWithStatus(http.StatusInternalServerError)
				errorMsg := "[500] http server internal error, method: " +
					context.Request.Method + ", path: " + context.Request.URL.Path
				context.String(http.StatusInternalServerError, errorMsg)
			}
		}()

		context.Next()
	}
}
