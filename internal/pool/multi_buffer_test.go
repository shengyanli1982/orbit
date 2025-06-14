package pool

import (
	"bytes"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiSizeBufferPool(t *testing.T) {
	t.Run("Basic Get/Put Operations", func(t *testing.T) {
		pool := NewMultiSizeBufferPool()

		testSizes := []struct {
			name     string
			size     uint32
			expected uint32
		}{
			{"Tiny", 100, SizeClass2KiB},
			{"Small", 3 * KiB, SizeClass8KiB},
			{"Medium", 20 * KiB, SizeClass32KiB},
			{"Large", 100 * KiB, SizeClass128KiB},
			{"Huge", 300 * KiB, SizeClass512KiB},
			{"Max", 600 * KiB, SizeClass1MiB},
		}

		for _, tc := range testSizes {
			t.Run(tc.name, func(t *testing.T) {
				buf := pool.Get(tc.size)

				assert.GreaterOrEqual(t, buf.Cap(), int(tc.size),
					"buffer capacity should be at least requested size")
				assert.Equal(t, int(tc.expected), buf.Cap(),
					"buffer should be allocated from correct size class")

				pool.Put(buf)
			})
		}
	})

	t.Run("Concurrent Operations", func(t *testing.T) {
		pool := NewMultiSizeBufferPool()
		const (
			numGoroutines = 10
			numOperations = 1000
		)

		var wg sync.WaitGroup
		wg.Add(numGoroutines)
		errChan := make(chan error, numGoroutines)

		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				buffers := make([]*bytes.Buffer, 0, numOperations/10)

				for j := 0; j < numOperations; j++ {
					size := uint32(r.Int31n(int32(SizeClass1MiB)))

					if j%10 == 0 {
						buf := pool.Get(size)
						assert.NotNil(t, buf, "got nil buffer from pool")
						buffers = append(buffers, buf)
					} else {
						buf := pool.Get(size)
						assert.NotNil(t, buf, "got nil buffer from pool")
						buf.Write([]byte("test data"))
						pool.Put(buf)
					}
				}

				for _, buf := range buffers {
					pool.Put(buf)
				}
			}()
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			assert.NoError(t, err, "concurrent operation error")
		}
	})

	t.Run("Edge Cases", func(t *testing.T) {
		pool := NewMultiSizeBufferPool()

		// Test zero size request
		buf := pool.Get(0)
		assert.GreaterOrEqual(t, buf.Cap(), int(SizeClass2KiB),
			"zero size request should return smallest size class")
		pool.Put(buf)

		// Test oversized request
		oversizedBuf := pool.Get(2 * MiB)
		assert.GreaterOrEqual(t, oversizedBuf.Cap(), int(SizeClass1MiB),
			"oversized request should return largest size class")
		pool.Put(oversizedBuf)

		// Test nil buffer
		assert.NotPanics(t, func() {
			pool.Put(nil)
		}, "putting nil buffer should not panic")

		// Test putting oversized buffer
		oversizedBuf = bytes.NewBuffer(make([]byte, 0, 2*MiB))
		assert.NotPanics(t, func() {
			pool.Put(oversizedBuf)
		}, "putting oversized buffer should not panic")
	})

	t.Run("Buffer Reuse", func(t *testing.T) {
		pool := NewMultiSizeBufferPool()

		// Get and write to buffer
		buf1 := pool.Get(100)
		buf1.WriteString("test data")
		originalCap := buf1.Cap()

		// Put back and get again
		pool.Put(buf1)
		buf2 := pool.Get(100)

		// Verify buffer state
		assert.Zero(t, buf2.Len(), "reused buffer should be empty")
		assert.Equal(t, originalCap, buf2.Cap(),
			"reused buffer should maintain its capacity")
	})
}

func BenchmarkMultiSizeBufferPool(b *testing.B) {
	pool := NewMultiSizeBufferPool()
	sizes := []uint32{
		100,       // tiny
		4 * KiB,   // small
		30 * KiB,  // medium
		100 * KiB, // large
		300 * KiB, // huge
	}

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			size := sizes[r.Intn(len(sizes))]
			buf := pool.Get(size)
			buf.Write([]byte("test data"))
			pool.Put(buf)
		}
	})
}

func BenchmarkWithoutPool(b *testing.B) {
	sizes := []uint32{
		100,       // tiny
		4 * KiB,   // small
		30 * KiB,  // medium
		100 * KiB, // large
		300 * KiB, // huge
	}

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			size := sizes[r.Intn(len(sizes))]
			buf := bytes.NewBuffer(make([]byte, 0, size))
			buf.Write([]byte("test data"))
			_ = buf
		}
	})
}

func BenchmarkGetSizeClassIndex(b *testing.B) {
	pool := NewMultiSizeBufferPool()
	sizes := []uint32{
		100,       // tiny
		2 * KiB,   // class 0 boundary
		4 * KiB,   // small
		8 * KiB,   // class 1 boundary
		16 * KiB,  // medium
		32 * KiB,  // class 2 boundary
		64 * KiB,  // large
		128 * KiB, // class 3 boundary
		256 * KiB, // huge
		512 * KiB, // class 4 boundary
		1 * MiB,   // max
		2 * MiB,   // oversized
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			size := sizes[r.Intn(len(sizes))]
			_ = pool.getSizeClassIndex(size)
		}
	})
}

func TestMultiSizeBufferPool_EdgeCases(t *testing.T) {
	pool := NewMultiSizeBufferPool()
	require := require.New(t)

	tests := []struct {
		name     string
		size     uint32
		expected uint32
	}{
		{"Zero Size", 0, SizeClass2KiB},                 // 测试 0 大小
		{"1 Byte", 1, SizeClass2KiB},                    // 测试最小值
		{"2047 Bytes", 2047, SizeClass2KiB},             // 测试 2KiB-1
		{"2048 Bytes", 2048, SizeClass2KiB},             // 测试 2KiB
		{"2049 Bytes", 2049, SizeClass8KiB},             // 测试 2KiB+1
		{"Just Below 8KiB", 8191, SizeClass8KiB},        // 测试 8KiB-1
		{"Exact 8KiB", 8192, SizeClass8KiB},             // 测试 8KiB
		{"Just Above 8KiB", 8193, SizeClass32KiB},       // 测试 8KiB+1
		{"Just Below 512KiB", 524287, SizeClass512KiB},  // 测试 512KiB-1
		{"Exact 512KiB", 524288, SizeClass512KiB},       // 测试 512KiB
		{"Just Above 512KiB", 524289, SizeClass1MiB},    // 测试 512KiB+1
		{"Just Below 1MiB", 1048575, SizeClass1MiB},     // 测试 1MiB-1
		{"Exact 1MiB", 1048576, SizeClass1MiB},          // 测试 1MiB
		{"Above 1MiB", 1048577, SizeClass1MiB},          // 测试超过 1MiB
		{"Way Above 1MiB", 2 * 1048576, SizeClass1MiB},  // 测试远超 1MiB
		{"Max uint32/2", ^uint32(0) / 2, SizeClass1MiB}, // 测试 uint32 最大值的一半
		{"Max uint32", ^uint32(0), SizeClass1MiB},       // 测试 uint32 最大值
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 Get
			buf := pool.Get(tt.size)
			require.NotNil(buf, "Buffer should not be nil")
			require.GreaterOrEqual(buf.Cap(), int(tt.expected),
				"Buffer capacity should be greater than or equal to expected size")

			// 测试写入对应大小的数据
			data := make([]byte, tt.size)
			n, err := buf.Write(data)
			require.NoError(err, "Write should not error")
			require.Equal(int(tt.size), n, "Should write all bytes")

			// 测试 Put
			pool.Put(buf)
		})
	}

	// 测试 nil buffer
	t.Run("Nil Buffer", func(t *testing.T) {
		pool.Put(nil) // 不应该 panic
	})

	// 测试并发安全性
	t.Run("Concurrent Access", func(t *testing.T) {
		const goroutines = 100
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(i int) {
				defer wg.Done()
				size := uint32(i % 1048577) // 0 到 1MiB+1 范围内的值
				buf := pool.Get(size)
				require.NotNil(buf)
				pool.Put(buf)
			}(i)
		}

		wg.Wait()
	})

	// 测试重复获取和归还
	t.Run("Repeated Get and Put", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			size := uint32(i % 1048577)
			buf := pool.Get(size)
			require.NotNil(buf)

			// 写入一些数据
			data := []byte("test data")
			_, err := buf.Write(data)
			require.NoError(err)

			// 确保数据正确
			require.Equal("test data", buf.String())

			// 归还并确保被重置
			pool.Put(buf)
		}
	})
}

// 测试在高负载下的行为
func TestMultiSizeBufferPool_HighLoad(t *testing.T) {
	pool := NewMultiSizeBufferPool()
	require := require.New(t)

	// 模拟高负载场景
	t.Run("High Load Test", func(t *testing.T) {
		const (
			goroutines = 50
			iterations = 1000
		)

		var wg sync.WaitGroup
		wg.Add(goroutines)

		for g := 0; g < goroutines; g++ {
			go func() {
				defer wg.Done()

				buffers := make([]*bytes.Buffer, 0, iterations/10)
				for i := 0; i < iterations; i++ {
					// 随机大小请求
					size := uint32(rand.Intn(1048577)) // 0 到 1MiB
					buf := pool.Get(size)
					require.NotNil(buf)

					// 随机决定是立即归还还是稍后归还
					if rand.Float32() < 0.9 { // 90% 立即归还
						pool.Put(buf)
					} else { // 10% 稍后归还
						buffers = append(buffers, buf)
					}
				}

				// 归还所有剩余的 buffer
				for _, buf := range buffers {
					pool.Put(buf)
				}
			}()
		}

		wg.Wait()
	})
}
