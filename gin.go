package orbit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

// HTTP 连接的默认空闲超时时间（秒）
const defaultHttpIdleTimeoutSeconds = int(com.DefaultHttpIdleTimeoutMillis / 1000)

// Engine 生命周期状态：new -> running -> stopped，stopped 为终态（单次使用语义）
const (
	stateNew = iota
	stateRunning
	stateStopped
)

// 引擎不处于 stateNew 状态时拒绝注册服务或中间件返回的错误
var errRegistrationRejected = errors.New("registration rejected: engine is not in new state")

// Service 接口定义了注册路由组的方法
type Service interface {
	RegisterGroup(routerGroup *gin.RouterGroup)
}

// Engine 结构体是 Orbit 框架的核心引擎，包含了 HTTP 服务器和相关配置
type Engine struct {
	endpoint string
	ginSvr   *gin.Engine
	root     *gin.RouterGroup
	config   *Config
	opts     *Options
	state    atomic.Int32
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	handlers []gin.HandlerFunc
	services []Service
	metric   *mtc.ServerMetrics
	initErr  error
	runErrMu sync.Mutex
	runErr   error

	// mu 守护生命周期装配段与注册列表：Run 的 state 复查、httpSvr 赋值与 wg.Add，
	// Stop 的 httpSvr 快照，以及 RegisterService/RegisterMiddleware 的 state 检查与 append，
	// 消除并发 Run/Stop/Register* 之间的竞态窗口（httpSvr 无同步读写、wg.Add 与 wg.Wait 并发、check-then-act）
	mu      sync.Mutex
	httpSvr *http.Server
}

// NewEngine 创建并返回一个新的引擎实例。
//
// 注意：本函数具有进程级全局副作用——config.ReleaseMode 为 true 时调用
// gin.SetMode(gin.ReleaseMode)，且无条件调用 gin.DisableConsoleColor()。
// 以不同 ReleaseMode 创建多个 Engine 时，后创建的实例会静默覆盖先前的 gin 全局模式。
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

	e.ginSvr.ForwardedByClientIP = options.forwardByClientIp
	e.ginSvr.RemoteIPHeaders = cloneStringSlice(e.config.RemoteIPHeaders)
	if options.forwardByClientIp {
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
		c.String(http.StatusNotFound, "[404] http request route mismatch, method: "+c.Request.Method+", path: "+c.Request.URL.Path)
	})

	// 设置 405 方法不允许的处理函数
	e.ginSvr.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "[405] http request method not allowed, method: "+c.Request.Method+", path: "+c.Request.URL.Path)
	})

	// 注册基本中间件
	e.ginSvr.Use(mid.Recovery(e.config.logger, e.config.recoveryLogEventFunc, e.opts.recReqBody)) // 恢复中间件
	// 访问日志中间件：紧随 Recovery 注册，先于用户中间件，保证被用户中间件 Abort 的请求仍产生访问日志
	e.ginSvr.Use(mid.AccessLogger(e.config.logger, e.config.accessLogEventFunc, e.opts.recReqBody))
	if e.opts.recRespBody {
		e.ginSvr.Use(mid.BodyBuffer()) // 响应体缓冲中间件
	}
	e.ginSvr.Use(mid.CorsWithPolicy(*e.config.CORSPolicy)) // CORS 中间件
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
		pprofService(e.root.Group(com.PprofURLPath), e.config.logger) // 注册 pprof 服务
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

	// 并发守卫：仅允许首个调用者将状态从 stateNew 迁移到 stateRunning，引擎为单次使用语义
	if !e.state.CompareAndSwap(stateNew, stateRunning) {
		return
	}

	// 装配段在 mu 内复查 state 消除与并发 Stop 的竞态窗口，并取注册列表快照。
	// 用户代码（RegisterGroup/ginSvr.Use）在锁外执行：避免用户在注册回调内
	// 重入 RegisterService/RegisterMiddleware/Stop 时因 mu 不可重入而死锁。
	// 快照安全性：CAS 成功后 state != new，Register* 的锁内检查必拒绝 append，
	// 快照切片事实不可变，锁外迭代无竞态
	e.mu.Lock()
	if e.state.Load() != stateRunning {
		e.mu.Unlock()
		return
	}
	services := e.services
	handlers := e.handlers
	e.mu.Unlock()

	// 注册用户中间件和服务（锁外迭代快照）
	e.ginSvr.Use(handlers...)
	for _, service := range services {
		service.RegisterGroup(e.root)
	}

	// 二次锁段：复查 state（并发 Stop 可能在用户代码执行期间完成关闭）后，
	// httpSvr 赋值与 wg.Add 仍在 mu 内原子完成，
	// 保证二者先于任何 Stop 调用者取得快照（wg.Add 不与 wg.Wait 并发）
	e.mu.Lock()
	if e.state.Load() != stateRunning {
		e.mu.Unlock()
		return
	}
	e.httpSvr = e.createHTTPServer()
	svr := e.httpSvr
	e.wg.Add(1)
	e.mu.Unlock()

	go e.startHTTPServer(svr)
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

// 启动 HTTP 服务器并处理可能的错误。
// svr 由 Run 在 mu 锁内的装配段传入，避免 goroutine 无锁读取 e.httpSvr
func (e *Engine) startHTTPServer(svr *http.Server) {
	defer e.wg.Done()
	svr.SetKeepAlivesEnabled(true)
	e.config.logger.Info("http server is ready", "address", e.endpoint)
	if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		e.config.logger.Error(err, "failed to start http server", "address", e.endpoint)
		e.runErrMu.Lock()
		e.runErr = err
		e.runErrMu.Unlock()
		e.state.Store(stateStopped)
		e.cancel()
		if e.opts.metric {
			e.metric.Unregister()
		}
	}
}

// Stop 使用配置的 ShutdownTimeout 超时上下文优雅地停止 HTTP 服务器。
// 它是 StopWithContext 的薄封装，使用 context.Background() 与 ShutdownTimeout 构建关闭上下文。
func (e *Engine) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.config.ShutdownTimeout)*time.Millisecond)
	defer cancel()
	e.StopWithContext(ctx)
}

// StopWithContext 使用传入的上下文优雅地停止 HTTP 服务器。
// 调用者负责提供有效的 ctx（不可为 nil），关闭超时由调用者通过 ctx 控制；
// Stop() 封装提供 ShutdownTimeout 默认超时。
func (e *Engine) StopWithContext(ctx context.Context) {
	// running->stopped 为正常关闭；new->stopped 为 Stop-before-Run；保证幂等。
	// 并发 Run 可能在两次 CAS 之间将 state 从 new 迁移到 running（两次 CAS 均失败但引擎在运行），
	// 此时重试 CAS；状态机单调迁移（stopped 为终态），循环最多两次迭代即收敛
	for {
		if e.state.CompareAndSwap(stateRunning, stateStopped) || e.state.CompareAndSwap(stateNew, stateStopped) {
			break
		}
		if e.state.Load() == stateStopped {
			// 后到的 Stop 调用者：等待首个 Stop 完成排水后再返回，保证 Stop 返回即"已停止"
			e.waitForStopped(ctx)
			return
		}
	}

	// 锁内取 httpSvr 快照后关闭 HTTP 服务器（与 Run 的装配段互斥，消除无同步读写）
	e.mu.Lock()
	svr := e.httpSvr
	e.mu.Unlock()
	e.shutdownHTTPServer(ctx, svr)

	// 取消上下文并等待所有协程完成
	e.cancel()
	e.wg.Wait()

	// 如果启用了指标收集，注销指标收集器
	if e.opts.metric {
		e.metric.Unregister()
	}
}

// waitForStopped 使后到的 Stop 调用者等待首个 Stop 调用者完成在途请求排水与服务协程退出。
// 先取得 mu 确保 Run 的装配段（httpSvr 赋值与 wg.Add）已结束，避免 wg.Add 与 wg.Wait 并发；
// http.Server.Shutdown 可安全重复调用：排水未完成时等待同一批在途请求，已完成时立即返回
func (e *Engine) waitForStopped(ctx context.Context) {
	e.mu.Lock()
	svr := e.httpSvr
	e.mu.Unlock()

	e.shutdownHTTPServer(ctx, svr)

	e.wg.Wait()
}

// 优雅地关闭 HTTP 服务器
func (e *Engine) shutdownHTTPServer(ctx context.Context, svr *http.Server) {
	if svr == nil {
		return
	}
	if err := svr.Shutdown(ctx); err != nil {
		e.config.logger.Error(err, "http server forced to shutdown", "address", e.endpoint)
	}
	e.config.logger.Info("http server is shutdown", "address", e.endpoint)
}

// 返回服务器的运行状态
func (e *Engine) IsRunning() bool {
	return e.state.Load() == stateRunning
}

// 添加用户定义的服务到服务列表中。
// state 检查与 append 在 mu 内原子完成，与 Run 的装配段互斥，消除 check-then-act 竞态
func (e *Engine) RegisterService(service Service) {
	if service == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.state.Load() != stateNew {
		e.config.logger.Error(errRegistrationRejected, "register service rejected", "service", fmt.Sprintf("%T", service))
		return
	}
	e.services = append(e.services, service)
}

// 添加中间件到处理器列表中。
// state 检查与 append 在 mu 内原子完成，与 Run 的装配段互斥，消除 check-then-act 竞态
func (e *Engine) RegisterMiddleware(handler gin.HandlerFunc) {
	if handler == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.state.Load() != stateNew {
		e.config.logger.Error(errRegistrationRejected, "register middleware rejected", "handler", fmt.Sprintf("%T", handler))
		return
	}
	e.handlers = append(e.handlers, handler)
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

// GetRunError 返回服务器监听启动失败的错误，未发生错误时返回 nil
func (e *Engine) GetRunError() error {
	e.runErrMu.Lock()
	defer e.runErrMu.Unlock()
	return e.runErr
}

// GetInitError 返回引擎初始化阶段（NewEngine）遇到的错误，初始化成功时返回 nil。
// 初始化失败时 Run 会拒绝启动，调用方可通过本方法程序化观测失败原因
func (e *Engine) GetInitError() error {
	return e.initErr
}

// GetGinEngine 返回底层的 Gin 引擎实例
func (e *Engine) GetGinEngine() *gin.Engine {
	return e.ginSvr
}
