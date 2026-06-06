package orbit

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
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
var defaultShutdownTimeout = time.Second * com.DefaultShutdownTimeoutSeconds

// HTTP 连接的默认空闲超时时间（秒）
const defaultHttpIdleTimeoutSeconds = int(com.DefaultHttpIdleTimeoutMillis / 1000)

// Service 接口定义了注册路由组的方法
type Service interface {
	RegisterGroup(routerGroup *gin.RouterGroup)
}

// Engine 结构体是 Orbit 框架的核心引擎，包含了 HTTP 服务器和相关配置
type Engine struct {
	endpoint string
	ginSvr   *gin.Engine
	httpSvr  *http.Server
	root     *gin.RouterGroup
	config   *Config
	opts     *Options
	running  atomic.Bool
	wg       sync.WaitGroup
	once     sync.Once
	ctx      context.Context
	cancel   context.CancelFunc
	handlers []gin.HandlerFunc
	services []Service
	metric   *mtc.ServerMetrics
	initErr  error
	runErrMu sync.Mutex
	runErr   error
}

// NewEngine 创建并返回一个新的引擎实例
func NewEngine(config *Config, options *Options) *Engine {
	// 验证配置和选项的有效性
	config = isConfigValid(config)
	options = isOptionsValid(options)

	// 如果是发布模式，设置 Gin 为发布模式并禁用控制台颜色
	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DisableConsoleColor()

	// 创建引擎实例并初始化基本属性
	engine := &Engine{
		endpoint: fmt.Sprintf("%s:%d", config.Address, config.Port), // 设置服务器监听地址
		config:   config,
		opts:     options,
		handlers: make([]gin.HandlerFunc, 0, 10),
		services: make([]Service, 0, 10),
		metric:   mtc.NewServerMetrics(config.prometheusRegistry),
	}

	// 创建可取消的上下文，用于服务器生命周期管理
	engine.ctx, engine.cancel = context.WithCancel(context.Background())

	// 初始化 Gin 引擎并设置基本配置
	engine.initErr = engine.initGinEngine(options)

	if engine.initErr == nil {
		// 注册内置服务（健康检查、Swagger、Pprof、指标收集等）
		engine.registerBuiltinServices()
	}

	return engine
}

// 初始化 Gin 引擎并设置基本配置
func (e *Engine) initGinEngine(options *Options) error {
	e.ginSvr = gin.New()
	e.root = &e.ginSvr.RouterGroup

	e.ginSvr.ForwardedByClientIP = options.forwordByClientIp
	e.ginSvr.RemoteIPHeaders = cloneStringSlice(e.config.RemoteIPHeaders)
	if options.forwordByClientIp {
		if err := e.ginSvr.SetTrustedProxies(cloneStringSlice(e.config.TrustedProxies)); err != nil {
			return fmt.Errorf("failed to set trusted proxies %v: %w", e.config.TrustedProxies, err)
		}
	}
	e.ginSvr.RedirectTrailingSlash = options.trailingSlash
	e.ginSvr.RedirectFixedPath = options.fixedPath
	e.ginSvr.HandleMethodNotAllowed = true

	e.setupBaseHandlers()
	return nil
}

// 设置基本的 HTTP 处理函数，包括 404、405 处理和中间件
func (e *Engine) setupBaseHandlers() {
	// 设置 404 路由未匹配的处理函数
	e.ginSvr.NoRoute(func(c *gin.Context) {
		var sb strings.Builder
		sb.WriteString("[404] http request route mismatch, method: ")
		sb.WriteString(c.Request.Method)
		sb.WriteString(", path: ")
		sb.WriteString(c.Request.URL.Path)
		c.String(http.StatusNotFound, sb.String())
	})

	// 设置 405 方法不允许的处理函数
	e.ginSvr.NoMethod(func(c *gin.Context) {
		var sb strings.Builder
		sb.WriteString("[405] http request method not allowed, method: ")
		sb.WriteString(c.Request.Method)
		sb.WriteString(", path: ")
		sb.WriteString(c.Request.URL.Path)
		c.String(http.StatusMethodNotAllowed, sb.String())
	})

	// 注册基本中间件
	e.ginSvr.Use(
		mid.Recovery(e.config.logger, e.config.recoveryLogEventFunc), // 恢复中间件
		mid.BodyBuffer(),                         // 请求体缓冲中间件
		mid.CorsWithPolicy(*e.config.CORSPolicy), // CORS 中间件
	)
}

// 注册内置的服务，包括健康检查、Swagger、pprof 和指标收集等
func (e *Engine) registerBuiltinServices() {
	// 根据配置注册可选服务
	if e.opts.healthCheck {
		healthcheckService(e.root.Group(com.HealthCheckURLPath)) // 注册健康检查服务
	}
	if e.opts.swagger {
		swaggerService(e.root.Group(com.SwaggerURLPath)) // 注册 Swagger 服务
	}
	if e.opts.pprof {
		pprofService(e.root.Group(com.PprofURLPath)) // 注册 pprof 服务
	}
	if e.opts.metric {
		e.setupMetricService() // 注册指标收集服务
	}
}

// 设置并注册 Prometheus 指标收集服务
func (e *Engine) setupMetricService() {
	e.metric.Register()                                                                              // 注册指标收集器
	e.ginSvr.Use(e.metric.HandlerFunc(e.config.logger))                                              // 添加指标收集中间件
	metricService(e.root.Group(com.PromMetricURLPath), e.config.prometheusRegistry, e.config.logger) // 注册指标服务路由
}

// 启动 HTTP 服务器
func (e *Engine) Run() {
	if e.initErr != nil {
		e.config.logger.Error(e.initErr, "engine initialization failed, startup aborted")
		return
	}

	// 检查服务器是否已经在运行
	if e.IsRunning() {
		return
	}

	// 注册用户中间件和服务
	e.registerUserMiddlewares()
	e.ginSvr.Use(mid.AccessLogger(e.config.logger, e.config.accessLogEventFunc, e.opts.recReqBody))
	e.registerUserServices()

	// 创建并启动 HTTP 服务器
	e.httpSvr = e.createHTTPServer()
	e.wg.Add(1)

	go e.startHTTPServer()

	// 更新服务器状态
	e.updateRunningState(true)
}

// 创建并配置 HTTP 服务器实例
func (e *Engine) createHTTPServer() *http.Server {
	// 使用合理的 MaxHeaderBytes 值
	maxHeaderBytes := com.DefaultMaxHeaderBytes
	if e.config.MaxHeaderBytes > 0 {
		maxHeaderBytes = int(e.config.MaxHeaderBytes)
	}

	// 设置合理的空闲超时时间
	idleTimeout := time.Duration(e.config.HttpIdleTimeout) * time.Millisecond
	if idleTimeout <= 0 {
		idleTimeout = time.Duration(defaultHttpIdleTimeoutSeconds) * time.Second // 默认 15 秒
	}

	return &http.Server{
		Addr:              e.endpoint,                                                       // 服务器监听地址
		Handler:           e.ginSvr,                                                         // Gin 引擎处理器
		ReadTimeout:       time.Duration(e.config.HttpReadTimeout) * time.Millisecond,       // 读取超时时间
		ReadHeaderTimeout: time.Duration(e.config.HttpReadHeaderTimeout) * time.Millisecond, // 读取头部超时时间
		WriteTimeout:      time.Duration(e.config.HttpWriteTimeout) * time.Millisecond,      // 写入超时时间
		IdleTimeout:       idleTimeout,                                                      // 空闲超时时间
		MaxHeaderBytes:    maxHeaderBytes,                                                   // 最大头部字节数
		ErrorLog:          ilog.NewStandardLoggerFromLogr(e.config.logger),                  // 错误日志记录器
	}
}

// 启动 HTTP 服务器并处理可能的错误
func (e *Engine) startHTTPServer() {
	defer e.wg.Done()
	e.httpSvr.SetKeepAlivesEnabled(true)
	e.config.logger.Info("http server is ready", "address", e.endpoint)
	if err := e.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		e.config.logger.Error(err, "failed to start http server", "address", e.endpoint)
		e.runErrMu.Lock()
		e.runErr = err
		e.runErrMu.Unlock()
	}
}

// 优雅地停止 HTTP 服务器
func (e *Engine) Stop() {
	e.once.Do(func() {
		// 更新服务器状态为停止
		e.updateRunningState(false)

		shutdownCtx, shutdownCancel := context.WithTimeout(e.ctx, defaultShutdownTimeout)
		defer shutdownCancel()

		// 关闭 HTTP 服务器
		e.shutdownHTTPServer(shutdownCtx)

		// 取消上下文并等待所有协程完成
		e.cancel()
		e.wg.Wait()

		// 如果启用了指标收集，注销指标收集器
		if e.opts.metric {
			e.metric.Unregister()
		}
	})
}

// 优雅地关闭 HTTP 服务器
func (e *Engine) shutdownHTTPServer(ctx context.Context) {
	if e.httpSvr == nil {
		return
	}
	if err := e.httpSvr.Shutdown(ctx); err != nil {
		e.config.logger.Error(err, "http server forced to shutdown", "address", e.endpoint)
	}
	e.config.logger.Info("http server is shutdown", "address", e.endpoint)
}

// 返回服务器的运行状态
func (e *Engine) IsRunning() bool {
	return e.running.Load()
}

// 更新服务器的运行状态
func (e *Engine) updateRunningState(status bool) {
	e.running.Store(status)
}

// 注册用户定义的服务
func (e *Engine) registerUserServices() {
	// 只在服务器未运行时注册服务
	if !e.running.Load() {
		for i := 0; i < len(e.services); i++ {
			e.services[i].RegisterGroup(e.root)
		}
	}
}

// 添加用户定义的服务到服务列表中
func (e *Engine) RegisterService(service Service) {
	if !e.running.Load() && service != nil {
		e.services = append(e.services, service)
	}
}

// 添加中间件到处理器列表中
func (e *Engine) RegisterMiddleware(handler gin.HandlerFunc) {
	if !e.running.Load() && handler != nil {
		e.handlers = append(e.handlers, handler)
	}
}

// 注册用户定义的中间件
func (e *Engine) registerUserMiddlewares() {
	// 只在服务器未运行时注册中间件
	if !e.running.Load() {
		e.ginSvr.Use(e.handlers...)
	}
}

// 返回是否启用了指标收集功能
func (e *Engine) IsMetricEnabled() bool {
	return e.opts.metric
}

// 返回服务器是否运行在发布模式
func (e *Engine) IsReleaseMode() bool {
	return e.config.ReleaseMode
}

// 返回服务器的日志记录器
func (e *Engine) GetLogger() *logr.Logger {
	return e.config.logger
}

// 返回 Prometheus 注册表
func (e *Engine) GetPrometheusRegistry() *prometheus.Registry {
	return e.config.prometheusRegistry
}

// 返回服务器的监听地址
func (e *Engine) GetListenEndpoint() string {
	return e.endpoint
}

func (e *Engine) GetRunError() error {
	e.runErrMu.Lock()
	defer e.runErrMu.Unlock()
	return e.runErr
}

func (e *Engine) GetGinEngine() *gin.Engine {
	return e.ginSvr
}
