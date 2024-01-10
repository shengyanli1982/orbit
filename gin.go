package orbit

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	com "github.com/shengyanli1982/orbit/common"
	mtc "github.com/shengyanli1982/orbit/internal/metric"
	mid "github.com/shengyanli1982/orbit/internal/middleware"

	"go.uber.org/zap"
)

// defaultShutdownTimeout is the default timeout for graceful shutdown.
var defaultShutdownTimeout = 10 * time.Second

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
	metric   *mtc.ServerMetrics // Prometheus metric
}

// NewEngine creates a new instance of the Engine.
func NewEngine(config *Config, options *Options) *Engine {
	// Validate config
	config = isConfigValid(config)

	// Validate options
	options = isOptionsValid(options)

	// Check running mode
	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Disable console color
	gin.DisableConsoleColor()

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

	// Create Gin engine
	engine.ginSvr = gin.New()

	// Create root router group
	engine.root = engine.ginSvr.Group("/")

	// Set Gin engine options
	engine.ginSvr.ForwardedByClientIP = options.forwordByClientIp
	engine.ginSvr.RedirectTrailingSlash = options.trailingSlash
	engine.ginSvr.RedirectFixedPath = options.fixedPath

	// Add custom 404/405 output
	engine.ginSvr.NoRoute(func(context *gin.Context) {
		context.String(http.StatusNotFound, "[404] http request route mismatch, method: "+context.Request.Method+", path: "+context.Request.URL.Path)
	})
	engine.ginSvr.NoMethod(func(context *gin.Context) {
		context.String(http.StatusMethodNotAllowed, "[405] http request method not allowed, method: "+context.Request.Method+", path: "+context.Request.URL.Path)
	})

	// Register middleware
	engine.ginSvr.Use(
		mid.Recovery(engine.config.logger, engine.config.recoveryLogEventFunc), // Recovery from panic
		mid.BodyBuffer(), // Buffer for request/response body
		mid.Cors(),       // Cross-origin resource sharing
	)

	// Add health check
	healthcheckService(engine.root.Group(com.HealthCheckURLPath))

	// Add swagger
	if engine.opts.swagger {
		swaggerService(engine.root.Group(com.SwaggerURLPath))
	}

	// Add performance monitoring interface
	if engine.opts.pprof {
		pprofService(engine.root.Group(com.PprofURLPath))
	}

	// Add Prometheus metric interface
	if engine.opts.metric {
		engine.metric.Register()
		metricService(engine.root.Group(com.PromMetricURLPath))
		engine.ginSvr.Use(engine.metric.HandlerFunc(engine.config.logger))
	}

	return &engine
}

// Run starts the Orbit engine.
func (e *Engine) Run() {
	// Prevent duplicate startup
	if e.IsRunning() {
		return
	}

	// Register all middlewares
	e.registerAllMiddlewares()

	// Register all services
	e.registerAllServices()

	// Register necessary components
	e.ginSvr.Use(mid.AccessLogger(e.config.logger, e.config.accessLogEventFunc, e.opts.recReqBody))

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

	// Start http server
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		// Enable keepalive by default, shared by grpc/http
		e.httpSvr.SetKeepAlivesEnabled(true)
		e.config.logger.Infow("http server is ready", "address", e.endpoint)
		if err := e.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.config.logger.Fatalw("failed to start http server", "error", err)
		}
	}()

	// Mark as running
	e.setRuningStatus(true)

	// Reset shutdown signal
	e.once = sync.Once{}
}

// Stop stops the Orbit engine.
func (e *Engine) Stop() {
	e.once.Do(func() {
		// Mark as not running
		e.setRuningStatus(false)

		// Signal shutdown
		e.cancel()
		e.wg.Wait()

		// Close http server
		if e.httpSvr != nil {
			if err := e.httpSvr.Shutdown(e.ctx); err != nil {
				e.config.logger.Fatalw("http server forced to shutdown", "address", e.endpoint, "error", err)
			}
		}
		e.config.logger.Infow("http server is shutdown", "address", e.endpoint)

		// Unregister Prometheus metric
		if e.opts.metric {
			e.metric.Unregister()
		}
	})
}

// IsRunning returns true if the Orbit engine is running.
func (e *Engine) IsRunning() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.running
}

func (e *Engine) setRuningStatus(status bool) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	e.running = status
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
		e.ginSvr.Use(e.handlers...)
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
