package orbit

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	com "github.com/shengyanli1982/orbit/common"
	"github.com/shengyanli1982/orbit/internal/conver"
	"github.com/shengyanli1982/orbit/internal/metric"
	wrap "github.com/shengyanli1982/orbit/utils/wrapper"
	swag "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

// 定义所有 pprof 处理器映射，包含基本处理器和命名处理器
// Define all pprof handler mappings, including basic and named pprofHandlers
var pprofHandlers = map[string]http.HandlerFunc{
	"/":             pprof.Index,
	"/cmdline":      pprof.Cmdline,
	"/profile":      pprof.Profile,
	"/symbol":       pprof.Symbol,
	"/trace":        pprof.Trace,
	"/allocs":       pprof.Handler("allocs").ServeHTTP,
	"/block":        pprof.Handler("block").ServeHTTP,
	"/goroutine":    pprof.Handler("goroutine").ServeHTTP,
	"/heap":         pprof.Handler("heap").ServeHTTP,
	"/mutex":        pprof.Handler("mutex").ServeHTTP,
	"/threadcreate": pprof.Handler("threadcreate").ServeHTTP,
}

// pprofService 函数将 pprof 处理器注册到给定的路由组。
// The pprofService function registers the pprof handlers to the given router group.
func pprofService(group *gin.RouterGroup) {
	// 创建一个 pprof 子路由组，避免重复的路径前缀
	// Create a pprof sub-router group to avoid repeated path prefixes
	pprofGroup := group.Group("/debug/pprof")

	// 统一注册所有 GET 处理器
	// Register all GET handlers uniformly
	for path, handler := range pprofHandlers {
		pprofGroup.GET(path, wrap.WrapHandlerFuncToGin(handler))
	}

	// 单独注册 POST 处理器
	// Register POST handler separately
	pprofGroup.POST("/symbol", wrap.WrapHandlerFuncToGin(pprof.Symbol))
}

// metricService 函数将 prometheus 指标处理器注册到给定的路由组。
// The metricService function registers the prometheus metrics handlers to the given router group.
func metricService(group *gin.RouterGroup, registry *prometheus.Registry, logger *logr.Logger) {
	if group == nil || registry == nil {
		return
	}

	// 创建处理器选项，设置错误日志记录器
	// Create handler options with error logger
	opts := promhttp.HandlerOpts{
		ErrorLog:      metric.NewErrorLog(logger),
		ErrorHandling: promhttp.ContinueOnError, // 继续处理后续请求 / Continue processing subsequent requests
	}

	// 创建指标处理器
	// Create metrics handler
	handler := promhttp.InstrumentMetricHandler(
		registry,
		promhttp.HandlerFor(registry, opts),
	)

	// 注册处理器到路由组
	// Register handler to router group
	group.GET(com.EmptyURLPath, wrap.WrapHandlerToGin(handler))
}

// swaggerService 函数将 swagger 处理器注册到给定的路由组。
// The swaggerService function registers the swagger handlers to the given router group.
func swaggerService(group *gin.RouterGroup) {
	if group == nil {
		return
	}

	// 使用缓存包装的 swagger 处理器
	// Use cached swagger handler
	handler := gs.WrapHandler(swag.Handler)

	// 注册处理器到路由组
	// Register handler to router group
	group.GET("/*any", handler)
}

// healthcheckService 函数将健康检查处理器注册到给定的路由组。
// The healthcheckService function registers the healthcheck handlers to the given router group.
func healthcheckService(group *gin.RouterGroup) {
	// 使用预定义的状态码和响应，避免每次请求时创建新的字符串
	// Use predefined status code and response to avoid creating new strings on each request
	group.GET(com.EmptyURLPath, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", conver.StringToBytes(com.RequestOK))
	})
}

// WrapRegisterService 是服务注册函数的包装器。
// WrapRegisterService is a wrapper for the service registration function.
type WrapRegisterService struct {
	registerFunc func(*gin.RouterGroup)
}

// RegisterGroup 函数将服务注册到给定的路由组。
// The RegisterGroup function registers the service to the given router group.
func (w *WrapRegisterService) RegisterGroup(group *gin.RouterGroup) {
	if w != nil && w.registerFunc != nil {
		w.registerFunc(group)
	}
}

// NewHttpService 函数创建一个新的 WrapRegisterService 实例。
// The NewHttpService function creates a new instance of the WrapRegisterService.
func NewHttpService(registerFunc func(*gin.RouterGroup)) *WrapRegisterService {
	if registerFunc == nil {
		return nil
	}
	return &WrapRegisterService{registerFunc: registerFunc}
}
