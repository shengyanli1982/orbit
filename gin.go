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

// 默认的服务器关闭超时时间
// Default server shutdown timeout
var defaultShutdownTimeout = 10 * time.Second

// Service 接口定义了注册路由组的方法。
// The Service interface defines the method for registering router groups.
type Service interface {
	RegisterGroup(routerGroup *gin.RouterGroup)
}

// Engine 结构体是 Orbit 框架的核心引擎，包含了 HTTP 服务器和相关配置。
// The Engine struct is the core engine of the Orbit framework, containing HTTP server and related configurations.
type Engine struct {
	endpoint string             // 服务器监听地址和端口 (server listen address and port)
	ginSvr   *gin.Engine        // Gin 引擎实例 (Gin engine instance)
	httpSvr  *http.Server       // HTTP 服务器实例 (HTTP server instance)
	root     *gin.RouterGroup   // 根路由组 (root router group)
	config   *Config            // 服务器配置 (server configuration)
	opts     *Options           // 服务器选项 (server options)
	running  bool               // 服务器运行状态 (server running status)
	lock     sync.RWMutex       // 读写锁，用于并发控制 (read-write lock for concurrency control)
	wg       sync.WaitGroup     // 等待组，用于优雅关闭 (wait group for graceful shutdown)
	once     sync.Once          // 确保某些操作只执行一次 (ensure certain operations execute only once)
	ctx      context.Context    // 上下文，用于控制服务器生命周期 (context for controlling server lifecycle)
	cancel   context.CancelFunc // 取消函数，用于停止服务器 (cancel function for stopping server)
	handlers []gin.HandlerFunc  // 中间件处理函数列表 (list of middleware handlers)
	services []Service          // 服务列表 (list of services)
	metric   *mtc.ServerMetrics // 服务器指标收集器 (server metrics collector)
}

// NewEngine 函数创建并返回一个新的引擎实例。
// The NewEngine function creates and returns a new engine instance.
func NewEngine(config *Config, options *Options) *Engine {
	// 验证配置和选项的有效性
	// Validate configuration and options
	config = isConfigValid(config)
	options = isOptionsValid(options)

	// 如果是发布模式，设置 Gin 为发布模式并禁用控制台颜色
	// If in release mode, set Gin to release mode and disable console color
	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DisableConsoleColor()

	// 创建引擎实例并初始化基本属性
	// Create engine instance and initialize basic properties
	engine := &Engine{
		endpoint: fmt.Sprintf("%s:%d", config.Address, config.Port), // 设置服务器监听地址 (Set server listen address)
		config:   config,                                            // 服务器配置 (Server configuration)
		opts:     options,                                           // 服务器选项 (Server options)
		handlers: make([]gin.HandlerFunc, 0, 10),                    // 初始化中间件处理函数切片 (Initialize middleware handlers slice)
		services: make([]Service, 0, 10),                            // 初始化服务切片 (Initialize services slice)
		metric:   mtc.NewServerMetrics(config.prometheusRegistry),   // 创建指标收集器 (Create metrics collector)
	}

	// 创建带超时的上下文，用于服务器生命周期管理
	// Create context with timeout for server lifecycle management
	engine.ctx, engine.cancel = context.WithTimeout(context.Background(), defaultShutdownTimeout)

	// 初始化 Gin 引擎并设置基本配置
	// Initialize Gin engine and set basic configurations
	engine.initGinEngine(options)

	// 注册内置服务（健康检查、Swagger、Pprof、指标收集等）
	// Register built-in services (health check, Swagger, Pprof, metrics collection, etc.)
	engine.registerBuiltinServices()

	return engine
}

// initGinEngine 方法初始化 Gin 引擎并设置基本配置。
// The initGinEngine method initializes the Gin engine and sets basic configurations.
func (e *Engine) initGinEngine(options *Options) {
	e.ginSvr = gin.New()
	e.root = &e.ginSvr.RouterGroup

	e.ginSvr.ForwardedByClientIP = options.forwordByClientIp
	e.ginSvr.RedirectTrailingSlash = options.trailingSlash
	e.ginSvr.RedirectFixedPath = options.fixedPath
	e.ginSvr.HandleMethodNotAllowed = true

	e.setupBaseHandlers()
}

// setupBaseHandlers 方法设置基本的 HTTP 处理函数，包括 404、405 处理和中间件。
// The setupBaseHandlers method sets up basic HTTP handlers, including 404, 405 handlers and middleware.
func (e *Engine) setupBaseHandlers() {
	// 设置 404 路由未匹配的处理函数
	// Set up handler for 404 route not found
	e.ginSvr.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "[404] http request route mismatch, method: "+c.Request.Method+", path: "+c.Request.URL.Path)
	})

	// 设置 405 方法不允许的处理函数
	// Set up handler for 405 method not allowed
	e.ginSvr.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "[405] http request method not allowed, method: "+c.Request.Method+", path: "+c.Request.URL.Path)
	})

	// 注册基本中间件
	// Register basic middleware
	e.ginSvr.Use(
		mid.Recovery(e.config.logger, e.config.recoveryLogEventFunc), // 恢复中间件 (recovery middleware)
		mid.BodyBuffer(), // 请求体缓冲中间件 (request body buffer middleware)
		mid.Cors(),       // CORS 中间件 (CORS middleware)
	)
}

// registerBuiltinServices 方法注册内置的服务，包括健康检查、Swagger、pprof 和指标收集等。
// The registerBuiltinServices method registers built-in services, including health check, Swagger, pprof, and metrics collection.
func (e *Engine) registerBuiltinServices() {
	// 注册健康检查服务
	// Register health check service
	healthcheckService(e.root.Group(com.HealthCheckURLPath))

	// 根据配置注册可选服务
	// Register optional services based on configuration
	if e.opts.swagger {
		swaggerService(e.root.Group(com.SwaggerURLPath)) // 注册 Swagger 服务 (Register Swagger service)
	}
	if e.opts.pprof {
		pprofService(e.root.Group(com.PprofURLPath)) // 注册 pprof 服务 (Register pprof service)
	}
	if e.opts.metric {
		e.setupMetricService() // 注册指标收集服务 (Register metrics collection service)
	}
}

// setupMetricService 方法设置并注册 Prometheus 指标收集服务。
// The setupMetricService method sets up and registers the Prometheus metrics collection service.
func (e *Engine) setupMetricService() {
	e.metric.Register()                                                                              // 注册指标收集器 (Register metrics collector)
	e.ginSvr.Use(e.metric.HandlerFunc(e.config.logger))                                              // 添加指标收集中间件 (Add metrics collection middleware)
	metricService(e.root.Group(com.PromMetricURLPath), e.config.prometheusRegistry, e.config.logger) // 注册指标服务路由 (Register metrics service route)
}

// Run 方法启动 HTTP 服务器。
// The Run method starts the HTTP server.
func (e *Engine) Run() {
	// 检查服务器是否已经在运行
	// Check if server is already running
	if e.IsRunning() {
		return
	}

	// 注册用户中间件和服务
	// Register user middleware and services
	e.registerUserMiddlewares()
	e.ginSvr.Use(mid.AccessLogger(e.config.logger, e.config.accessLogEventFunc, e.opts.recReqBody))
	e.registerUserServices()

	// 创建并启动 HTTP 服务器
	// Create and start HTTP server
	e.httpSvr = e.createHTTPServer()
	e.wg.Add(1)

	go e.startHTTPServer()

	// 更新服务器状态
	// Update server status
	e.updateRunningState(true)
	e.once = sync.Once{}
}

// createHTTPServer 方法创建并配置 HTTP 服务器实例。
// The createHTTPServer method creates and configures an HTTP server instance.
func (e *Engine) createHTTPServer() *http.Server {
	return &http.Server{
		Addr:              e.endpoint,                                                       // 服务器监听地址 (Server listen address)
		Handler:           e.ginSvr,                                                         // Gin 引擎处理器 (Gin engine handler)
		ReadTimeout:       time.Duration(e.config.HttpReadTimeout) * time.Millisecond,       // 读取超时时间 (Read timeout)
		ReadHeaderTimeout: time.Duration(e.config.HttpReadHeaderTimeout) * time.Millisecond, // 读取头部超时时间 (Read header timeout)
		WriteTimeout:      time.Duration(e.config.HttpWriteTimeout) * time.Millisecond,      // 写入超时时间 (Write timeout)
		IdleTimeout:       0,                                                                // 空闲超时时间，0 表示不限制 (Idle timeout, 0 means no limit)
		MaxHeaderBytes:    math.MaxUint32,                                                   // 最大头部字节数 (Maximum header bytes)
		ErrorLog:          ilog.NewStandardLoggerFromLogr(e.config.logger),                  // 错误日志记录器 (Error logger)
	}
}

// startHTTPServer 方法启动 HTTP 服务器并处理可能的错误。
// The startHTTPServer method starts the HTTP server and handles potential errors.
func (e *Engine) startHTTPServer() {
	defer e.wg.Done() // 确保在函数退出时减少等待组计数 (Ensure wait group count is decreased when function exits)

	// 启用 Keep-Alive
	// Enable Keep-Alive
	e.httpSvr.SetKeepAlivesEnabled(true)
	e.config.logger.Info("http server is ready", "address", e.endpoint)

	// 启动服务器并处理错误
	// Start server and handle errors
	if err := e.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		e.config.logger.Error(err, "failed to start http server", "address", e.endpoint)
	}
}

// Stop 方法优雅地停止 HTTP 服务器。
// The Stop method gracefully stops the HTTP server.
func (e *Engine) Stop() {
	e.once.Do(func() {
		// 更新服务器状态为停止
		// Update server status to stopped
		e.updateRunningState(false)

		// 关闭 HTTP 服务器
		// Shutdown HTTP server
		e.shutdownHTTPServer()

		// 取消上下文并等待所有协程完成
		// Cancel context and wait for all goroutines to complete
		e.cancel()
		e.wg.Wait()

		// 如果启用了指标收集，注销指标收集器
		// If metrics are enabled, unregister the metrics collector
		if e.opts.metric {
			e.metric.Unregister()
		}
	})
}

// shutdownHTTPServer 方法优雅地关闭 HTTP 服务器。
// The shutdownHTTPServer method gracefully shuts down the HTTP server.
func (e *Engine) shutdownHTTPServer() {
	if e.httpSvr == nil {
		return
	}

	// 创建带超时的上下文
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	// 尝试优雅关闭服务器
	// Attempt to gracefully shutdown the server
	if err := e.httpSvr.Shutdown(ctx); err != nil {
		e.config.logger.Error(err, "http server forced to shutdown", "address", e.endpoint)
	}
	e.config.logger.Info("http server is shutdown", "address", e.endpoint)
}

// IsRunning 方法返回服务器的运行状态。
// The IsRunning method returns the server's running status.
func (e *Engine) IsRunning() bool {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.running
}

// updateRunningState 方法更新服务器的运行状态。
// The updateRunningState method updates the server's running status.
func (e *Engine) updateRunningState(status bool) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.running = status
}

// registerUserServices 方法注册用户定义的服务。
// The registerUserServices method registers user-defined services.
func (e *Engine) registerUserServices() {
	e.lock.Lock()
	defer e.lock.Unlock()

	// 只在服务器未运行时注册服务
	// Register services only when server is not running
	if !e.running {
		for i := 0; i < len(e.services); i++ {
			e.services[i].RegisterGroup(e.root)
		}
	}
}

// RegisterService 方法添加用户定义的服务到服务列表中。
// The RegisterService method adds a user-defined service to the service list.
func (e *Engine) RegisterService(service Service) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.services = append(e.services, service)
	}
}

// RegisterMiddleware 方法添加中间件到处理器列表中。
// The RegisterMiddleware method adds middleware to the handler list.
func (e *Engine) RegisterMiddleware(handler gin.HandlerFunc) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.handlers = append(e.handlers, handler)
	}
}

// registerUserMiddlewares 方法注册用户定义的中间件。
// The registerUserMiddlewares method registers user-defined middleware.
func (e *Engine) registerUserMiddlewares() {
	e.lock.Lock()
	defer e.lock.Unlock()

	// 只在服务器未运行时注册中间件
	// Register middleware only when server is not running
	if !e.running {
		e.ginSvr.Use(e.handlers...)
	}
}

// IsMetricEnabled 方法返回是否启用了指标收集功能。
// The IsMetricEnabled method returns whether metric collection is enabled.
func (e *Engine) IsMetricEnabled() bool {
	return e.opts.metric
}

// IsReleaseMode 方法返回服务器是否运行在发布模式。
// The IsReleaseMode method returns whether the server is running in release mode.
func (e *Engine) IsReleaseMode() bool {
	return e.config.ReleaseMode
}

// GetLogger 方法返回服务器的日志记录器。
// The GetLogger method returns the server's logger.
func (e *Engine) GetLogger() *logr.Logger {
	return e.config.logger
}

// GetPrometheusRegistry 方法返回 Prometheus 注册表。
// The GetPrometheusRegistry method returns the Prometheus registry.
func (e *Engine) GetPrometheusRegistry() *prometheus.Registry {
	return e.config.prometheusRegistry
}

// GetListenEndpoint 方法返回服务器的监听地址。
// The GetListenEndpoint method returns the server's listen endpoint.
func (e *Engine) GetListenEndpoint() string {
	return e.endpoint
}
