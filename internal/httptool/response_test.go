package httptool

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	bp "github.com/shengyanli1982/orbit/internal/pool"
	"github.com/stretchr/testify/assert"
)

type mockResponseWriter struct {
	gin.ResponseWriter
	written []byte
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	m.written = append(m.written, p...)
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	p := []byte(s)
	m.written = append(m.written, p...)
	return len(p), nil
}

func (m *mockResponseWriter) Flush() {}

func (m *mockResponseWriter) Size() int {
	return len(m.written)
}

func TestResponseBodyWriter_Write(t *testing.T) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)

	data := []byte("test data")
	n, err := w.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, len(data), w.Size())
	assert.Equal(t, data, buf.Bytes())
}

func TestResponseBodyWriter_WriteString(t *testing.T) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)

	data := "test data"
	n, err := w.WriteString(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, len(data), w.Size())
	assert.Equal(t, []byte(data), buf.Bytes())
}

func TestResponseBodyWriter_Flush(t *testing.T) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)

	testData := "test data"
	_, err := w.WriteString(testData)
	assert.NoError(t, err)

	assert.Equal(t, testData, buf.String())
	assert.Equal(t, testData, string(mock.written))

	mock.written = mock.written[:0]

	w.Flush()

	// Flush should not write duplicate payload to the underlying writer.
	assert.Empty(t, mock.written)
	// Keep buffer unchanged for later response body inspection.
	assert.Equal(t, testData, buf.String())
}

func BenchmarkResponseBodyWriter_Write(b *testing.B) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)
	data := []byte("test data")

	b.ResetTimer()
	for range b.N {
		_, _ = w.Write(data)
	}
}

func BenchmarkResponseBodyWriter_WriteString(b *testing.B) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)
	data := "test data"

	b.ResetTimer()
	for range b.N {
		_, _ = w.WriteString(data)
	}
}

func TestResponseBodyWriter_WriteStopsBufferingOverThreshold(t *testing.T) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)

	// 共写入 2MB，分 8 次每次 256KB
	chunk := bytes.Repeat([]byte("a"), 256*1024)
	for range 8 {
		n, err := w.Write(chunk)
		assert.NoError(t, err)
		assert.Equal(t, len(chunk), n)
	}

	// 缓冲达到阈值后停止增长，底层 writer 仍收到完整的 2MB 响应
	assert.GreaterOrEqual(t, buf.Len(), bp.DefaultMaxCapacity)
	assert.LessOrEqual(t, buf.Len(), bp.DefaultMaxCapacity+len(chunk))
	assert.Equal(t, 8*len(chunk), len(mock.written))
}

func TestResponseBodyWriter_WriteStringStopsBufferingOverThreshold(t *testing.T) {
	mock := &mockResponseWriter{written: make([]byte, 0)}
	buf := bytes.NewBuffer(nil)
	w := NewResponseBodyWriter(mock, buf)

	// 共写入 2MB，分 8 次每次 256KB
	chunk := strings.Repeat("b", 256*1024)
	for range 8 {
		n, err := w.WriteString(chunk)
		assert.NoError(t, err)
		assert.Equal(t, len(chunk), n)
	}

	// 缓冲达到阈值后停止增长，底层 writer 仍收到完整的 2MB 响应
	assert.GreaterOrEqual(t, buf.Len(), bp.DefaultMaxCapacity)
	assert.LessOrEqual(t, buf.Len(), bp.DefaultMaxCapacity+len(chunk))
	assert.Equal(t, 8*len(chunk), len(mock.written))
}

// deadlineRecorder 基于 httptest.NewRecorder 的垫片，额外实现 SetWriteDeadline。
// ResponseRecorder 自身不支持连接截止（真实场景由底层 *http.response 实现该方法），
// 此垫片用于验证 ResponseController 能穿透 ResponseBodyWriter 包装触达底层 writer。
type deadlineRecorder struct {
	*httptest.ResponseRecorder
	deadlines []time.Time
}

func (w *deadlineRecorder) SetWriteDeadline(deadline time.Time) error {
	w.deadlines = append(w.deadlines, deadline)
	return nil
}

// TestResponseBodyWriter_UnwrapEnablesResponseController 验证 ResponseBodyWriter 通过
// Unwrap() 暴露底层 writer，使 http.NewResponseController 能穿透 BodyBuffer 包装
// 调整连接写截止（修复 DebugOptions=pprof+recRespBody 下 pprof 写截止清除失效）。
// 穿透链：ResponseBodyWriter → gin.responseWriter → 底层 http writer。
func TestResponseBodyWriter_UnwrapEnablesResponseController(t *testing.T) {
	rec := &deadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
	context, _ := gin.CreateTestContext(rec)
	w := NewResponseBodyWriter(context.Writer, nil)

	// Unwrap 返回底层 gin.ResponseWriter（其自身可被 ResponseController 继续解包）
	assert.Equal(t, http.ResponseWriter(context.Writer), w.Unwrap())

	// 零值时间表示移除连接写截止（与 pprof 包装层行为一致），必须无错误穿透到底层
	err := http.NewResponseController(w).SetWriteDeadline(time.Time{})
	assert.NoError(t, err, "ResponseController must penetrate ResponseBodyWriter via Unwrap")
	assert.Equal(t, []time.Time{{}}, rec.deadlines)
}
