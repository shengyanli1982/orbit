package httptool

import (
	"bytes"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
)

// ResponseBodyWriter 结构体包装了 gin.ResponseWriter，添加了缓冲区功能。
// The ResponseBodyWriter struct wraps gin.ResponseWriter and adds buffer functionality.
type ResponseBodyWriter struct {
	gin.ResponseWriter               // 嵌入 gin 的 ResponseWriter (Embedded gin ResponseWriter)
	buffer             *bytes.Buffer // 用于存储响应数据的缓冲区 (Buffer for storing response data)
}

// NewResponseBodyWriter 函数返回一个新的 ResponseBodyWriter 实例。
// The NewResponseBodyWriter function returns a new ResponseBodyWriter instance.
func NewResponseBodyWriter(w gin.ResponseWriter, buf *bytes.Buffer) *ResponseBodyWriter {
	if buf == nil {
		buf = com.ResponseBodyBufferPool.Get()
	}
	return &ResponseBodyWriter{
		ResponseWriter: w,
		buffer:         buf,
	}
}

// Write 方法实现了 io.Writer 接口，将数据同时写入缓冲区和响应写入器。
// The Write method implements the io.Writer interface, writing data to both buffer and response writer.
func (w *ResponseBodyWriter) Write(b []byte) (int, error) {
	if n, err := w.buffer.Write(b); err != nil {
		return n, err
	}
	return w.ResponseWriter.Write(b)
}

// WriteString 方法提供了字符串写入功能。
// The WriteString method provides string writing functionality.
func (w *ResponseBodyWriter) WriteString(s string) (int, error) {
	if n, err := w.buffer.WriteString(s); err != nil {
		return n, err
	}
	return w.ResponseWriter.WriteString(s)
}

// Reset 方法清空并回收缓冲区。
// The Reset method clears and recycles the buffer.
func (w *ResponseBodyWriter) Reset() {
	w.buffer.Reset()
	com.ResponseBodyBufferPool.Put(w.buffer)
	w.buffer = nil
}

// GetBuffer 方法返回当前缓冲区。
// The GetBuffer method returns the current buffer.
func (w *ResponseBodyWriter) GetBuffer() *bytes.Buffer {
	return w.buffer
}

// GetResponseWriter 方法返回原始的 gin.ResponseWriter。
// The GetResponseWriter method returns the original gin.ResponseWriter.
func (w *ResponseBodyWriter) GetResponseWriter() gin.ResponseWriter {
	return w.ResponseWriter
}

// Size 方法返回已写入的数据大小。
// The Size method returns the size of written data.
func (w *ResponseBodyWriter) Size() int {
	return w.buffer.Len()
}

// Flush 方法将缓冲区数据写入底层的 ResponseWriter。
// The Flush method writes buffered data to the underlying ResponseWriter.
func (w *ResponseBodyWriter) Flush() {
	if w.buffer.Len() > 0 {
		_, _ = w.ResponseWriter.Write(w.buffer.Bytes())
		w.buffer.Reset()
	}
	w.ResponseWriter.Flush()
}
