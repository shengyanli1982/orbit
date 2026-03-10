package middleware

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/internal/conver"
	"github.com/shengyanli1982/orbit/utils/httptool"
)

// 返回一个处理跨域请求的 Gin 中间件
func Cors() gin.HandlerFunc {
	legacyPolicy := com.CORSPolicy{
		Enabled:          true,
		AllowAllOrigins:  true,
		AllowedMethods:   []string{"POST", "GET", "OPTIONS", "PUT", "DELETE", "UPDATE"},
		AllowedHeaders:   []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Cache-Control", "Content-Language", "Content-Type"},
		AllowCredentials: true,
		MaxAgeSeconds:    172800,
	}
	return CorsWithPolicy(legacyPolicy)
}

// CorsWithPolicy 返回一个按策略处理跨域请求的 Gin 中间件
func CorsWithPolicy(policy com.CORSPolicy) gin.HandlerFunc {
	allowMethods := strings.Join(policy.AllowedMethods, ", ")
	allowHeaders := strings.Join(policy.AllowedHeaders, ", ")
	exposeHeaders := strings.Join(policy.ExposeHeaders, ", ")
	maxAge := strconv.Itoa(policy.MaxAgeSeconds)

	return func(context *gin.Context) {
		if !policy.Enabled {
			context.Next()
			return
		}

		origin := context.GetHeader("Origin")
		// Fast path: non-browser requests without Origin do not need CORS headers.
		if origin == "" && !policy.AllowAllOrigins {
			if context.Request.Method == "OPTIONS" {
				context.AbortWithStatus(http.StatusNoContent)
				return
			}
			context.Next()
			return
		}

		if policy.AllowAllOrigins {
			context.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" && isOriginAllowed(origin, policy.AllowedOrigins) {
			context.Header("Access-Control-Allow-Origin", origin)
			context.Header("Vary", "Origin")
		} else if origin != "" {
			// Keep behavior explicit for disallowed origins: no CORS headers returned.
			if context.Request.Method == "OPTIONS" {
				context.AbortWithStatus(http.StatusNoContent)
				return
			}
			context.Next()
			return
		}

		if allowMethods != "" {
			context.Header("Access-Control-Allow-Methods", allowMethods)
		}
		if allowHeaders != "" {
			context.Header("Access-Control-Allow-Headers", allowHeaders)
		}
		if exposeHeaders != "" {
			context.Header("Access-Control-Expose-Headers", exposeHeaders)
		}
		context.Header("Access-Control-Allow-Credentials", strconv.FormatBool(policy.AllowCredentials))
		if maxAge != "0" {
			context.Header("Access-Control-Max-Age", maxAge)
		}

		// 处理 OPTIONS 预检请求
		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
			return
		}

		context.Next()
	}
}

func isOriginAllowed(origin string, allowed []string) bool {
	for _, item := range allowed {
		if item == "*" || item == origin {
			return true
		}
	}
	return false
}

// 返回一个用于记录访问日志的 Gin 中间件
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
		forwardedFor := header.Get(com.HttpHeaderForwardedFor)
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
		latency := time.Since(start)
		event.Latency = latency.String()
		event.LatencyMs = latency.Milliseconds()
		event.Agent = userAgent
		event.ForwardedFor = forwardedFor
		event.ReqContentType = requestContentType
		event.ReqQuery = rawQuery
		event.ReqBody = conver.BytesToString(requestBody)

		logEventFunc(logger, event)
	}
}

// 返回一个用于处理 panic 恢复的 Gin 中间件
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
				forwardedFor := req.Header.Get(com.HttpHeaderForwardedFor)
				userAgent := req.UserAgent()
				remoteAddr := req.RemoteAddr
				requestContentType := httptool.StringFilterFlags(
					req.Header.Get(com.HttpHeaderContentType),
				)
				rawQuery := req.URL.RawQuery

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

				statusCode := http.StatusInternalServerError

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
				event.Code = statusCode
				event.Status = http.StatusText(statusCode)
				event.Agent = userAgent
				event.ForwardedFor = forwardedFor
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

				context.AbortWithStatus(statusCode)
				context.String(statusCode, sb.String())
			}
		}()

		context.Next()
	}
}
