package httptool

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	bp "github.com/shengyanli1982/orbit/internal/pool"
)

// 包装了 gin.ResponseWriter，添加了缓冲区功能
//
// 缓冲存在容量上限（bp.DefaultMaxCapacity，即 1MB）：已缓冲长度达到上限后
// 停止写入缓冲区，后续数据只直通底层 ResponseWriter，超限内容不再被缓冲，
// 以避免大响应场景下每请求重复分配与拷贝 MB 级内存。
type ResponseBodyWriter struct {
	gin.ResponseWriter               // 嵌入 gin 的 ResponseWriter
	buffer             *bytes.Buffer // 用于存储响应数据的缓冲区
	bufferFull         bool          // 缓冲是否已达到容量上限（一次性判定，达到后不再缓冲）
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
	rw.bufferFull = false
	return rw
}

func (w *ResponseBodyWriter) Write(b []byte) (int, error) {
	// 缓冲达到容量上限后只直通底层 writer，不再双写
	if !w.bufferFull {
		if n, err := w.buffer.Write(b); err != nil {
			return n, err
		}
		if w.buffer.Len() >= bp.DefaultMaxCapacity {
			w.bufferFull = true
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *ResponseBodyWriter) WriteString(s string) (int, error) {
	// 缓冲达到容量上限后只直通底层 writer，不再双写
	if !w.bufferFull {
		if n, err := w.buffer.WriteString(s); err != nil {
			return n, err
		}
		if w.buffer.Len() >= bp.DefaultMaxCapacity {
			w.bufferFull = true
		}
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

// Unwrap 返回底层的 http.ResponseWriter，使 http.NewResponseController 能穿透
// BodyBuffer 包装层触达底层连接（如清除/调整 pprof 路由的连接写截止）。
// 返回的 gin.ResponseWriter 具体类型自身也实现了 Unwrap，控制器可继续向下解包
func (w *ResponseBodyWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// 返回已写入的数据大小
func (w *ResponseBodyWriter) Size() int {
	if w.ResponseWriter != nil {
		return w.ResponseWriter.Size()
	}
	if w.buffer == nil {
		return 0
	}
	return w.buffer.Len()
}

// 将缓冲区数据写入底层的 ResponseWriter
func (w *ResponseBodyWriter) Flush() {
	w.ResponseWriter.Flush()
}
