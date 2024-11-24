package orbit

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	ilog "github.com/shengyanli1982/orbit/internal/log"
	mtc "github.com/shengyanli1982/orbit/internal/metric"
	mid "github.com/shengyanli1982/orbit/internal/middleware"
)

// defaultShutdownTimeout 定义了服务优雅关闭的默认超时时间
// defaultShutdownTimeout defines the default timeout duration for graceful shutdown
var defaultShutdownTimeout = 10 * time.Second

// Service 定义了服务的基本接口
// Service defines the basic interface for a service
type Service interface {
	// RegisterGroup 将服务注册到指定的路由组
	// RegisterGroup registers the service to the specified router group
	RegisterGroup(routerGroup *gin.RouterGroup)
}

// Engine 是 Orbit 框架的核心引擎，管理所有服务和中间件
// Engine is the core engine of Orbit framework, managing all services and middlewares
type Engine struct {
	// HTTP 相关配置
	// HTTP related configurations
	endpoint string       // 服务监听地址 / Service listen address
	ginSvr   *gin.Engine  // Gin 引擎实例 / Gin engine instance
	httpSvr  *http.Server // HTTP 服务器实例 / HTTP server instance
	root     *gin.RouterGroup

	// 配置相关
	// Configuration related
	config *Config  // 引擎配置 / Engine configuration
	opts   *Options // 可选配置项 / Optional settings

	// 状态管理
	// State management
	running bool         // 运行状态标志 / Running status flag
	lock    sync.RWMutex // 状态锁 / State lock
	wg      sync.WaitGroup
	once    sync.Once

	// 上下文管理
	// Context management
	ctx    context.Context
	cancel context.CancelFunc

	// 组件管理
	// Component management
	handlers []gin.HandlerFunc // 中间件处理器列表 / Middleware handler list
	services []Service         // 服务列表 / Service list
	metric   *mtc.ServerMetrics
}

// NewEngine 函数用于创建一个新的 Engine 实例。
// The NewEngine function is used to create a new instance of the Engine.
func NewEngine(config *Config, options *Options) *Engine {
	config = isConfigValid(config)
	options = isOptionsValid(options)

	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DisableConsoleColor()

	// 初始化基础引擎
	engine := &Engine{
		endpoint: fmt.Sprintf("%s:%d", config.Address, config.Port),
		config:   config,
		opts:     options,
		handlers: make([]gin.HandlerFunc, 0, 10), // 预分配容量
		services: make([]Service, 0, 10),         // 预分配容量
		metric:   mtc.NewServerMetrics(config.prometheusRegistry),
	}

	// 初始化上下文
	engine.ctx, engine.cancel = context.WithTimeout(context.Background(), defaultShutdownTimeout)

	// 初始化 Gin 服务器
	engine.initGinEngine(options)

	// 注册基础服务
	engine.registerBuiltinServices()

	return engine
}

// 将 Gin 服务器初始化逻辑拆分出来
func (e *Engine) initGinEngine(options *Options) {
	e.ginSvr = gin.New()
	e.root = &e.ginSvr.RouterGroup

	// 设置 Gin 基本配置
	e.ginSvr.ForwardedByClientIP = options.forwordByClientIp
	e.ginSvr.RedirectTrailingSlash = options.trailingSlash
	e.ginSvr.RedirectFixedPath = options.fixedPath
	e.ginSvr.HandleMethodNotAllowed = true

	// 设置基础处理器
	e.setupBaseHandlers()
}

// 将基础处理器设置逻辑拆分出来
func (e *Engine) setupBaseHandlers() {
	e.ginSvr.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "[404] http request route mismatch, method: "+c.Request.Method+", path: "+c.Request.URL.Path)
	})

	e.ginSvr.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "[405] http request method not allowed, method: "+c.Request.Method+", path: "+c.Request.URL.Path)
	})

	// 注册基础中间件
	e.ginSvr.Use(
		mid.Recovery(e.config.logger, e.config.recoveryLogEventFunc),
		mid.BodyBuffer(),
		mid.Cors(),
	)
}

// 将基础服务注册逻辑拆分出来
func (e *Engine) registerBuiltinServices() {
	// 注册健康检查
	healthcheckService(e.root.Group(com.HealthCheckURLPath))

	// 注册可选服务
	if e.opts.swagger {
		swaggerService(e.root.Group(com.SwaggerURLPath))
	}
	if e.opts.pprof {
		pprofService(e.root.Group(com.PprofURLPath))
	}
	if e.opts.metric {
		e.setupMetricService()
	}
}

func (e *Engine) setupMetricService() {
	e.metric.Register()
	e.ginSvr.Use(e.metric.HandlerFunc(e.config.logger))
	metricService(e.root.Group(com.PromMetricURLPath), e.config.prometheusRegistry, e.config.logger)
}

func (e *Engine) Run() {
	if e.IsRunning() {
		return
	}

	e.registerUserMiddlewares()
	e.ginSvr.Use(mid.AccessLogger(e.config.logger, e.config.accessLogEventFunc, e.opts.recReqBody))
	e.registerUserServices()

	e.httpSvr = e.createHTTPServer()
	e.wg.Add(1)

	go e.startHTTPServer()

	e.updateRunningState(true)
	e.once = sync.Once{}
}

// 将 HTTP 服务器创建逻辑拆分出来
func (e *Engine) createHTTPServer() *http.Server {
	return &http.Server{
		Addr:              e.endpoint,
		Handler:           e.ginSvr,
		ReadTimeout:       time.Duration(e.config.HttpReadTimeout) * time.Millisecond,
		ReadHeaderTimeout: time.Duration(e.config.HttpReadHeaderTimeout) * time.Millisecond,
		WriteTimeout:      time.Duration(e.config.HttpWriteTimeout) * time.Millisecond,
		IdleTimeout:       0,
		MaxHeaderBytes:    math.MaxUint32,
		ErrorLog:          ilog.NewStandardLoggerFromLogr(e.config.logger),
	}
}

// 将 HTTP 服务器启动逻辑拆分出来
func (e *Engine) startHTTPServer() {
	defer e.wg.Done()

	e.httpSvr.SetKeepAlivesEnabled(true)
	e.config.logger.Info("http server is ready", "address", e.endpoint)

	if err := e.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		e.config.logger.Error(err, "failed to start http server", "address", e.endpoint)
	}
}

// Stop 方法用于停止 Engine
func (e *Engine) Stop() {
	e.once.Do(func() {
		// 更新运行状态
		e.updateRunningState(false)

		// 关闭 HTTP 服务器
		e.shutdownHTTPServer()

		// 清理资源
		e.cancel()
		e.wg.Wait()

		if e.opts.metric {
			e.metric.Unregister()
		}
	})
}

// shutdownHTTPServer 优雅关闭 HTTP 服务器
func (e *Engine) shutdownHTTPServer() {
	if e.httpSvr == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := e.httpSvr.Shutdown(ctx); err != nil {
		e.config.logger.Error(err, "http server forced to shutdown", "address", e.endpoint)
	}
	e.config.logger.Info("http server is shutdown", "address", e.endpoint)
}

// IsRunning 返回引擎当前的运行状态
// IsRunning returns the current running status of the engine
func (e *Engine) IsRunning() bool {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.running
}

// updateRunningState 方法设置 Orbit 引擎的运行状态。
// The updateRunningState method sets the running status of the Orbit engine.
func (e *Engine) updateRunningState(status bool) {
	// 获取写锁
	// Acquire the write lock
	e.lock.Lock()
	defer e.lock.Unlock()

	// 设置引擎的运行状态
	// Set the running status of the engine
	e.running = status
}

// registerUserServices 方法将所有服务注册到根路由组。
// The registerUserServices method registers all services to the root router group.
func (e *Engine) registerUserServices() {
	// 获取写锁
	// Acquire the write lock
	e.lock.Lock()
	defer e.lock.Unlock()

	// 如果引擎没有在运行
	// If the engine is not running
	if !e.running {
		// 遍历所有服务
		// Iterate over all services
		for i := 0; i < len(e.services); i++ {
			// 将服务注册到根路由组
			// Register the service to the root router group
			e.services[i].RegisterGroup(e.root)
		}
	}
}

// registerUserMiddlewares 方法将所有中间件注册到引擎。
// The registerUserMiddlewares method registers all middlewares to the engine.
func (e *Engine) registerUserMiddlewares() {
	// 获取写锁
	// Acquire the write lock
	e.lock.Lock()
	defer e.lock.Unlock()

	// 如果引擎没有在运行
	// If the engine is not running
	if !e.running {
		// 将所有中间件添加到 Gin 引擎
		// Add all middlewares to the Gin engine
		e.ginSvr.Use(e.handlers...)
	}
}

// RegisterService 注册一个新的服务到引擎
// RegisterService registers a new service to the engine
func (e *Engine) RegisterService(service Service) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.services = append(e.services, service)
	}
}

// RegisterMiddleware 注册一个新的中间件到引擎
// RegisterMiddleware registers a new middleware to the engine
func (e *Engine) RegisterMiddleware(handler gin.HandlerFunc) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.handlers = append(e.handlers, handler)
	}
}

// IsMetricEnabled 返回是否启用了指标收集功能
// IsMetricEnabled returns whether metric collection is enabled
func (e *Engine) IsMetricEnabled() bool {
	return e.opts.metric
}

// IsReleaseMode 返回当前是否处于发布模式
// IsReleaseMode returns whether the engine is in release mode
func (e *Engine) IsReleaseMode() bool {
	return e.config.ReleaseMode
}

// GetLogger 返回引擎的日志记录器
// GetLogger returns the engine's logger
func (e *Engine) GetLogger() *logr.Logger {
	return e.config.logger
}

// GetPrometheusRegistry 返回 Prometheus 注册表实例
// GetPrometheusRegistry returns the Prometheus registry instance
func (e *Engine) GetPrometheusRegistry() *prometheus.Registry {
	return e.config.prometheusRegistry
}

// GetListenEndpoint 返回服务监听的地址和端口
// GetListenEndpoint returns the service's listen address and port
func (e *Engine) GetListenEndpoint() string {
	return e.endpoint
}
