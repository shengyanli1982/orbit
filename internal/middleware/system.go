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
	bp "github.com/shengyanli1982/orbit/internal/pool"
	"github.com/shengyanli1982/orbit/utils/httptool"
	omid "github.com/shengyanli1982/orbit/utils/middleware"
	"go.uber.org/zap"
)

type LogEventFunc func(l *zap.SugaredLogger, e *bp.LogEvent)

func DefaultAcceseEventFunc(l *zap.SugaredLogger, e *bp.LogEvent) {
	l.Infow(
		e.Message,
		"id", e.ID,
		"ip", e.IP,
		"endpoint", e.EndPoint,
		"path", e.Path,
		"method", e.Method,
		"code", e.Code,
		"status", e.Status,
		"latency", e.Latency,
		"agent", e.Agent,
		"query", e.ReqQuery,
		"reqContentType", e.ReqContentType,
		"reqBody", e.ReqBody,
	)
}

func DefaultRecoveryEventFunc(l *zap.SugaredLogger, e *bp.LogEvent) {
	l.Errorw(
		e.Message,
		"id", e.ID,
		"ip", e.IP,
		"endpoint", e.EndPoint,
		"path", e.Path,
		"method", e.Method,
		"code", e.Code,
		"status", e.Status,
		"latency", e.Latency,
		"agent", e.Agent,
		"query", e.ReqQuery,
		"reqContentType", e.ReqContentType,
		"reqBody", e.ReqBody,
		"error", e.Error,
		"errorStack", e.ErrorStack,
	)
}

// 解决跨站
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		//设置缓存时间
		c.Header("Access-Control-Max-Age", "172800")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		// 下一个
		c.Next()
	}
}

func AccessLogger(l *zap.SugaredLogger, fn LogEventFunc, record bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置日志指针
		c.Set(com.RequestLoggerKey, l)

		// 跳过不需要记录的路径
		if omid.SkipResources(c) {
			c.Next()
			// 回收日志读取对象指针
			c.Set(com.RequestLoggerKey, nil)
			return
		}

		// 正常处理系统日志
		start := time.Now()
		path := httptool.GenerateRequestPath(c)
		requestContentType := httptool.StringFilterFlags(c.Request.Header.Get(com.HttpHeaderContentType))

		var reqBody []byte
		if record && httptool.CanRecordContextBody(c.Request.Header) {
			reqBody, _ = httptool.GenerateRequestBody(c)
		}

		// 下一个
		c.Next()

		// response 返回
		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				l.Error(e)
			}
		} else {
			latency := time.Since(start)

			e := com.LogEventPool.Get()
			e.Message = "http server access log"
			e.ID = c.GetHeader(com.HttpRequestID)
			e.IP = c.ClientIP()
			e.EndPoint = c.Request.RemoteAddr
			e.Path = path
			e.Method = c.Request.Method
			e.Code = c.Writer.Status()
			e.Status = http.StatusText(e.Code)
			e.Latency = latency.String()
			e.Agent = c.Request.UserAgent()
			e.ReqContentType = requestContentType
			e.ReqQuery = c.Request.URL.RawQuery
			e.ReqBody = conver.BytesToString(reqBody)

			fn(l, e)
			com.LogEventPool.Put(e)
		}

		// 回收日志读取对象指针
		c.Set(com.RequestLoggerKey, nil)
	}
}

// Recovery recover 掉项目可能出现的panic，并使用zap记录相关日志
func Recovery(l *zap.SugaredLogger, fn LogEventFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic f trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						errMessage := strings.ToLower(se.Error())
						if strings.Contains(errMessage, "broken pipe") || strings.Contains(errMessage, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if brokenPipe {
					l.Error("broken connection", zap.Any("error", err))
				} else {
					var body []byte
					path := httptool.GenerateRequestPath(c)
					body, _ = httptool.GenerateRequestBody(c)
					requestContentType := httptool.StringFilterFlags(c.Request.Header.Get(com.HttpHeaderContentType))

					e := com.LogEventPool.Get()
					e.Message = "http server recovery from panic"
					e.ID = c.GetHeader(com.HttpRequestID)
					e.IP = c.ClientIP()
					e.EndPoint = c.Request.RemoteAddr
					e.Path = path
					e.Method = c.Request.Method
					e.Code = c.Writer.Status()
					e.Status = http.StatusText(e.Code)
					e.Agent = c.Request.UserAgent()
					e.ReqContentType = requestContentType
					e.ReqQuery = c.Request.URL.RawQuery
					e.ReqBody = conver.BytesToString(body)
					e.Error = err
					e.ErrorStack = conver.BytesToString(debug.Stack())

					fn(l, e)
					com.LogEventPool.Put(e)
				}

				// If the connection is dead, we can't write a status to it.
				if brokenPipe {
					_ = c.Error(err.(error)) // nolint: errcheck
					c.Abort()
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
					c.String(http.StatusInternalServerError, "[500] http server internal error, method: "+c.Request.Method+", path: "+c.Request.URL.Path)
				}
			}
		}()

		// 下一个
		c.Next()
	}
}
