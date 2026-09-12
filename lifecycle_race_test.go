package orbit

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pathService 注册唯一路径的 GET 路由，避免并发注册压力测试中出现 gin 路由冲突
type pathService struct {
	path string
}

func (s *pathService) RegisterGroup(g *gin.RouterGroup) {
	g.GET(s.path, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
}

// waitServerReady 轮询健康检查端点直到服务器可以接受请求
func waitServerReady(tb testing.TB, endpoint string) {
	tb.Helper()

	client := &http.Client{Timeout: 200 * time.Millisecond}
	url := "http://" + endpoint + com.HealthCheckURLPath
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	tb.Fatalf("server %s did not become ready in time", endpoint)
}

// TestEngineConcurrentRunStopStress 验证并发 Run/Stop 无数据竞争且无监听器泄漏：
// 两个调用都返回后，引擎必须处于终态且端口不再被残留的监听器占用。
func TestEngineConcurrentRunStopStress(t *testing.T) {
	const iterations = 60
	for i := 0; i < iterations; i++ {
		port := getFreePort(t)
		engine := NewEngine(NewConfig().WithRelease().WithPort(port), NewOptions())
		require.NoError(t, engine.initErr)

		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			engine.Run()
		}()
		go func() {
			defer wg.Done()
			<-start
			engine.Stop()
		}()
		close(start)
		wg.Wait()

		assert.False(t, engine.IsRunning(), "iteration %d: engine still running after Run/Stop", i)

		// Run/Stop 返回后不应再有监听器占用端口（泄漏的 startHTTPServer 会永久持有端口）
		time.Sleep(20 * time.Millisecond)
		probe, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			t.Fatalf("iteration %d: port %d still occupied after Run/Stop returned: %v (listener leak)", i, port, err)
		}
		probe.Close()
	}
}

// TestEngineConcurrentRegisterDuringRun 验证 RegisterService/RegisterMiddleware 与 Run 并发时无数据竞争：
// 注册的 state 检查与 append 必须和 Run 的迭代互斥（check-then-act 原子化）。
func TestEngineConcurrentRegisterDuringRun(t *testing.T) {
	const iterations = 40
	for i := 0; i < iterations; i++ {
		engine := NewEngine(NewConfig().WithRelease().WithPort(getFreePort(t)), NewOptions())
		require.NoError(t, engine.initErr)

		start := make(chan struct{})
		var wg sync.WaitGroup
		for j := 0; j < 4; j++ {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				<-start
				engine.RegisterService(&pathService{path: fmt.Sprintf("/svc-%d", j)})
				engine.RegisterMiddleware(func(c *gin.Context) { c.Next() })
			}(j)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			engine.Run()
		}()
		close(start)
		wg.Wait()

		engine.Stop()
		assert.False(t, engine.IsRunning(), "iteration %d: engine still running after Stop", i)
	}
}

// TestConcurrentStopSecondCallerWaitsForDrain 验证并发 Stop 的第二个调用者在 CAS 失败时
// 等待首个调用者完成在途请求排水后才返回，保证 Stop 返回即"已停止"。
func TestConcurrentStopSecondCallerWaitsForDrain(t *testing.T) {
	port := getFreePort(t)
	engine := NewEngine(NewConfig().WithRelease().WithPort(port), NewOptions())
	require.NoError(t, engine.initErr)

	inHandler := make(chan struct{})
	var handlerDone atomic.Bool
	engine.RegisterService(NewHttpService(func(g *gin.RouterGroup) {
		g.GET("/slow", func(c *gin.Context) {
			close(inHandler)
			time.Sleep(300 * time.Millisecond)
			handlerDone.Store(true)
			c.String(http.StatusOK, "done")
		})
	}))
	engine.Run()
	t.Cleanup(engine.Stop)
	waitServerReady(t, fmt.Sprintf("localhost:%d", port))

	client := &http.Client{Timeout: 10 * time.Second}
	reqDone := make(chan struct{})
	go func() {
		defer close(reqDone)
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/slow", port))
		if err != nil {
			t.Errorf("slow request failed: %v", err)
			return
		}
		defer resp.Body.Close()
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			t.Errorf("slow request body read failed: %v", err)
		}
	}()

	<-inHandler // 确保请求已在处理中

	firstStopReturned := make(chan struct{})
	go func() {
		engine.Stop() // 首个 Stop：进入排水等待
		close(firstStopReturned)
	}()
	time.Sleep(50 * time.Millisecond) // 确保首个 Stop 已开始排水

	// 第二个 Stop：CAS 失败，必须等待排水完成后才返回
	engine.Stop()

	assert.True(t, handlerDone.Load(), "second Stop returned before in-flight request was drained")

	<-firstStopReturned
	<-reqDone
}
