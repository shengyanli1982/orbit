package orbit

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	mid "github.com/shengyanli1982/orbit/internal/middleware"
	wrap "github.com/shengyanli1982/orbit/utils/wrapper"
	swag "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

var (
	defaultShutdownTimeout = 10 * time.Second
)

func pprofService(g *gin.RouterGroup) {
	// Get
	g.GET("/", wrap.HandlerFuncWrapToGin(pprof.Index))
	g.GET("/cmdline", wrap.HandlerFuncWrapToGin(pprof.Cmdline))
	g.GET("/profile", wrap.HandlerFuncWrapToGin(pprof.Profile))
	g.GET("/symbol", wrap.HandlerFuncWrapToGin(pprof.Symbol))
	g.GET("/trace", wrap.HandlerFuncWrapToGin(pprof.Trace))
	g.GET("/allocs", wrap.HandlerFuncWrapToGin(pprof.Handler("allocs").ServeHTTP))
	g.GET("/block", wrap.HandlerFuncWrapToGin(pprof.Handler("block").ServeHTTP))
	g.GET("/goroutine", wrap.HandlerFuncWrapToGin(pprof.Handler("goroutine").ServeHTTP))
	g.GET("/heap", wrap.HandlerFuncWrapToGin(pprof.Handler("heap").ServeHTTP))
	g.GET("/mutex", wrap.HandlerFuncWrapToGin(pprof.Handler("mutex").ServeHTTP))
	g.GET("/threadcreate", wrap.HandlerFuncWrapToGin(pprof.Handler("threadcreate").ServeHTTP))

	// Post
	g.POST("/pprof/symbol", wrap.HandlerFuncWrapToGin(pprof.Symbol))
}

type Service interface {
	RegisterGroup(routerGroup *gin.RouterGroup)
}

type Engine struct {
	running  bool
	endpoint string
	ginSvr   *gin.Engine
	httpSvr  *http.Server
	root     *gin.RouterGroup
	config   *Config
	opts     *Options
	lock     sync.RWMutex
	wg       sync.WaitGroup
	once     sync.Once
	ctx      context.Context
	cancel   context.CancelFunc
	handlers []gin.HandlerFunc
	services []Service
}

func NewEngine(config *Config, options *Options) *Engine {
	config = isConfigValid(config)
	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DisableConsoleColor()
	engine := Engine{
		running:  false,
		endpoint: fmt.Sprintf("%s:%d", config.Address, config.Port),
		config:   config,
		opts:     options,
		lock:     sync.RWMutex{},
		wg:       sync.WaitGroup{},
		once:     sync.Once{},
	}
	engine.ctx, engine.cancel = context.WithTimeout(context.Background(), defaultShutdownTimeout)
	engine.ginSvr = gin.New()
	engine.root = engine.ginSvr.Group("/")

	// Add custom 404/405 output
	engine.ginSvr.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "[404] http request route mismatch, method: "+c.Request.Method+", path: "+c.Request.URL.Path)
	})
	engine.ginSvr.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "[405] http request method not allowed, method: "+c.Request.Method+", path: "+c.Request.URL.Path)
	})

	// Add health check
	engine.root.GET(com.HealthCheckURLPath, func(c *gin.Context) {
		c.String(http.StatusOK, "ok!")
	})

	// Add swagger
	if engine.opts.swagger {
		engine.root.GET(com.SwaggerURLPath+"/*any", gs.WrapHandler(swag.Handler))
	}

	// Add performance monitoring interface
	if engine.opts.pprof {
		pprofService(engine.root.Group(com.PprofURLPath))
	}

	// Register middleware
	engine.ginSvr.Use(mid.BodyBuffer(), mid.Recovery(engine.config.Logger, engine.config.RecoveryLogEventFunc), mid.Cors())

	return &engine
}

func (e *Engine) Run() {
	// Prevent duplicate startup
	e.lock.Lock()
	if e.running {
		e.lock.Unlock()
		return
	}

	// Register all middleware
	e.registerAllMiddlewares()

	// Register all services
	e.registerAllServices()

	// Register necessary components
	e.ginSvr.Use(mid.AccessLogger(e.config.Logger, e.config.AccessLogEventFunc, e.opts.recReqBody))

	// Initialize http server
	e.httpSvr = &http.Server{
		Addr:              e.endpoint,
		Handler:           e.ginSvr,
		ReadTimeout:       time.Duration(e.config.HttpReadTimeout) * time.Millisecond,
		ReadHeaderTimeout: time.Duration(e.config.HttpReadHeaderTimeout) * time.Millisecond,
		WriteTimeout:      time.Duration(e.config.HttpWriteTimeout) * time.Millisecond,
		IdleTimeout:       0, // Use HttpReadTimeout as the value here
		MaxHeaderBytes:    math.MaxUint32,
		ErrorLog:          zap.NewStdLog(e.config.Logger.Desugar()),
	}

	// Start http server
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.config.Logger.Infow("http server is ready", "address", e.endpoint)
		e.httpSvr.SetKeepAlivesEnabled(true) // Enable keepalive by default, shared by grpc/http
		if err := e.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.config.Logger.Fatalw("failed to start http server", "error", err)
		}
	}()

	// Mark as running
	e.lock.Lock()
	e.running = true
	e.lock.Unlock()

	// Reset shutdown signal
	e.once = sync.Once{}
}

func (e *Engine) Stop() {
	e.once.Do(func() {
		e.lock.Lock()
		e.running = false
		e.lock.Unlock()
		e.cancel()
		e.wg.Wait()
		// Close http server
		if e.httpSvr != nil {
			if err := e.httpSvr.Shutdown(e.ctx); err != nil {
				e.config.Logger.Fatalw("http server forced to shutdown", "address", e.endpoint, "error", err)
			}
		}
		e.config.Logger.Infow("http server is shutdown", "address", e.endpoint)
	})
}

func (e *Engine) IsRunning() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.running
}

func (e *Engine) registerAllServices() {
	e.lock.RLock()
	defer e.lock.RUnlock()
	if !e.running {
		for i := 0; i < len(e.services); i++ {
			e.services[i].RegisterGroup(e.root)
		}
	}
}

func (e *Engine) registerAllMiddlewares() {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		for i := 0; i < len(e.handlers); i++ {
			e.ginSvr.Use(e.handlers[i])
		}
	}
}

func (e *Engine) RegisterService(service Service) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.services = append(e.services, service)
	}
}

func (e *Engine) RegisterMiddleware(handler gin.HandlerFunc) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.handlers = append(e.handlers, handler)
	}
}
