package pool

import (
	"bytes"
	"sync"
)

// DefaultBufferSize 是缓冲区的默认大小。
// DefaultBufferSize is the default size of the buffer.
const DefaultBufferSize = 2048

// BufferPool 表示一个缓冲区池。
// BufferPool represents a pool of buffers.
type BufferPool struct {
	pool sync.Pool // 用于存储缓冲区的同步池 (A sync pool for storing buffers)
}

// NewBufferPool 使用指定的缓冲区大小创建一个新的缓冲区池。
// 如果 bufferSize 小于或等于 0，则使用默认的缓冲区大小。
// NewBufferPool creates a new buffer pool with the specified buffer size.
// If the bufferSize is less than or equal to 0, the default buffer size is used.
func NewBufferPool(bufferSize uint32) *BufferPool {
	if bufferSize <= 0 {
		bufferSize = DefaultBufferSize
	}
	return &BufferPool{
		pool: sync.Pool{
			// 当池中没有可用对象时，New 函数将创建一个新的缓冲区。
			// The New function creates a new buffer when there are no available objects in the pool.
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, bufferSize))
			},
		},
	}
}

// Get 从池中检索一个缓冲区。
// Get retrieves a buffer from the pool.
func (p *BufferPool) Get() *bytes.Buffer {
	// 从 pool 中获取一个缓冲区对象，并将其转换为正确的类型。
	// Get a buffer object from the pool and cast it to the correct type.
	return p.pool.Get().(*bytes.Buffer)
}

// Put 将一个缓冲区返回到池中。
// 如果缓冲区不为空，它会在被放回池中之前被重置。
// Put returns a buffer to the pool.
// If the buffer is not nil, it is reset before being put back into the pool.
func (p *BufferPool) Put(buffer *bytes.Buffer) {
	if buffer != nil {
		// 如果缓冲区对象不为空，则重置对象并将其放回到池中。
		// If the buffer object is not nil, reset the object and put it back into the pool.
		buffer.Reset()
		p.pool.Put(buffer)
	}
}
