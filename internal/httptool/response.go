package httptool

import (
	"bytes"
	"sync"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

// 包装了 gin.ResponseWriter，添加了缓冲区功能
type ResponseBodyWriter struct {
	gin.ResponseWriter               // 嵌入 gin 的 ResponseWriter
	buffer             *bytes.Buffer // 用于存储响应数据的缓冲区
}

var responseBodyWriterPool = sync.Pool{
	New: func() interface{} {
		return &ResponseBodyWriter{}
	},
}

// 返回一个新的 ResponseBodyWriter 实例
func NewResponseBodyWriter(w gin.ResponseWriter, buf *bytes.Buffer) *ResponseBodyWriter {
	if buf == nil {
		buf = com.ResponseBodyBufferPool.Get()
	}

	rw := responseBodyWriterPool.Get().(*ResponseBodyWriter)
	rw.ResponseWriter = w
	rw.buffer = buf
	return rw
}

// 实现了 io.Writer 接口，将数据同时写入缓冲区和响应写入器
func (w *ResponseBodyWriter) Write(b []byte) (int, error) {
	if n, err := w.buffer.Write(b); err != nil {
		return n, err
	}
	return w.ResponseWriter.Write(b)
}

// 提供了字符串写入功能
func (w *ResponseBodyWriter) WriteString(s string) (int, error) {
	if n, err := w.buffer.WriteString(s); err != nil {
		return n, err
	}
	return w.ResponseWriter.WriteString(s)
}

// 清空并回收缓冲区
func (w *ResponseBodyWriter) Reset() {
	if w.buffer != nil {
		w.buffer.Reset()
		com.ResponseBodyBufferPool.Put(w.buffer)
		w.buffer = nil
	}
	w.ResponseWriter = nil
	responseBodyWriterPool.Put(w)
}

// 返回当前缓冲区
func (w *ResponseBodyWriter) GetBuffer() *bytes.Buffer {
	return w.buffer
}

// 返回原始的 gin.ResponseWriter
func (w *ResponseBodyWriter) GetResponseWriter() gin.ResponseWriter {
	return w.ResponseWriter
}

// 返回已写入的数据大小
func (w *ResponseBodyWriter) Size() int {
	if w.buffer == nil {
		return 0
	}
	return w.buffer.Len()
}

// 将缓冲区数据写入底层的 ResponseWriter
func (w *ResponseBodyWriter) Flush() {
	w.ResponseWriter.Flush()
}
