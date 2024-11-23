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

// Cors 是一个中间件，用于向响应中添加 CORS 头。
// Cors is a middleware that adds CORS headers to the response.
func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 设置响应头 "Access-Control-Allow-Origin" 的值为 "*"，允许所有来源的跨域请求
		// Set the value of the response header "Access-Control-Allow-Origin" to "*", allowing cross-origin requests from all sources
		context.Header("Access-Control-Allow-Origin", "*")

		// 设置响应头 "Access-Control-Allow-Methods" 的值为 "POST, GET, OPTIONS, PUT, DELETE, UPDATE"，允许这些 HTTP 方法的跨域请求
		// Set the value of the response header "Access-Control-Allow-Methods" to "POST, GET, OPTIONS, PUT, DELETE, UPDATE", allowing cross-origin requests with these HTTP methods
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")

		// 设置响应头 "Access-Control-Allow-Headers" 的值为 "*"，允许所有 HTTP 请求头的跨域请求
		// Set the value of the response header "Access-Control-Allow-Headers" to "*", allowing cross-origin requests with any HTTP request headers
		context.Header("Access-Control-Allow-Headers", "*")

		// 设置响应头 "Access-Control-Expose-Headers" 的值，指定哪些响应头可以在响应中暴露给客户端
		// Set the value of the response header "Access-Control-Expose-Headers", specifying which response headers can be exposed to the client in the response
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")

		// 设置响应头 "Access-Control-Allow-Credentials" 的值为 "true"，允许跨域请求携带凭证信息（如 cookies）
		// Set the value of the response header "Access-Control-Allow-Credentials" to "true", allowing cross-origin requests to carry credential information (such as cookies)
		context.Header("Access-Control-Allow-Credentials", "true")

		// 设置响应头 "Access-Control-Max-Age" 的值为 "172800"，指定预检请求的结果可以被缓存多久（单位：秒）
		// Set the value of the response header "Access-Control-Max-Age" to "172800", specifying how long the result of the preflight request can be cached (in seconds)
		context.Header("Access-Control-Max-Age", "172800")

		// 检查请求方法是否为 "OPTIONS"
		// Check if the request method is "OPTIONS"
		if context.Request.Method == "OPTIONS" {
			// 如果请求方法是 "OPTIONS"，则终止请求并返回无内容状态码（204）
			// If the request method is "OPTIONS", abort the request and return a no content status code (204)
			context.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 调用 context.Next() 转到下一个中间件或路由处理器
		// Call context.Next() to go to the next middleware or route handler
		context.Next()
	}
}

// AccessLogger 是一个中间件，用于记录 HTTP 访问日志。
// AccessLogger is a middleware for recording HTTP access logs.
func AccessLogger(logger *logr.Logger, logEventFunc com.LogEventFunc, record bool) gin.HandlerFunc {
	return func(context *gin.Context) {
		// 提前声明变量
		var (
			requestBody []byte
			req        = context.Request
			header     = req.Header
		)

		// 使用 defer 确保 logger 被清理
		defer context.Set(com.RequestLoggerKey, nil)
		
		// 设置 logger 到上下文
		context.Set(com.RequestLoggerKey, logger)
		start := time.Now()

		// 生成请求信息
		path := httptool.GenerateRequestPath(context)
		requestContentType := httptool.StringFilterFlags(header.Get(com.HttpHeaderContentType))

		// 条件判断合并，减少嵌套
		if record && httptool.CanRecordContextBody(header) {
			requestBody, _ = httptool.GenerateRequestBody(context)
		}

		context.Next()

		// 错误处理
		if errors := context.Errors; len(errors) > 0 {
			for _, err := range errors {
				logger.Error(err, "Error occurred")
			}
			return
		}

		// 获取事件对象并使用 defer 确保返回池
		event := com.LogEventPool.Get()
		defer com.LogEventPool.Put(event)

		// 计算延迟
		latency := time.Since(start)

		// 设置事件属性
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

		// 记录日志
		logEventFunc(logger, event)
	}
}

// Recovery 是一个中间件，用于捕获和处理 panic。
// Recovery is a middleware for capturing and handling panic.
func Recovery(logger *logr.Logger, logEventFunc com.LogEventFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 提前声明变量
				var (
					brokenPipe bool
					errObj     error
					method     = context.Request.Method
					path       = context.Request.URL.Path
				)

				// 优化错误类型断言
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						errMsg := strings.ToLower(se.Error())
						brokenPipe = strings.Contains(errMsg, "broken pipe") ||
							strings.Contains(errMsg, "connection reset by peer")
					}
				}

				// 统一错误处理
				switch e := err.(type) {
				case error:
					errObj = e
				default:
					errObj = fmt.Errorf("%v", e)
				}

				if brokenPipe {
					logger.Error(errObj, "broken connection")
					_ = context.Error(errObj)
					context.Abort()
					return
				}

				// 生成请求相关信息
				reqPath := httptool.GenerateRequestPath(context)
				body, _ := httptool.GenerateRequestBody(context)

				// 获取并设置事件信息
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

				// 记录日志
				logEventFunc(logger, event)
				com.LogEventPool.Put(event)

				// 返回 500 错误
				context.AbortWithStatus(http.StatusInternalServerError)

				// 使用 strings.Builder 优化错误消息拼接
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
