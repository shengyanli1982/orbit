package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	omid "github.com/shengyanli1982/orbit/utils/middleware"
)

func BodyBuffer() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过不需要记录的路径
		if omid.SkipResources(c) {
			c.Next()
			return
		}

		// 设置会话 Body Buffer
		c.Set(com.RequestBodyBufferKey, com.ReqBodyBuffPool.Get())
		c.Set(com.ResponseBodyBufferKey, com.RespBodyBuffPool.Get())

		// 执行下一个 middleware
		c.Next()

		// 回收 Buffer Pool 对象
		if o, ok := c.Get(com.RequestBodyBufferKey); ok {
			buf := o.(*bytes.Buffer)
			com.ReqBodyBuffPool.Put(buf)
		}
		if o, ok := c.Get(com.ResponseBodyBufferKey); ok {
			buf := o.(*bytes.Buffer)
			com.RespBodyBuffPool.Put(buf)
		}
	}
}
