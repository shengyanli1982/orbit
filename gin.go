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

// defaultShutdownTimeout 是优雅关闭的默认超时时间。
// defaultShutdownTimeout is the default timeout for graceful shutdown.
var defaultShutdownTimeout = 10 * time.Second

// Service 是表示服务的接口。
// Service is the interface that represents a service.
type Service interface {
	// RegisterGroup 方法用于将服务注册到路由组。
	// The RegisterGroup method is used to register the service to the router group.
	RegisterGroup(routerGroup *gin.RouterGroup)
}

// Engine 是表示 Orbit 引擎的主要结构体。
// Engine is the main struct that represents the Orbit engine.
type Engine struct {
	// running 表示引擎是否正在运行。
	// running indicates whether the engine is running.
	running bool

	// endpoint 是引擎的端点。
	// endpoint is the endpoint of the engine.
	endpoint string

	// ginSvr 是 Gin 引擎。
	// ginSvr is the Gin engine.
	ginSvr *gin.Engine

	// httpSvr 是 HTTP 服务器。
	// httpSvr is the HTTP server.
	httpSvr *http.Server

	// root 是根路由组。
	// root is the root router group.
	root *gin.RouterGroup

	// config 是引擎配置。
	// config is the engine configuration.
	config *Config

	// opts 是引擎选项。
	// opts are the engine options.
	opts *Options

	// lock 是用于并发访问的互斥锁。
	// lock is a mutex for concurrent access.
	lock sync.RWMutex

	// wg 是用于优雅关闭的 WaitGroup。
	// wg is a WaitGroup for graceful shutdown.
	wg sync.WaitGroup

	// once 是用于优雅关闭的 Once。
	// once is a Once for graceful shutdown.
	once sync.Once

	// ctx 是用于优雅关闭的上下文。
	// ctx is a context for graceful shutdown.
	ctx context.Context

	// cancel 是用于优雅关闭的取消函数。
	// cancel is a cancel function for graceful shutdown.
	cancel context.CancelFunc

	// handlers 是中间件处理器的列表。
	// handlers is a list of middleware handlers.
	handlers []gin.HandlerFunc

	// services 是已注册服务的列表。
	// services is a list of registered services.
	services []Service

	// metric 是 Prometheus 指标。
	// metric is a Prometheus metric.
	metric *mtc.ServerMetrics
}

// NewEngine 函数用于创建一个新的 Engine 实例。
// The NewEngine function is used to create a new instance of the Engine.
func NewEngine(config *Config, options *Options) *Engine {
	// 验证配置是否有效
	// Validate if the config is valid
	config = isConfigValid(config)

	// 验证选项是否有效
	// Validate if the options are valid
	options = isOptionsValid(options)

	// 如果配置的运行模式为发布模式，则设置 Gin 的运行模式为发布模式
	// If the running mode in the config is release mode, then set the running mode of Gin to release mode
	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// 禁用控制台颜色
	// Disable console color
	gin.DisableConsoleColor()

	// 创建一个新的 Engine 实例
	// Create a new instance of the Engine
	engine := Engine{
		// 设置 running 为 false，表示引擎初始状态为未运行
		// Set running to false, indicating that the initial state of the engine is not running
		running: false,

		// 使用配置中的地址和端口设置 endpoint
		// Set endpoint using the address and port in the config
		endpoint: fmt.Sprintf("%s:%d", config.Address, config.Port),

		// 将传入的 config 赋值给 engine 的 config
		// Assign the incoming config to the config of the engine
		config: config,

		// 将传入的 options 赋值给 engine 的 opts
		// Assign the incoming options to the opts of the engine
		opts: options,

		// 初始化一个空的互斥锁
		// Initialize an empty mutex lock
		lock: sync.RWMutex{},

		// 初始化一个空的 WaitGroup，用于等待所有 goroutine 完成
		// Initialize an empty WaitGroup, used to wait for all goroutines to complete
		wg: sync.WaitGroup{},

		// 初始化一个空的 Once，用于确保某个操作只执行一次
		// Initialize an empty Once, used to ensure that an operation is performed only once
		once: sync.Once{},

		// 初始化一个空的 gin.HandlerFunc 切片，用于存储中间件处理函数
		// Initialize an empty slice of gin.HandlerFunc, used to store middleware handlers
		handlers: make([]gin.HandlerFunc, 0),

		// 初始化一个空的 Service 切片，用于存储已注册的服务
		// Initialize an empty slice of Service, used to store registered services
		services: make([]Service, 0),

		// 创建一个新的 ServerMetrics 实例，用于收集和报告 HTTP 服务器的指标
		// Create a new ServerMetrics instance, used to collect and report metrics of the HTTP server
		metric: mtc.NewServerMetrics(config.prometheusRegistry),
	}

	// 创建一个新的上下文，并设置默认的优雅关闭超时时间
	// Create a new context and set the default graceful shutdown timeout
	engine.ctx, engine.cancel = context.WithTimeout(context.Background(), defaultShutdownTimeout)

	// 创建一个新的 Gin 引擎
	// Create a new Gin engine
	engine.ginSvr = gin.New()

	// 创建根路由组
	// (*) 所有默认中间件都注册到根路由组，不要更改这一点
	// Create root router group
	// (*) all default middlewares are registered to the root router group, don't change this
	engine.root = &engine.ginSvr.RouterGroup

	// 设置 Gin 引擎的 ForwardedByClientIP 选项。如果为 true，服务器将使用 HTTP 头部的 X-Forwarded-For 属性来获取客户端的 IP 地址。
	// Set the ForwardedByClientIP option of the Gin engine. If true, the server will use the X-Forwarded-For property of the HTTP header to get the client's IP address.
	engine.ginSvr.ForwardedByClientIP = options.forwordByClientIp

	// 设置 Gin 引擎的 RedirectTrailingSlash 选项。如果为 true，服务器将会自动重定向路径末尾带斜线和不带斜线的请求。
	// Set the RedirectTrailingSlash option of the Gin engine. If true, the server will automatically redirect requests with and without a trailing slash at the end of the path.
	engine.ginSvr.RedirectTrailingSlash = options.trailingSlash

	// 设置 Gin 引擎的 RedirectFixedPath 选项。如果为 true，服务器将会自动重定向大小写不匹配的请求路径。
	// Set the RedirectFixedPath option of the Gin engine. If true, the server will automatically redirect requests with case-insensitive path.
	engine.ginSvr.RedirectFixedPath = options.fixedPath

	// 设置 Gin 引擎的 HandleMethodNotAllowed 属性为 true。
	// 这意味着如果客户端发送了服务器不允许的 HTTP 方法（例如，服务器只允许 GET 和 POST，但客户端发送了 PUT），服务器将返回一个 405 状态码，表示“方法不被允许”。
	// Set the HandleMethodNotAllowed property of the Gin engine to true.
	// This means that if the client sends an HTTP method that the server does not allow (for example, the server only allows GET and POST, but the client sends PUT), the server will return a 405 status code, indicating "Method Not Allowed".
	engine.ginSvr.HandleMethodNotAllowed = true

	// 设置 Gin 引擎的 NoRoute 处理函数
	// Set the NoRoute handler function of the Gin engine
	engine.ginSvr.NoRoute(func(context *gin.Context) {
		context.String(http.StatusNotFound, "[404] http request route mismatch, method: "+context.Request.Method+", path: "+context.Request.URL.Path)
	})

	// 设置 Gin 引擎的 NoMethod 处理函数
	// Set the NoMethod handler function of the Gin engine
	engine.ginSvr.NoMethod(func(context *gin.Context) {
		context.String(http.StatusMethodNotAllowed, "[405] http request method not allowed, method: "+context.Request.Method+", path: "+context.Request.URL.Path)
	})

	// 注册健康检查服务
	// Register health check service
	healthcheckService(engine.root.Group(com.HealthCheckURLPath))

	// 如果启用了 swagger，则注册 swagger 服务
	// If swagger is enabled, register swagger service
	if engine.opts.swagger {
		swaggerService(engine.root.Group(com.SwaggerURLPath))
	}

	// 如果启用了 pprof，则注册 pprof 服务
	// If pprof is enabled, register pprof service
	if engine.opts.pprof {
		pprofService(engine.root.Group(com.PprofURLPath))
	}

	// 如果启用了 metric，则注册 metric 服务
	// If metric is enabled, register metric service
	// 如果启用了 metric 选项
	// If the metric option is enabled
	if engine.opts.metric {
		// 注册 metric
		// Register metric
		engine.metric.Register()

		// 将 metric 的处理函数添加到 Gin 引擎的中间件中
		// Add the handler function of metric to the middleware of the Gin engine
		engine.ginSvr.Use(engine.metric.HandlerFunc(engine.config.logger))

		// 在指定的路径上注册 metric 服务
		// Register the metric service on the specified path
		metricService(engine.root.Group(com.PromMetricURLPath), engine.config.prometheusRegistry, engine.config.logger)
	}

	// 注册中间件
	// Register middleware
	engine.ginSvr.Use(
		// Recovery 是一个中间件，用于恢复可能在处理 HTTP 请求时发生的 panic，并将错误记录到日志中。
		// Recovery is a middleware that recovers from any panics that might occur during the handling of an HTTP request and logs the error.
		mid.Recovery(engine.config.logger, engine.config.recoveryLogEventFunc),

		// BodyBuffer 是一个中间件，用于读取和存储请求体，以便在后续的处理中重复使用。
		// BodyBuffer is a middleware that reads and stores the request body for reuse in subsequent processing.
		mid.BodyBuffer(),

		// Cors 是一个中间件，用于处理跨域资源共享（CORS）的请求。
		// Cors is a middleware that handles Cross-Origin Resource Sharing (CORS) requests.
		mid.Cors(),
	)

	// 返回新创建的 Engine 实例
	// Return the newly created Engine instance
	return &engine
}

// Run 方法用于启动 Engine
// The Run method is used to start the Engine
func (e *Engine) Run() {
	// 如果 Engine 已经在运行，就直接返回
	// If the Engine is already running, return directly
	if e.IsRunning() {
		return
	}

	// 注册所有中间件
	// Register all middlewares
	e.registerAllMiddlewares()

	// 使用 AccessLogger 中间件记录访问日志
	// Use the AccessLogger middleware to log access
	e.ginSvr.Use(mid.AccessLogger(e.config.logger, e.config.accessLogEventFunc, e.opts.recReqBody))

	// 注册所有服务
	// Register all services
	e.registerAllServices()

	// 初始化 http 服务器
	// Initialize the http server
	e.httpSvr = &http.Server{
		// 设置服务器监听的地址和端口
		// Set the address and port the server listens on
		Addr: e.endpoint,

		// 设置处理请求的 Handler，这里使用 Gin 引擎
		// Set the Handler that handles the request, here use the Gin engine
		Handler: e.ginSvr,

		// 设置读取请求体的超时时间
		// Set the timeout for reading the request body
		ReadTimeout: time.Duration(e.config.HttpReadTimeout) * time.Millisecond,

		// 设置读取请求头的超时时间
		// Set the timeout for reading the request header
		ReadHeaderTimeout: time.Duration(e.config.HttpReadHeaderTimeout) * time.Millisecond,

		// 设置写入响应的超时时间
		// Set the timeout for writing the response
		WriteTimeout: time.Duration(e.config.HttpWriteTimeout) * time.Millisecond,

		// 设置空闲连接的超时时间，这里使用 HttpReadTimeout 的值
		// Set the timeout for idle connections, here use the value of HttpReadTimeout
		IdleTimeout: 0,

		// 设置请求头的最大字节数
		// Set the maximum number of bytes in the request header
		MaxHeaderBytes: math.MaxUint32,

		// 设置错误日志，这里使用 zap 日志库
		// Set the error log, here use the zap logging library
		ErrorLog: zap.NewStdLog(e.config.logger.Desugar()),
	}

	// 增加等待组的计数
	// Increase the count of the wait group
	e.wg.Add(1)

	// 在新的 goroutine 中启动 http 服务器
	// Start the http server in a new goroutine
	go func() {
		// 在 goroutine 结束时减少等待组的计数
		// Decrease the count of the wait group when the goroutine ends
		defer e.wg.Done()

		// 设置服务器保持活动连接
		// Set the server to keep active connections
		e.httpSvr.SetKeepAlivesEnabled(true)

		// 记录日志，表示 http 服务器已准备好
		// Log that the http server is ready
		e.config.logger.Infow("http server is ready", "address", e.endpoint)

		// 启动 http 服务器，如果启动失败并且错误不是服务器已关闭，就记录致命错误日志
		// Start the http server, if it fails to start and the error is not that the server has been closed, log a fatal error
		if err := e.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.config.logger.Fatalw("failed to start http server", "error", err)
		}
	}()

	// 设置 Engine 的运行状态为 true
	// Set the running status of the Engine to true
	e.setRuningStatus(true)

	// 重置 once，确保下一次调用 Run 方法时，可以再次执行其中的代码
	// Reset once to ensure that the next time the Run method is called, the code in it can be executed again
	e.once = sync.Once{}
}

// Stop 方法用于停止 Engine
// The Stop method is used to stop the Engine
func (e *Engine) Stop() {
	// 使用 once.Do 确保以下的代码只执行一次
	// Use once.Do to ensure that the following code is executed only once
	e.once.Do(func() {
		// 设置 Engine 的运行状态为 false
		// Set the running status of the Engine to false
		e.setRuningStatus(false)

		// 如果 http 服务器不为 nil
		// If the http server is not nil
		if e.httpSvr != nil {
			// 尝试优雅地关闭 http 服务器，如果失败就记录致命错误日志
			// Try to gracefully shut down the http server, if it fails, log a fatal error
			if err := e.httpSvr.Shutdown(e.ctx); err != nil {
				e.config.logger.Fatalw("http server forced to shutdown", "address", e.endpoint, "error", err)
			}
		}
		// 记录日志，表示 http 服务器已关闭
		// Log that the http server has been shut down
		e.config.logger.Infow("http server is shutdown", "address", e.endpoint)

		// 取消所有的 context
		// Cancel all contexts
		e.cancel()

		// 等待所有的 goroutine 结束
		// Wait for all goroutines to finish
		e.wg.Wait()

		// 如果启用了 metric 选项
		// If the metric option is enabled
		if e.opts.metric {
			// 注销 metric
			// Unregister metric
			e.metric.Unregister()
		}
	})
}

// IsRunning 方法返回 Orbit 引擎是否正在运行。
// The IsRunning method returns whether the Orbit engine is running.
func (e *Engine) IsRunning() bool {
	// 获取写锁
	// Acquire the write lock
	e.lock.Lock()
	defer e.lock.Unlock()

	// 返回引擎的运行状态
	// Return the running status of the engine
	return e.running
}

// setRuningStatus 方法设置 Orbit 引擎的运行状态。
// The setRuningStatus method sets the running status of the Orbit engine.
func (e *Engine) setRuningStatus(status bool) {
	// 获取写锁
	// Acquire the write lock
	e.lock.Lock()
	defer e.lock.Unlock()

	// 设置引擎的运行状态
	// Set the running status of the engine
	e.running = status
}

// registerAllServices 方法将所有服务注册到根路由组。
// The registerAllServices method registers all services to the root router group.
func (e *Engine) registerAllServices() {
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

// registerAllMiddlewares 方法将所有中间件注册到引擎。
// The registerAllMiddlewares method registers all middlewares to the engine.
func (e *Engine) registerAllMiddlewares() {
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

// RegisterService 方法将一个服务注册到 Orbit 引擎。
// The RegisterService method registers a service to the Orbit engine.
func (e *Engine) RegisterService(service Service) {
	// 获取写锁
	// Acquire the write lock
	e.lock.Lock()
	defer e.lock.Unlock()

	// 如果引擎没有在运行
	// If the engine is not running
	if !e.running {
		// 将服务添加到服务列表
		// Add the service to the service list
		e.services = append(e.services, service)
	}
}

// RegisterMiddleware 方法将一个中间件注册到 Orbit 引擎。
// The RegisterMiddleware method registers a middleware to the Orbit engine.
func (e *Engine) RegisterMiddleware(handler gin.HandlerFunc) {
	// 获取写锁
	// Acquire the write lock
	e.lock.Lock()
	defer e.lock.Unlock()

	// 如果引擎没有在运行
	// If the engine is not running
	if !e.running {
		// 将中间件添加到中间件列表
		// Add the middleware to the middleware list
		e.handlers = append(e.handlers, handler)
	}
}

// IsMetricEnabled 方法返回 Orbit 引擎的 metric 状态。
// The IsMetricEnabled method returns the metric status of the Orbit engine.
func (e *Engine) IsMetricEnabled() bool {
	// 返回引擎的 metric 选项
	// Return the metric option of the engine
	return e.opts.metric
}

// IsReleaseMode 方法返回 Orbit 引擎的运行模式。
// The IsReleaseMode method returns the running mode of the Orbit engine.
func (e *Engine) IsReleaseMode() bool {
	// 返回引擎的 ReleaseMode 配置
	// Return the ReleaseMode configuration of the engine
	return e.config.ReleaseMode
}

// GetLogger 方法返回 Orbit 引擎的日志记录器。
// The GetLogger method returns the logger of the Orbit engine.
func (e *Engine) GetLogger() *zap.SugaredLogger {
	// 返回引擎的 logger 配置
	// Return the logger configuration of the engine
	return e.config.logger
}

// GetPrometheusRegistry 方法返回 Orbit 引擎的 Prometheus 注册表。
// The GetPrometheusRegistry method returns the Prometheus registry of the Orbit engine.
func (e *Engine) GetPrometheusRegistry() *prometheus.Registry {
	// 返回引擎的 prometheusRegistry 配置
	// Return the prometheusRegistry configuration of the engine
	return e.config.prometheusRegistry
}

// GetListenEndpoint 方法返回 Orbit 引擎的监听端点。
// The GetListenEndpoint method returns the listen endpoint of the Orbit engine.
func (e *Engine) GetListenEndpoint() string {
	// 返回引擎的 endpoint 配置
	// Return the endpoint configuration of the engine
	return e.endpoint
}
