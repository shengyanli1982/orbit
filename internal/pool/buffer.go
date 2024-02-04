package pool

import (
	"bytes"
	"sync"
)

// DefaultBufferSize 是缓冲区的默认大小。
// DefaultBufferSize is the default size of the buffer.
const DefaultBufferSize = 2048

// BufferPool 表示缓冲区池。
// BufferPool represents a pool of buffers.
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool 创建一个具有指定缓冲区大小的新缓冲区池。
// 如果 bufferSize 小于或等于 0，则使用默认缓冲区大小。
// NewBufferPool creates a new buffer pool with the specified buffer size.
// If the bufferSize is less than or equal to 0, the default buffer size is used.
func NewBufferPool(bufferSize uint32) *BufferPool {
	if bufferSize <= 0 {
		bufferSize = DefaultBufferSize
	}
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, bufferSize))
			},
		},
	}
}

// Get 从对象池中获得一个 buffer
// Get retrieves a buffer from the pool.
func (p *BufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

// Put 将 buffer 放回到对象池中。
// 如果 buffer 不为 nil，则在放回对象池之前会被重置。
// Put returns a buffer to the pool.
// If the buffer is not nil, it is reset before being put back into the pool.
func (p *BufferPool) Put(buffer *bytes.Buffer) {
	if buffer != nil {
		buffer.Reset()
		p.pool.Put(buffer)
	}
}
