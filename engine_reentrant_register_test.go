package orbit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// reentrantService 在 RegisterGroup 回调内重入 engine.RegisterMiddleware，
// 复现"用户在注册回调中调用引擎注册方法"的场景：
// Run 若在持 mu 时调用 RegisterGroup，重入方再取 mu（sync.Mutex 不可重入）即死锁
type reentrantService struct {
	engine *Engine
}

func (s *reentrantService) RegisterGroup(g *gin.RouterGroup) {
	g.GET("/reentrant", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	// 重入注册：此时 state 已为 running，预期被 state 检查拒绝（记日志），但不得因锁重入而阻塞
	s.engine.RegisterMiddleware(func(c *gin.Context) { c.Next() })
}

// TestRunRegisterGroupReentrantRegistrationDoesNotDeadlock 验证 Run 装配段
// 不在持 mu 时调用用户代码：RegisterGroup 内重入 RegisterMiddleware 不死锁，
// 重入注册被 state 检查拒绝（不 append），且服务路由正常注册、引擎可正常服务
func TestRunRegisterGroupReentrantRegistrationDoesNotDeadlock(t *testing.T) {
	config := NewConfig().
		WithRelease().
		WithPort(getFreePort(t)).
		WithLogger(newBenchmarkDiscardLogger())
	engine := NewEngine(config, NewOptions())
	require.NoError(t, engine.initErr)
	engine.RegisterService(&reentrantService{engine: engine})

	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		engine.Run()
	}()

	select {
	case <-runDone:
		// Run 正常返回后才注册清理；死锁路径下 Stop 会永久阻塞在 mu 上，
		// 故不能在 select 之前注册 Cleanup
		t.Cleanup(engine.Stop)
	case <-time.After(5 * time.Second):
		t.Fatal("Run deadlocked: RegisterGroup re-entered RegisterMiddleware while Run held mu (non-reentrant mutex)")
	}

	// 锁外迭代快照仍然完成了服务装配：路由已注册且可正常服务
	require.True(t, engine.IsRunning())
	req := httptest.NewRequest(http.MethodGet, "/reentrant", nil)
	resp := httptest.NewRecorder()
	engine.GetGinEngine().ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)

	// 重入的 RegisterMiddleware 被 state 检查拒绝：handlers 未被 append。
	// 读取发生在 Run 返回之后（装配段已结束，无并发写），与 mu 守护语义一致
	require.Empty(t, engine.handlers)
}
