package pool

import (
	"bytes"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
