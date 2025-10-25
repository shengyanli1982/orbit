package pool

import (
	"bytes"
	"sync"
	"sync/atomic"
)

// GC 影响：
// - sync.Pool 的对象在 GC 时会被清理
// - 大对象会增加 GC 的压力和时间
// - 影响 GC 的效率和应用性能
// 使用场景：
// - 大对象通常是临时的或特殊场景使用
// - 重用的概率较低
// - 维护大对象在池中的成本高于重新创建

const (
	// 缓冲区的默认初始大小 (2KB)
	DefaultInitSize = 2048

	// 缓冲区的默认最大容量 (1MB)
	DefaultMaxCapacity = 1 << 20
)

// 将一个数向上取整到最接近的2的幂
func ceilToPowerOfTwo(n uint32) uint32 {
	if n&(n-1) == 0 {
		return n // 已经是2的幂
	}

	// 将最高位1后面的位全部置1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16

	return n + 1
}

// 用于管理和复用 bytes.Buffer 的对象池
type BufferPool struct {
	pool        sync.Pool // 对象池，用于存储和复用 buffer
	maxCapacity uint32    // buffer 的最大容量限制
	initSize    uint32    // buffer 的初始大小
	gets        uint64    // 获取操作计数
	puts        uint64    // 放回操作计数
}

// 创建一个新的 BufferPool 实例
func NewBufferPool(initSize uint32) *BufferPool {
	// 参数校验：如果初始大小小于等于0，使用默认初始大小
	if initSize <= 0 {
		initSize = DefaultInitSize
	}

	// 计算最大容量：如果初始大小超过默认最大容量，则将最大容量设置为能容纳 initSize*2 的最小2次方数
	maxCapacity := uint32(DefaultMaxCapacity)
	if initSize > DefaultMaxCapacity {
		// 计算 2 倍初始大小的最接近 2 次方数
		maxCapacity = ceilToPowerOfTwo(initSize << 1)
	}

	bp := &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				// 创建一个新的 buffer，初始容量为 initSize
				return bytes.NewBuffer(make([]byte, 0, initSize))
			},
		},
		maxCapacity: maxCapacity,
		initSize:    initSize,
	}

	return bp
}

// 从池中获取一个 bytes.Buffer 对象
func (p *BufferPool) Get() *bytes.Buffer {
	// 增加获取操作计数
	atomic.AddUint64(&p.gets, 1)
	return p.pool.Get().(*bytes.Buffer)
}

// 将一个 bytes.Buffer 对象放回池中
func (p *BufferPool) Put(buf *bytes.Buffer) {
	// 快速返回：如果 buffer 为空，直接返回
	if buf == nil {
		return
	}

	// 增加放回操作计数
	atomic.AddUint64(&p.puts, 1)

	// 容量检查：如果 buffer 容量超过最大限制，直接丢弃
	if int64(buf.Cap()) > int64(p.maxCapacity) {
		return
	}

	// 重置并放回池中
	buf.Reset()
	p.pool.Put(buf)
}

// 返回当前池的最大容量限制
func (p *BufferPool) GetMaxCapacity() uint32 {
	return p.maxCapacity
}

// 返回当前池的初始缓冲区大小
func (p *BufferPool) GetInitSize() uint32 {
	return p.initSize
}

// 返回缓冲池的使用情况
func (p *BufferPool) GetUsage() (gets, puts uint64) {
	return atomic.LoadUint64(&p.gets), atomic.LoadUint64(&p.puts)
}
