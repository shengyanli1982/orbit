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
	omid "github.com/shengyanli1982/orbit/utils/middleware"
	"go.uber.org/zap"
)

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

func AccessLogger(logger *zap.SugaredLogger, logEventFunc com.LogEventFunc, record bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置日志指针
		c.Set(com.RequestLoggerKey, logger)

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

		var requestBody []byte
		if record && httptool.CanRecordContextBody(c.Request.Header) {
			requestBody, _ = httptool.GenerateRequestBody(c)
		}

		// 下一个
		c.Next()

		// response 返回
		if len(c.Errors) > 0 {
			for _, err := range c.Errors.Errors() {
				logger.Error(err)
			}
		} else {
			latency := time.Since(start)

			event := com.LogEventPool.Get()
			event.Message = "http server access log"
			event.ID = c.GetHeader(com.HttpHeaderRequestID)
			event.IP = c.ClientIP()
			event.EndPoint = c.Request.RemoteAddr
			event.Path = path
			event.Method = c.Request.Method
			event.Code = c.Writer.Status()
			event.Status = http.StatusText(event.Code)
			event.Latency = latency.String()
			event.Agent = c.Request.UserAgent()
			event.ReqContentType = requestContentType
			event.ReqQuery = c.Request.URL.RawQuery
			event.ReqBody = conver.BytesToString(requestBody)

			logEventFunc(logger, event)
			com.LogEventPool.Put(event)
		}

		// 回收日志读取对象指针
		c.Set(com.RequestLoggerKey, nil)
	}
}

// Recovery recover 掉项目可能出现的panic，并使用zap记录相关日志
func Recovery(logger *zap.SugaredLogger, logEventFunc com.LogEventFunc) gin.HandlerFunc {
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
					logger.Error("broken connection", zap.Any("error", err))
				} else {
					var body []byte
					path := httptool.GenerateRequestPath(c)
					body, _ = httptool.GenerateRequestBody(c)
					requestContentType := httptool.StringFilterFlags(c.Request.Header.Get(com.HttpHeaderContentType))

					event := com.LogEventPool.Get()
					event.Message = "http server recovery from panic"
					event.ID = c.GetHeader(com.HttpHeaderRequestID)
					event.IP = c.ClientIP()
					event.EndPoint = c.Request.RemoteAddr
					event.Path = path
					event.Method = c.Request.Method
					event.Code = c.Writer.Status()
					event.Status = http.StatusText(event.Code)
					event.Agent = c.Request.UserAgent()
					event.ReqContentType = requestContentType
					event.ReqQuery = c.Request.URL.RawQuery
					event.ReqBody = conver.BytesToString(body)
					event.Error = err
					event.ErrorStack = conver.BytesToString(debug.Stack())

					logEventFunc(logger, event)
					com.LogEventPool.Put(event)
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
