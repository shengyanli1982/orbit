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
		}

		// 调用 context.Next() 转到下一个中间件或路由处理器
		// Call context.Next() to go to the next middleware or route handler
		context.Next()
	}
}

// AccessLogger 是一个中间件，用于记录 HTTP 访问日志。
// AccessLogger is a middleware for recording HTTP access logs.
func AccessLogger(logger *zap.SugaredLogger, logEventFunc com.LogEventFunc, record bool) gin.HandlerFunc {
	return func(context *gin.Context) {

		// 将 logger 设置到请求上下文中
		// Set the logger into the request context
		context.Set(com.RequestLoggerKey, logger)

		// 记录请求开始的时间
		// Record the start time of the request
		start := time.Now()

		// 生成请求路径和请求内容类型
		// Generate the request path and request content type
		path := httptool.GenerateRequestPath(context)
		requestContentType := httptool.StringFilterFlags(context.Request.Header.Get(com.HttpHeaderContentType))

		// 如果需要记录请求体，并且请求头允许记录请求体，则生成请求体
		// If the request body needs to be recorded and the request header allows the request body to be recorded, generate the request body
		var requestBody []byte
		if record && httptool.CanRecordContextBody(context.Request.Header) {
			requestBody, _ = httptool.GenerateRequestBody(context)
		}

		// 处理下一个中间件或路由处理器
		// Handle the next middleware or route handler
		context.Next()

		// 如果请求上下文中有错误
		// If there are errors in the request context
		if len(context.Errors) > 0 {
			// 记录每一个错误
			// Log each error
			for _, err := range context.Errors.Errors() {
				logger.Error(err)
			}
		} else {
			// 如果没有错误，计算请求的延迟
			// If there are no errors, calculate the latency of the request
			latency := time.Since(start)
			// 从对象池中获取日志事件对象
			// Get the log event object from the object pool
			event := com.LogEventPool.Get()

			// 设置日志事件的各种属性
			// Set various properties of the log event
			event.Message = "http server access log"              // 设置日志消息
			event.ID = context.GetHeader(com.HttpHeaderRequestID) // 获取请求 ID
			event.IP = context.ClientIP()                         // 获取客户端 IP
			event.EndPoint = context.Request.RemoteAddr           // 获取请求的远程地址
			event.Path = path                                     // 设置请求路径
			event.Method = context.Request.Method                 // 获取请求方法
			event.Code = context.Writer.Status()                  // 获取响应状态码
			event.Status = http.StatusText(event.Code)            // 获取响应状态文本
			event.Latency = latency.String()                      // 获取请求延迟
			event.Agent = context.Request.UserAgent()             // 获取用户代理
			event.ReqContentType = requestContentType             // 设置请求的内容类型
			event.ReqQuery = context.Request.URL.RawQuery         // 获取请求的查询参数
			event.ReqBody = conver.BytesToString(requestBody)     // 获取请求体

			// 调用日志事件函数
			// Call the log event function
			logEventFunc(logger, event)

			// 将事件对象放回对象池
			// Put the event object back into the object pool
			com.LogEventPool.Put(event)
		}

		// 将请求上下文中的 logger 设置为 nil
		// Set the logger in the request context to nil
		context.Set(com.RequestLoggerKey, nil)
	}
}

// Recovery 是一个中间件，用于捕获和处理 panic。
// Recovery is a middleware for capturing and handling panic.
func Recovery(logger *zap.SugaredLogger, logEventFunc com.LogEventFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		// 使用 defer 和 recover 来捕获 panic
		// Use defer and recover to capture panic
		defer func() {
			// 如果有 panic 发生
			// If a panic occurs
			if err := recover(); err != nil {
				var brokenPipe bool
				// 检查错误是否为网络操作错误
				// Check if the error is a network operation error
				if ne, ok := err.(*net.OpError); ok {
					// 检查错误是否为系统调用错误
					// Check if the error is a system call error
					if se, ok := ne.Err.(*os.SyscallError); ok {
						errMessage := strings.ToLower(se.Error())
						// 检查错误消息是否包含 "broken pipe" 或 "connection reset by peer"
						// Check if the error message contains "broken pipe" or "connection reset by peer"
						if strings.Contains(errMessage, "broken pipe") || strings.Contains(errMessage, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// 如果是 broken pipe 错误
				// If it's a broken pipe error
				if brokenPipe {
					// 记录错误日志
					// Log the error
					logger.Error("broken connection", zap.Any("error", err))
				} else {
					// 如果不是 broken pipe 错误，生成请求相关的日志信息
					// If it's not a broken pipe error, generate log information related to the request
					var body []byte
					path := httptool.GenerateRequestPath(context)
					body, _ = httptool.GenerateRequestBody(context)
					requestContentType := httptool.StringFilterFlags(context.Request.Header.Get(com.HttpHeaderContentType))

					// 从对象池中获取日志事件对象
					// Get the log event object from the object pool
					event := com.LogEventPool.Get()
					event.Message = "http server recovery from panic"      // 设置日志消息
					event.ID = context.GetHeader(com.HttpHeaderRequestID)  // 获取请求 ID
					event.IP = context.ClientIP()                          // 获取客户端 IP
					event.EndPoint = context.Request.RemoteAddr            // 获取请求的远程地址
					event.Path = path                                      // 设置请求路径
					event.Method = context.Request.Method                  // 获取请求方法
					event.Code = context.Writer.Status()                   // 获取响应状态码
					event.Status = http.StatusText(event.Code)             // 获取响应状态文本
					event.Agent = context.Request.UserAgent()              // 获取用户代理
					event.ReqContentType = requestContentType              // 设置请求的内容类型
					event.ReqQuery = context.Request.URL.RawQuery          // 获取请求的查询参数
					event.ReqBody = conver.BytesToString(body)             // 获取请求体
					event.Error = err                                      // 设置错误
					event.ErrorStack = conver.BytesToString(debug.Stack()) // 设置错误堆栈

					// 调用日志事件函数
					// Call the log event function
					logEventFunc(logger, event)

					// 将事件对象放回对象池
					// Put the event object back into the object pool
					com.LogEventPool.Put(event)
				}

				// 如果是 broken pipe 错误，返回错误并终止请求
				// If it's a broken pipe error, return the error and abort the request
				if brokenPipe {
					_ = context.Error(err.(error))
					context.Abort()
				} else {
					// 如果不是 broken pipe 错误，返回 500 错误并终止请求
					// If it's not a broken pipe error, return a 500 error and abort the request
					context.AbortWithStatus(http.StatusInternalServerError)
					context.String(http.StatusInternalServerError, "[500] http server internal error, method: "+context.Request.Method+", path: "+context.Request.URL.Path)
				}
			}
		}()

		// 继续处理下一个中间件或路由处理器
		// Continue to handle the next middleware or route handler
		context.Next()
	}
}
