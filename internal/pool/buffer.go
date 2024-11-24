package pool

import (
	"bytes"
	"sync"
)

// DefaultBufferSize 是缓冲池的默认大小（2048字节）。
// DefaultBufferSize is the default size of the buffer pool (2048 bytes).
const DefaultBufferSize = 2048

// BufferPool 是一个用于管理和复用 bytes.Buffer 的对象池。
// BufferPool is an object pool for managing and reusing bytes.Buffer objects.
type BufferPool struct {
	pool sync.Pool // 内部使用 sync.Pool 实现对象池 (Internal sync.Pool for object pooling)
}

// NewBufferPool 创建一个新的 BufferPool 实例，可以指定缓冲区的初始大小。
// NewBufferPool creates a new BufferPool instance with specified initial buffer size.
func NewBufferPool(bufferSize uint32) *BufferPool {
	// 如果指定的大小小于等于0，使用默认大小
	// If specified size is less than or equal to 0, use default size
	if bufferSize <= 0 {
		bufferSize = DefaultBufferSize
	}

	return &BufferPool{
		pool: sync.Pool{
			// 定义创建新缓冲区的函数
			// Define function to create new buffer
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, bufferSize))
			},
		},
	}
}

// Get 从池中获取一个 bytes.Buffer 对象。
// Get retrieves a bytes.Buffer object from the pool.
func (p *BufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

// Put 将一个 bytes.Buffer 对象放回池中。
// Put returns a bytes.Buffer object back to the pool.
func (p *BufferPool) Put(buffer *bytes.Buffer) {
	// 确保传入的缓冲区不为空
	// Ensure the input buffer is not nil
	if buffer != nil {
		buffer.Reset()
		p.pool.Put(buffer)
	}
}
