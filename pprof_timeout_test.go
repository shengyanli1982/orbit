package orbit

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deadlineRecorderWriter 包装 httptest.ResponseRecorder 并记录 ResponseController
// 对连接写截止时间的调整，用于验证 pprof 包装层清除了写截止。
type deadlineRecorderWriter struct {
	*httptest.ResponseRecorder
	deadlines []time.Time
}

func (w *deadlineRecorderWriter) SetWriteDeadline(deadline time.Time) error {
	w.deadlines = append(w.deadlines, deadline)
	return nil
}

// TestPprofRoutesClearConnectionWriteDeadline 验证 pprof 路由包装层清除该连接的写截止：
// 不带 seconds 参数的长响应路由（heap/goroutine 全量 dump 等）不被全局 WriteTimeout 截断。
// 注：Go >= 1.23 标准库 pprof 对 seconds 类路由（profile/trace/delta）已自带
// configureWriteDeadline 延展，包装层的清除对其余 pprof 路由仍然必要。
func TestPprofRoutesClearConnectionWriteDeadline(t *testing.T) {
	engine := NewEngine(NewConfig().WithRelease(), NewOptions().EnablePProf())
	require.NoError(t, engine.initErr)

	rec := &deadlineRecorderWriter{ResponseRecorder: httptest.NewRecorder()}
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/cmdline", nil)
	engine.ginSvr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.deadlines, time.Time{}, "pprof wrapper must clear the connection write deadline")
}

// TestPprofProfileCompletesBeyondWriteTimeout 验证 pprof 长采样路由不被服务器 WriteTimeout 截断：
// WriteTimeout=500ms 下请求 /debug/pprof/profile?seconds=2，响应必须完整写出（gzip 流可解析到结尾）。
// 该用例为端到端回归护栏：在 Go >= 1.23 上标准库 configureWriteDeadline 与 orbit 包装层共同保证此性质。
func TestPprofProfileCompletesBeyondWriteTimeout(t *testing.T) {
	port := getFreePort(t)
	config := NewConfig().WithRelease().WithPort(port).WithHttpWriteTimeout(500)
	engine := NewEngine(config, NewOptions().EnablePProf())
	require.NoError(t, engine.initErr)

	engine.Run()
	t.Cleanup(engine.Stop)
	waitServerReady(t, fmt.Sprintf("localhost:%d", port))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/debug/pprof/profile?seconds=2", port))
	require.NoError(t, err, "profile request failed")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "profile response truncated (write deadline not cleared for pprof route)")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 响应必须是完整的 gzip 压缩 profile 数据
	gz, err := gzip.NewReader(bytes.NewReader(body))
	require.NoError(t, err, "profile body is not valid gzip")
	_, err = io.ReadAll(gz)
	require.NoError(t, err, "profile gzip stream truncated")
	require.NoError(t, gz.Close())
}
