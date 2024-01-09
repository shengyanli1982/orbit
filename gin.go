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

// defaultShutdownTimeout is the default timeout for graceful shutdown.
var defaultShutdownTimeout = 10 * time.Second

// pprofService registers the pprof handlers to the given router group.
func pprofService(g *gin.RouterGroup) {
	// Get
	g.GET("/", wrap.WrapHandlerFuncToGin(pprof.Index))                                         // Get the pprof index page
	g.GET("/cmdline", wrap.WrapHandlerFuncToGin(pprof.Cmdline))                                // Get the command line arguments
	g.GET("/profile", wrap.WrapHandlerFuncToGin(pprof.Profile))                                // Get the profiling goroutine stack traces
	g.GET("/symbol", wrap.WrapHandlerFuncToGin(pprof.Symbol))                                  // Get the symbol table
	g.GET("/trace", wrap.WrapHandlerFuncToGin(pprof.Trace))                                    // Get the execution trace
	g.GET("/allocs", wrap.WrapHandlerFuncToGin(pprof.Handler("allocs").ServeHTTP))             // Get the heap allocations
	g.GET("/block", wrap.WrapHandlerFuncToGin(pprof.Handler("block").ServeHTTP))               // Get the goroutine blocking profile
	g.GET("/goroutine", wrap.WrapHandlerFuncToGin(pprof.Handler("goroutine").ServeHTTP))       // Get the goroutine profile
	g.GET("/heap", wrap.WrapHandlerFuncToGin(pprof.Handler("heap").ServeHTTP))                 // Get the heap profile
	g.GET("/mutex", wrap.WrapHandlerFuncToGin(pprof.Handler("mutex").ServeHTTP))               // Get the mutex profile
	g.GET("/threadcreate", wrap.WrapHandlerFuncToGin(pprof.Handler("threadcreate").ServeHTTP)) // Get the thread creation profile

	// Post
	g.POST("/pprof/symbol", wrap.WrapHandlerFuncToGin(pprof.Symbol)) // Get the symbol table
}

// Service is the interface that represents a service.
type Service interface {
	RegisterGroup(routerGroup *gin.RouterGroup) // Register the service to the router group
}

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
}

// NewEngine creates a new instance of the Engine.
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
		c.String(http.StatusOK, com.RequestOK)
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

// Run starts the Orbit engine.
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

// Stop stops the Orbit engine.
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

// IsRunning returns true if the Orbit engine is running.
func (e *Engine) IsRunning() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.running
}

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

// registerAllMiddlewares registers all middlewares to the engine.
func (e *Engine) registerAllMiddlewares() {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		for i := 0; i < len(e.handlers); i++ {
			e.ginSvr.Use(e.handlers[i])
		}
	}
}

// RegisterService registers a service to the Orbit engine.
func (e *Engine) RegisterService(service Service) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.services = append(e.services, service)
	}
}

// RegisterMiddleware registers a middleware to the Orbit engine.
func (e *Engine) RegisterMiddleware(handler gin.HandlerFunc) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.handlers = append(e.handlers, handler)
	}
}
