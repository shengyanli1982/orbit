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
