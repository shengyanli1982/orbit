package httptool

import (
	"bytes"

	"github.com/gin-gonic/gin"
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
	return &ResponseBodyWriter{
		ResponseWriter: w,   // 初始化响应写入器 (Initialize response writer)
		buffer:         buf, // 初始化缓冲区 (Initialize buffer)
	}
}

// Write 方法实现了 io.Writer 接口，将数据同时写入缓冲区和响应写入器。
// The Write method implements the io.Writer interface, writing data to both buffer and response writer.
func (w *ResponseBodyWriter) Write(b []byte) (int, error) {
	if count, err := w.buffer.Write(b); err != nil {
		return count, err
	}

	return w.ResponseWriter.Write(b)
}

// WriteString 方法提供了字符串写入功能。
// The WriteString method provides string writing functionality.
func (w *ResponseBodyWriter) WriteString(s string) (int, error) {
	if count, err := w.buffer.WriteString(s); err != nil {
		return count, err
	}

	return w.ResponseWriter.WriteString(s)
}

// Reset 方法清空缓冲区。
// The Reset method clears the buffer.
func (w *ResponseBodyWriter) Reset() {
	w.buffer = nil // 将缓冲区设置为 nil (Set buffer to nil)
}

// GetBuffer 方法返回当前缓冲区。
// The GetBuffer method returns the current buffer.
func (w *ResponseBodyWriter) GetBuffer() *bytes.Buffer {
	return w.buffer // 返回缓冲区 (Return buffer)
}

// GetResponseWriter 方法返回原始的 gin.ResponseWriter。
// The GetResponseWriter method returns the original gin.ResponseWriter.
func (w *ResponseBodyWriter) GetResponseWriter() gin.ResponseWriter {
	return w.ResponseWriter // 返回响应写入器 (Return response writer)
}
