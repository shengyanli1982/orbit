package orbit

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	ulog "github.com/shengyanli1982/orbit/utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccessLoggerRecordsRequestAbortedByUserMiddleware 验证用户中间件 Abort 的请求（如鉴权拒绝 401）
// 仍会产生访问日志事件：AccessLogger 必须注册在用户中间件之前。
func TestAccessLoggerRecordsRequestAbortedByUserMiddleware(t *testing.T) {
	var mu sync.Mutex
	var capturedCodes []int

	config := NewConfig().
		WithRelease().
		WithPort(getFreePort(t)).
		WithAccessLogEventFunc(func(_ *logr.Logger, event *ulog.LogEvent) {
			mu.Lock()
			capturedCodes = append(capturedCodes, event.Code)
			mu.Unlock()
		}).
		WithRecoveryLogEventFunc(benchmarkNoopLogEvent)

	engine := NewEngine(config, NewOptions())
	require.NoError(t, engine.initErr)

	engine.RegisterMiddleware(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	})
	engine.RegisterService(NewHttpService(func(g *gin.RouterGroup) {
		g.GET("/protected", func(c *gin.Context) {
			c.String(http.StatusOK, "secret")
		})
	}))
	engine.Run()
	t.Cleanup(engine.Stop)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, capturedCodes, 1, "aborted request must still produce one access log event")
	assert.Equal(t, http.StatusUnauthorized, capturedCodes[0])
}

// TestAccessLogger_AbortedRequestStillLogged 验证用户中间件 Abort 的请求（如鉴权拒绝 401）
// 仍产生访问日志事件，且 AccessLogger 在 context.Next() 后读取 requestID，
// 可正确捕获用户中间件注入的请求 ID（Abort + RequestID 组合场景，与 P0-1 修复互补）。
func TestAccessLogger_AbortedRequestStillLogged(t *testing.T) {
	var mu sync.Mutex
	var capturedCodes []int
	var capturedIDs []string

	config := NewConfig().
		WithRelease().
		WithPort(getFreePort(t)).
		WithAccessLogEventFunc(func(_ *logr.Logger, event *ulog.LogEvent) {
			mu.Lock()
			capturedCodes = append(capturedCodes, event.Code)
			capturedIDs = append(capturedIDs, event.ID)
			mu.Unlock()
		}).
		WithRecoveryLogEventFunc(benchmarkNoopLogEvent)

	engine := NewEngine(config, NewOptions())
	require.NoError(t, engine.initErr)

	// 用户中间件：先注入请求 ID，再 Abort 401
	engine.RegisterMiddleware(func(c *gin.Context) {
		c.Request.Header.Set("X-Request-Id", "abort-bridge-id")
		c.AbortWithStatus(http.StatusUnauthorized)
	})
	engine.RegisterService(NewHttpService(func(g *gin.RouterGroup) {
		g.GET("/protected", func(c *gin.Context) {
			c.String(http.StatusOK, "secret")
		})
	}))
	engine.Run()
	t.Cleanup(engine.Stop)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp := httptest.NewRecorder()
	engine.ginSvr.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, capturedCodes, 1, "aborted request must still produce one access log event")
	assert.Equal(t, http.StatusUnauthorized, capturedCodes[0])
	assert.Equal(t, "abort-bridge-id", capturedIDs[0],
		"AccessLogger must capture requestID set by user middleware before Abort")
}
