package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/utils/httptool"
)

func BodyBuffer() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置日志指针
		buf := httptool.RequestBodyBuffPool.Get()
		c.Set(com.RequestBodyBufferKey, buf)

		// 跳过不需要记录的路径
		if httptool.SkipResources(c) {
			c.Next()
			// 回收日志读取对象指针
			c.Set(com.RequestLoggerKey, nil)
			return
		}

		// 执行下一个 middleware
		c.Next()

		// 回收日志读取对象指针
		if o, ok := c.Get(com.RequestBodyBufferKey); ok {
			buf := o.(*bytes.Buffer)
			httptool.RequestBodyBuffPool.Put(buf)
		}
	}

}
