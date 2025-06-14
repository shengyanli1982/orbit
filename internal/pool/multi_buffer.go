package pool

import (
	"bytes"
	"math/bits"
)

const (
	// 缓冲区大小常量（字节）
	KiB = 1 << 10 // 1 KiB = 1024 bytes
	MiB = 1 << 20 // 1 MiB = 1024 KiB

	// 缓冲区大小类别定义
	SizeClass2KiB   = 2 * KiB   // 2 KiB
	SizeClass8KiB   = 8 * KiB   // 8 KiB
	SizeClass32KiB  = 32 * KiB  // 32 KiB
	SizeClass128KiB = 128 * KiB // 128 KiB
	SizeClass512KiB = 512 * KiB // 512 KiB
	SizeClass1MiB   = MiB       // 1 MiB
)

// 定义缓冲池的大小类别
var bufferSizeClasses = []uint32{
	SizeClass2KiB,   // Class 0: 2 KiB
	SizeClass8KiB,   // Class 1: 8 KiB
	SizeClass32KiB,  // Class 2: 32 KiB
	SizeClass128KiB, // Class 3: 128 KiB
	SizeClass512KiB, // Class 4: 512 KiB
	SizeClass1MiB,   // Class 5: 1 MiB
}

// 实现了一个多规格的缓冲池，管理不同大小类别的缓冲区
type MultiSizeBufferPool struct {
	sizeClassPools []*BufferPool // 不同大小类别缓冲池的数组
}

// 创建一个新的多规格缓冲池
func NewMultiSizeBufferPool() *MultiSizeBufferPool {
	pool := &MultiSizeBufferPool{
		sizeClassPools: make([]*BufferPool, len(bufferSizeClasses)),
	}

	// 初始化每个大小类别的缓冲池
	for classIndex, classSize := range bufferSizeClasses {
		pool.sizeClassPools[classIndex] = NewBufferPool(classSize)
	}

	return pool
}

// 根据给定的缓冲区大小确定合适的大小类别
func (p *MultiSizeBufferPool) getSizeClassIndex(bufferSize uint32) int {
	// 快速路径：小于等于最小类别
	if bufferSize <= SizeClass2KiB {
		return 0
	}

	// 快速路径：大于等于最大类别
	if bufferSize > SizeClass512KiB {
		return 5 // 1MiB class
	}

	// 向上取整到最接近的 2 的幂
	powerOfTwo := ceilToPowerOfTwo(bufferSize)

	// 通过 powerOfTwo 直接映射到索引
	// 2KiB (1<<11): 11 -> 0
	// 8KiB (1<<13): 13 -> 1
	// 32KiB (1<<15): 15 -> 2
	// 128KiB (1<<17): 17 -> 3
	// 512KiB (1<<19): 19 -> 4
	// 1MiB (1<<20): 20 -> 5
	return int((bits.Len32(powerOfTwo) - 11) >> 1)
}

// 获取一个至少具有指定大小的缓冲区
func (p *MultiSizeBufferPool) Get(minSize uint32) *bytes.Buffer {
	return p.sizeClassPools[p.getSizeClassIndex(minSize)].Get()
}

// 将缓冲区返回到相应大小类别的池中
func (p *MultiSizeBufferPool) Put(buf *bytes.Buffer) {
	if buf == nil {
		return
	}

	bufferSize := uint32(buf.Cap())

	// 如果大于最大大小类别，则丢弃
	if bufferSize > bufferSizeClasses[len(bufferSizeClasses)-1] {
		return
	}

	p.sizeClassPools[p.getSizeClassIndex(bufferSize)].Put(buf)
}
