package orbit

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	com "github.com/shengyanli1982/orbit/common"
	mtc "github.com/shengyanli1982/orbit/internal/metric"
	mid "github.com/shengyanli1982/orbit/internal/middleware"

	"go.uber.org/zap"
)

// defaultShutdownTimeout 优雅关闭的默认超时时间。
// defaultShutdownTimeout is the default timeout for graceful shutdown.
var defaultShutdownTimeout = 10 * time.Second

// Service 是标准服务的接口
// Service is the interface that represents a service.
type Service interface {
	// Register 注册 Service 到路由组
	// Register the service to the router group
	RegisterGroup(routerGroup *gin.RouterGroup)
}

// Engine 是代表 Orbit 引擎的主要结构。
// Engine is the main struct that represents the Orbit engine.
type Engine struct {
	running  bool               // Indicates if the engine is running
	endpoint string             // The endpoint of the engine
	ginSvr   *gin.Engine        // The Gin engine
	httpSvr  *http.Server       // The HTTP server
	root     *gin.RouterGroup   // The root router group
	config   *Config            // The engine configuration
	opts     *Options           // The engine options
	lock     sync.RWMutex       // Mutex for concurrent access
	wg       sync.WaitGroup     // WaitGroup for graceful shutdown
	once     sync.Once          // Once for graceful shutdown
	ctx      context.Context    // Context for graceful shutdown
	cancel   context.CancelFunc // Cancel function for graceful shutdown
	handlers []gin.HandlerFunc  // List of middleware handlers
	services []Service          // List of registered services
	metric   *mtc.ServerMetrics // Prometheus metric
}

// NewEngine 创建一个新的 Engine 实例。
// NewEngine creates a new instance of the Engine.
func NewEngine(config *Config, options *Options) *Engine {
	// 验证配置
	// Validate config
	config = isConfigValid(config)

	// 验证选项
	// Validate options
	options = isOptionsValid(options)

	// 检测是否是 running 模式
	// Check running mode
	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// 关闭控制台颜色
	// Disable console color
	gin.DisableConsoleColor()

	// 创建引擎
	// Create engine
	engine := Engine{
		running:  false,
		endpoint: fmt.Sprintf("%s:%d", config.Address, config.Port),
		config:   config,
		opts:     options,
		lock:     sync.RWMutex{},
		wg:       sync.WaitGroup{},
		once:     sync.Once{},
		handlers: make([]gin.HandlerFunc, 0),
		services: make([]Service, 0),
		metric:   mtc.NewServerMetrics(config.prometheusRegistry),
	}
	engine.ctx, engine.cancel = context.WithTimeout(context.Background(), defaultShutdownTimeout)

	// 创建 Gin 引擎
	// Create Gin engine
	engine.ginSvr = gin.New()

	// 设置 root 路由组
	// (*) 所有默认中间件都注册到 root 路由组，不要更改这个
	// Create root router group
	// (*) all default middlewares are registered to the root router group, don't change this
	engine.root = &engine.ginSvr.RouterGroup

	// 设置 Gin 引擎选项
	// Set Gin engine options
	engine.ginSvr.ForwardedByClientIP = options.forwordByClientIp
	engine.ginSvr.RedirectTrailingSlash = options.trailingSlash
	engine.ginSvr.RedirectFixedPath = options.fixedPath

	// 添加自定义 405 输出
	// Add custom 405 output
	engine.ginSvr.HandleMethodNotAllowed = true
	engine.ginSvr.NoRoute(func(context *gin.Context) {
		context.String(http.StatusNotFound, "[404] http request route mismatch, method: "+context.Request.Method+", path: "+context.Request.URL.Path)
	})

	// 添加自定义 404 输出
	// Add custom 404 output
	engine.ginSvr.NoMethod(func(context *gin.Context) {
		context.String(http.StatusMethodNotAllowed, "[405] http request method not allowed, method: "+context.Request.Method+", path: "+context.Request.URL.Path)
	})

	// 添加健康检查组件
	// Add health check
	healthcheckService(engine.root.Group(com.HealthCheckURLPath))

	// 添加 swagger 组件
	// Add swagger
	if engine.opts.swagger {
		swaggerService(engine.root.Group(com.SwaggerURLPath))
	}

	// 添加性能监控接口
	// Add performance monitoring interface
	if engine.opts.pprof {
		pprofService(engine.root.Group(com.PprofURLPath))
	}

	// 添加 Prometheus 指标接口
	// Add Prometheus metric interface
	if engine.opts.metric {
		engine.metric.Register()
		// 添加 Prometheus 中间件，必须在路由注册之前调用
		// HandlerFunc Must be called before router registered
		engine.ginSvr.Use(engine.metric.HandlerFunc(engine.config.logger))
		metricService(engine.root.Group(com.PromMetricURLPath), engine.config.prometheusRegistry, engine.config.logger)
	}

	// 添加必要的自定义中间件
	// Register necessary middlewares
	engine.ginSvr.Use(
		mid.Recovery(engine.config.logger, engine.config.recoveryLogEventFunc), // Recovery from panic
		mid.BodyBuffer(), // Buffer for request/response body
		mid.Cors(),       // Cross-origin resource sharing
	)

	// 返回引擎
	// Return engine
	return &engine
}

// Run 启动 Orbit 引擎。
// Run starts the Orbit engine.
func (e *Engine) Run() {
	// 阻止重复启动
	// Prevent duplicate startup
	if e.IsRunning() {
		return
	}

	// 注册所有自定义的中间件
	// Register all middlewares
	e.registerAllMiddlewares()

	// 注册访问日志中间件
	// Register access logger middleware
	e.ginSvr.Use(mid.AccessLogger(e.config.logger, e.config.accessLogEventFunc, e.opts.recReqBody))

	// 注册所有服务，必须在所有中间件注册之后调用
	// Register all services, it Must be called after all middlewares are registered
	e.registerAllServices()

	// 初始化 http 服务器
	// Initialize http server
	e.httpSvr = &http.Server{
		Addr:              e.endpoint,
		Handler:           e.ginSvr,
		ReadTimeout:       time.Duration(e.config.HttpReadTimeout) * time.Millisecond,
		ReadHeaderTimeout: time.Duration(e.config.HttpReadHeaderTimeout) * time.Millisecond,
		WriteTimeout:      time.Duration(e.config.HttpWriteTimeout) * time.Millisecond,
		IdleTimeout:       0, // Use HttpReadTimeout as the value here
		MaxHeaderBytes:    math.MaxUint32,
		ErrorLog:          zap.NewStdLog(e.config.logger.Desugar()),
	}

	// 启动 http 服务器
	// Start http server
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		// 设置 keepalive，默认开启
		// Enable keepalive by default
		e.httpSvr.SetKeepAlivesEnabled(true)
		e.config.logger.Infow("http server is ready", "address", e.endpoint)
		if err := e.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.config.logger.Fatalw("failed to start http server", "error", err)
		}
	}()

	// 标记为运行中
	// Mark as running
	e.setRuningStatus(true)

	// 重置等待关闭信号
	// Reset shutdown signal
	e.once = sync.Once{}
}

// Stop 停止 Orbit 引擎。
// Stop stops the Orbit engine.
func (e *Engine) Stop() {
	e.once.Do(func() {
		// 标记为已经关闭
		// Mark as not running
		e.setRuningStatus(false)

		// 关闭 http 服务器
		// Close http server
		if e.httpSvr != nil {
			if err := e.httpSvr.Shutdown(e.ctx); err != nil {
				e.config.logger.Fatalw("http server forced to shutdown", "address", e.endpoint, "error", err)
			}
		}
		e.config.logger.Infow("http server is shutdown", "address", e.endpoint)

		// 等待所有协程完成
		// Wait for all goroutines to complete
		e.cancel()
		e.wg.Wait()

		// 注销 Prometheus 指标
		// Unregister Prometheus metric
		if e.opts.metric {
			e.metric.Unregister()
		}
	})
}

// IsRunning 返回 Orbit 引擎是否正在运行。
// IsRunning returns true if the Orbit engine is running.
func (e *Engine) IsRunning() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.running
}

// setRuningStatus 设置 Orbit 引擎的运行状态。
// setRuningStatus sets the running status of the Orbit engine.
func (e *Engine) setRuningStatus(status bool) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	e.running = status
}

// registerAllServices 注册所有服务到根路由组。
// registerAllServices registers all services to the root router group.
func (e *Engine) registerAllServices() {
	e.lock.RLock()
	defer e.lock.RUnlock()
	if !e.running {
		for i := 0; i < len(e.services); i++ {
			e.services[i].RegisterGroup(e.root)
		}
	}
}

// registerAllMiddlewares 注册所有中间件到引擎。
// registerAllMiddlewares registers all middlewares to the engine.
func (e *Engine) registerAllMiddlewares() {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.ginSvr.Use(e.handlers...)
	}
}

// RegisterService 注册服务到 Orbit 引擎。
// RegisterService registers a service to the Orbit engine.
func (e *Engine) RegisterService(service Service) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.services = append(e.services, service)
	}
}

// RegisterMiddleware 注册中间件到 Orbit 引擎。
// RegisterMiddleware registers a middleware to the Orbit engine.
func (e *Engine) RegisterMiddleware(handler gin.HandlerFunc) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.handlers = append(e.handlers, handler)
	}
}

// GetConfig 返回 Orbit 引擎的配置。
// GetConfig returns the metric status of the Orbit engine.
func (e *Engine) IsMetricEnabled() bool {
	return e.opts.metric
}

// IsReleaseMode 返回 Orbit 引擎是否是运行在 release 模式。
// IsReleaseMode returns true if the Orbit engine is running in release mode.
func (e *Engine) IsReleaseMode() bool {
	return e.config.ReleaseMode
}

// GetLogger 返回 Orbit 引擎的日志记录器。
// GetLogger returns the logger of the Orbit engine.
func (e *Engine) GetLogger() *zap.SugaredLogger {
	return e.config.logger
}

// GetPrometheusRegistry 返回 Orbit 引擎的 Prometheus 注册器。
// GetPrometheusRegistry returns the Prometheus registry of the Orbit engine.
func (e *Engine) GetPrometheusRegistry() *prometheus.Registry {
	return e.config.prometheusRegistry
}

// GetListenEndpoint 返回 Orbit 引擎的监听端点。
// GetListenEndpoint returns the listen endpoint of the Orbit engine.
func (e *Engine) GetListenEndpoint() string {
	return e.endpoint
}
