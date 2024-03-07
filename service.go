package orbit

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	com "github.com/shengyanli1982/orbit/common"
	wrap "github.com/shengyanli1982/orbit/utils/wrapper"
	swag "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// pprofService 函数将 pprof 处理器注册到给定的路由组。
// The pprofService function registers the pprof handlers to the given router group.
func pprofService(group *gin.RouterGroup) {
	// 获取 pprof 索引页面
	// Get the pprof index page
	group.GET(com.EmptyURLPath, wrap.WrapHandlerFuncToGin(pprof.Index))

	// 获取命令行参数
	// Get the command line arguments
	group.GET("/cmdline", wrap.WrapHandlerFuncToGin(pprof.Cmdline))

	// 获取分析 goroutine 堆栈跟踪
	// Get the profiling goroutine stack traces
	group.GET("/profile", wrap.WrapHandlerFuncToGin(pprof.Profile))

	// 获取符号表
	// Get the symbol table
	group.GET("/symbol", wrap.WrapHandlerFuncToGin(pprof.Symbol))

	// 获取执行跟踪
	// Get the execution trace
	group.GET("/trace", wrap.WrapHandlerFuncToGin(pprof.Trace))

	// 获取堆分配
	// Get the heap allocations
	group.GET("/allocs", wrap.WrapHandlerFuncToGin(pprof.Handler("allocs").ServeHTTP))

	// 获取 goroutine 阻塞配置文件
	// Get the goroutine blocking profile
	group.GET("/block", wrap.WrapHandlerFuncToGin(pprof.Handler("block").ServeHTTP))

	// 获取 goroutine 配置文件
	// Get the goroutine profile
	group.GET("/goroutine", wrap.WrapHandlerFuncToGin(pprof.Handler("goroutine").ServeHTTP))

	// 获取堆配置文件
	// Get the heap profile
	group.GET("/heap", wrap.WrapHandlerFuncToGin(pprof.Handler("heap").ServeHTTP))

	// 获取互斥锁配置文件
	// Get the mutex profile
	group.GET("/mutex", wrap.WrapHandlerFuncToGin(pprof.Handler("mutex").ServeHTTP))

	// 获取线程创建配置文件
	// Get the thread creation profile
	group.GET("/threadcreate", wrap.WrapHandlerFuncToGin(pprof.Handler("threadcreate").ServeHTTP))

	// 获取符号表
	// Get the symbol table
	group.POST("/pprof/symbol", wrap.WrapHandlerFuncToGin(pprof.Symbol))
}

// metricService 函数将 prometheus 指标处理器注册到给定的路由组。
// The metricService function registers the prometheus metrics handlers to the given router group.
func metricService(group *gin.RouterGroup, registry *prometheus.Registry, logger *zap.SugaredLogger) {
	group.GET(com.EmptyURLPath, wrap.WrapHandlerToGin(promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			ErrorLog: zap.NewStdLog(logger.Desugar()),
		}),
	)))
}

// swaggerService 函数将 swagger 处理器注册到给定的路由组。
// The swaggerService function registers the swagger handlers to the given router group.
func swaggerService(group *gin.RouterGroup) {
	group.GET("/*any", gs.WrapHandler(swag.Handler))
}

// healthcheckService 函数将健康检查处理器注册到给定的路由组。
// The healthcheckService function registers the healthcheck handlers to the given router group.
func healthcheckService(group *gin.RouterGroup) {
	group.GET(com.EmptyURLPath, func(c *gin.Context) {
		c.String(http.StatusOK, com.RequestOK)
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
	w.registerFunc(group)
}

// NewHttpService 函数创建一个新的 WrapRegisterService 实例。
// The NewHttpService function creates a new instance of the WrapRegisterService.
func NewHttpService(registerFunc func(*gin.RouterGroup)) *WrapRegisterService {
	return &WrapRegisterService{registerFunc: registerFunc}
}
