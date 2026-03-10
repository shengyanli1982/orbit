package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func minInt64(addr *int64, value int64) {
	for {
		old := atomic.LoadInt64(addr)
		if value >= old {
			return
		}
		if atomic.CompareAndSwapInt64(addr, old, value) {
			return
		}
	}
}

func maxInt64(addr *int64, value int64) {
	for {
		old := atomic.LoadInt64(addr)
		if value <= old {
			return
		}
		if atomic.CompareAndSwapInt64(addr, old, value) {
			return
		}
	}
}

func main() {
	url := flag.String("url", "http://127.0.0.1:18080/bench", "target URL")
	method := flag.String("method", http.MethodGet, "HTTP method")
	concurrency := flag.Int("c", runtime.NumCPU()*8, "concurrency")
	duration := flag.Duration("d", 30*time.Second, "test duration")
	timeout := flag.Duration("timeout", 3*time.Second, "request timeout")
	flag.Parse()

	transport := &http.Transport{
		MaxIdleConns:        *concurrency * 2,
		MaxIdleConnsPerHost: *concurrency * 2,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true,
	}
	client := &http.Client{
		Timeout:   *timeout,
		Transport: transport,
	}
	defer transport.CloseIdleConnections()

	start := time.Now()
	end := start.Add(*duration)

	var total uint64
	var success uint64
	var failed uint64
	var totalNs uint64
	var minNs int64 = math.MaxInt64
	var maxNs int64

	var wg sync.WaitGroup
	wg.Add(*concurrency)
	for i := 0; i < *concurrency; i++ {
		go func() {
			defer wg.Done()
			for time.Now().Before(end) {
				reqStart := time.Now()
				req, err := http.NewRequest(*method, *url, nil)
				if err != nil {
					atomic.AddUint64(&failed, 1)
					atomic.AddUint64(&total, 1)
					continue
				}

				resp, err := client.Do(req)
				elapsedNs := time.Since(reqStart).Nanoseconds()
				atomic.AddUint64(&totalNs, uint64(elapsedNs))
				minInt64(&minNs, elapsedNs)
				maxInt64(&maxNs, elapsedNs)

				atomic.AddUint64(&total, 1)
				if err != nil {
					atomic.AddUint64(&failed, 1)
					continue
				}

				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					atomic.AddUint64(&success, 1)
				} else {
					atomic.AddUint64(&failed, 1)
				}
			}
		}()
	}

	wg.Wait()

	elapsed := time.Since(start)
	reqTotal := atomic.LoadUint64(&total)
	reqSuccess := atomic.LoadUint64(&success)
	reqFailed := atomic.LoadUint64(&failed)
	latTotalNs := atomic.LoadUint64(&totalNs)
	minLatencyNs := atomic.LoadInt64(&minNs)
	maxLatencyNs := atomic.LoadInt64(&maxNs)

	var avgLatency time.Duration
	if reqTotal > 0 {
		avgLatency = time.Duration(latTotalNs / reqTotal)
	} else {
		minLatencyNs = 0
		maxLatencyNs = 0
	}

	rps := 0.0
	if elapsed > 0 {
		rps = float64(reqTotal) / elapsed.Seconds()
	}

	fmt.Printf("target=%s method=%s concurrency=%d duration=%s\n", *url, *method, *concurrency, elapsed.Truncate(time.Millisecond))
	fmt.Printf("total=%d success=%d failed=%d\n", reqTotal, reqSuccess, reqFailed)
	fmt.Printf("rps=%.2f avg=%s min=%s max=%s\n",
		rps,
		avgLatency,
		time.Duration(minLatencyNs),
		time.Duration(maxLatencyNs),
	)
}
