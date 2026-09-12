package middleware

import (
	"bytes"
	"fmt"
	"math"
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
	"github.com/shengyanli1982/orbit/utils/middleware"
)

func formatDurationMs(ns int64) string {
	ms := float64(ns) / 1e6
	var buf [32]byte
	b := strconv.AppendFloat(buf[:0], math.Round(ms*100)/100, 'f', -1, 64)
	b = append(b, "ms"...)
	return string(b)
}

// Pre-computed canonical CORS header keys (already in textproto canonical form)
// to avoid CanonicalMIMEHeaderKey allocations on every request.
const (
	corsHeaderAllowOrigin      = "Access-Control-Allow-Origin"
	corsHeaderAllowMethods     = "Access-Control-Allow-Methods"
	corsHeaderAllowHeaders     = "Access-Control-Allow-Headers"
	corsHeaderExposeHeaders    = "Access-Control-Expose-Headers"
	corsHeaderAllowCredentials = "Access-Control-Allow-Credentials"
	corsHeaderMaxAge           = "Access-Control-Max-Age"
	corsHeaderVary             = "Vary"
)

// Pre-computed static header value slices shared across all requests.
// Safe for concurrent read: the CORS middleware only sets (never adds to) these keys.
var (
	corsAllowOriginAll = []string{"*"}
	corsVaryOrigin     = []string{"Origin"}
	corsBoolTrue       = []string{"true"}
	corsBoolFalse      = []string{"false"}
)

func headerFirstValue(h http.Header, key string) string {
	if v := h[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}

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

	var allowMethodsVal, allowHeadersVal, exposeHeadersVal, maxAgeVal []string
	if allowMethods != "" {
		allowMethodsVal = []string{allowMethods}
	}
	if allowHeaders != "" {
		allowHeadersVal = []string{allowHeaders}
	}
	if exposeHeaders != "" {
		exposeHeadersVal = []string{exposeHeaders}
	}
	if policy.MaxAgeSeconds != 0 {
		maxAgeVal = []string{strconv.Itoa(policy.MaxAgeSeconds)}
	}

	credentialsVal := corsBoolFalse
	if policy.AllowCredentials {
		credentialsVal = corsBoolTrue
	}

	return func(context *gin.Context) {
		if !policy.Enabled {
			context.Next()
			return
		}

		origin := headerFirstValue(context.Request.Header, com.HttpHeaderOrigin)
		// 快路径：无 Origin 的请求不需要 CORS 头。CORS 预检必带 Origin，
		// 无 Origin 的 OPTIONS 属普通流量，一律放行（用户显式注册的 OPTIONS 路由可达），
		// 与 gin-contrib/cors 行为一致
		if origin == "" {
			context.Next()
			return
		}

		// Write headers directly to the map to bypass CanonicalMIMEHeaderKey allocations.
		h := context.Writer.Header()

		// 快路径已保证此处 origin 非空，无需重复判断
		if policy.AllowAllOrigins {
			h[corsHeaderAllowOrigin] = corsAllowOriginAll
		} else if isOriginAllowed(origin, policy.AllowedOrigins) {
			h[corsHeaderAllowOrigin] = []string{origin}
			h[corsHeaderVary] = corsVaryOrigin
		} else {
			// Keep behavior explicit for disallowed origins: no CORS headers returned.
			// 预检返回 403（对齐 gin-contrib/cors），避免服务端观测为"成功 204"
			// 而浏览器侧表现为 CORS 失败的口径错位
			if context.Request.Method == "OPTIONS" {
				context.AbortWithStatus(http.StatusForbidden)
				return
			}
			context.Next()
			return
		}

		if allowMethodsVal != nil {
			h[corsHeaderAllowMethods] = allowMethodsVal
		}
		if allowHeadersVal != nil {
			h[corsHeaderAllowHeaders] = allowHeadersVal
		}
		if exposeHeadersVal != nil {
			h[corsHeaderExposeHeaders] = exposeHeadersVal
		}
		h[corsHeaderAllowCredentials] = credentialsVal
		if maxAgeVal != nil {
			h[corsHeaderMaxAge] = maxAgeVal
		}

		// 处理 OPTIONS 预检请求
		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
			return
		}

		context.Next()
	}
}

// isOriginAllowed 检查给定的 origin 是否在允许的列表中
// 支持通配符 "*" 匹配所有来源
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
		if middleware.SkipResources(context) {
			context.Next()
			return
		}

		// 预先获取所有需要的值，避免重复获取
		req := context.Request
		header := req.Header
		method := req.Method
		path := httptool.GenerateRequestPath(context)
		requestContentType := httptool.StringFilterFlags(headerFirstValue(header, com.HttpHeaderContentType))
		requestID := headerFirstValue(header, com.HttpHeaderRequestID)
		forwardedFor := headerFirstValue(header, com.HttpHeaderForwardedFor)
		userAgent := headerFirstValue(header, "User-Agent")
		remoteAddr := req.RemoteAddr
		// IP 口径与 Recovery 对齐：ClientIP() 经 TrustedProxies 解析转发头，
		// ForwardedByClientIP=false 时退化为 RemoteAddr 的 IP 部分，行为安全
		clientIP := context.ClientIP()
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
		event.Latency = formatDurationMs(time.Since(start).Nanoseconds())
		event.Agent = userAgent
		event.ForwardedFor = forwardedFor
		event.ReqContentType = requestContentType
		event.ReqQuery = rawQuery
		event.ReqBody = conver.BytesToString(requestBody)

		logEventFunc(logger, event)

		if buffer, exists := context.Get(com.RequestBodyBufferKey); exists {
			if buf, ok := buffer.(*bytes.Buffer); ok {
				com.RequestBodyBufferPool.Put(buf)
			}
		}
	}
}

// 返回一个用于处理 panic 恢复的 Gin 中间件。
// record 控制 panic 日志事件是否读取请求体：仅当 record 为 true 且内容类型通过
// httptool.CanRecordContextBody 过滤时才记录（与 AccessLogger 的门控一致），
// 避免默认配置下 panic 日志泄露敏感请求体
func Recovery(logger *logr.Logger, logEventFunc com.LogEventFunc, record bool) gin.HandlerFunc {
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

				// 只在记录开关开启且内容类型通过过滤时才读取请求体
				if record && httptool.CanRecordContextBody(req.Header) {
					if body, err := httptool.GenerateRequestBody(context); err == nil {
						event.ReqBody = conver.BytesToString(body)
					}
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

				// 单次 String 即完成状态码与响应体写入（AbortWithStatus 后再 String 属双写），
				// Abort 保持中止后续处理链的语义不变
				context.String(statusCode, sb.String())
				context.Abort()
			}
		}()

		context.Next()
	}
}
