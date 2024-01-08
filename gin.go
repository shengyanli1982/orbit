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
	m "github.com/shengyanli1982/orbit/internal/middleware"
	w "github.com/shengyanli1982/orbit/utils/wrapper"
	ginSwaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

var (
	defaultShutdownTimeout = 10 * time.Second
)

func pprofService(g *gin.RouterGroup) {
	// Get
	g.GET("/", w.HandlerFuncToGin(pprof.Index))
	g.GET("/cmdline", w.HandlerFuncToGin(pprof.Cmdline))
	g.GET("/profile", w.HandlerFuncToGin(pprof.Profile))
	g.GET("/symbol", w.HandlerFuncToGin(pprof.Symbol))
	g.GET("/trace", w.HandlerFuncToGin(pprof.Trace))
	g.GET("/allocs", w.HandlerFuncToGin(pprof.Handler("allocs").ServeHTTP))
	g.GET("/block", w.HandlerFuncToGin(pprof.Handler("block").ServeHTTP))
	g.GET("/goroutine", w.HandlerFuncToGin(pprof.Handler("goroutine").ServeHTTP))
	g.GET("/heap", w.HandlerFuncToGin(pprof.Handler("heap").ServeHTTP))
	g.GET("/mutex", w.HandlerFuncToGin(pprof.Handler("mutex").ServeHTTP))
	g.GET("/threadcreate", w.HandlerFuncToGin(pprof.Handler("threadcreate").ServeHTTP))

	// Post
	g.POST("/pprof/symbol", w.HandlerFuncToGin(pprof.Symbol))
}

type Service interface {
	RegisterGroup(g *gin.RouterGroup)
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

func NewEngine(conf *Config, opts *Options) *Engine {
	conf = isConfigValid(conf)
	if conf.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DisableConsoleColor()
	e := Engine{
		running:  false,
		endpoint: fmt.Sprintf("%s:%d", conf.Address, conf.Port),
		config:   conf,
		opts:     opts,
		lock:     sync.RWMutex{},
		wg:       sync.WaitGroup{},
		once:     sync.Once{},
	}
	e.ctx, e.cancel = context.WithTimeout(context.Background(), defaultShutdownTimeout)
	e.ginSvr = gin.New()
	e.root = e.ginSvr.Group("/")

	// 增加自定义 404/405 输出
	e.ginSvr.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "[404] http request route mismatch, method: "+c.Request.Method+", path: "+c.Request.URL.Path)
	})
	e.ginSvr.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "[405] http request method not allowed, method: "+c.Request.Method+", path: "+c.Request.URL.Path)
	})

	// 添加健康检测
	e.root.GET(com.HttpHealthCheckUrlPath, func(c *gin.Context) {
		c.String(http.StatusOK, "ok!")
	})

	// 添加 swagger
	if e.opts.swagger {
		e.root.GET(com.HttpSwaggerUrlPath+"/*any", ginSwagger.WrapHandler(ginSwaggerFiles.Handler))
	}

	// 添加性能监控接口
	if e.opts.pprof {
		pprofService(e.root.Group(com.HttpPprofUrlPath))
	}

	// 注册中间件
	e.ginSvr.Use(m.BodyBuffer(), m.Recovery(e.config.Logger), m.Cors())

	return &e
}

func (e *Engine) Run() {
	// 防止重复启动
	e.lock.Lock()
	if e.running {
		e.lock.Unlock()
		return
	}

	// 注册所有中间件
	e.registerAllMiddlewares()

	// 注册所有服务
	e.registerAllServices()

	// 注册必要的组件
	e.ginSvr.Use(m.AccessLogger(e.config.Logger, false))

	// 初始化 http server
	e.httpSvr = &http.Server{
		Addr:              e.endpoint,
		Handler:           e.ginSvr,
		ReadTimeout:       time.Duration(e.config.HttpReadTimeout) * time.Millisecond,
		ReadHeaderTimeout: time.Duration(e.config.HttpReadHeaderTimeout) * time.Millisecond,
		WriteTimeout:      time.Duration(e.config.HttpWriteTimeout) * time.Millisecond,
		IdleTimeout:       0, // 使用 HttpReadTimeout 做为这里的值
		MaxHeaderBytes:    math.MaxUint32,
		ErrorLog:          zap.NewStdLog(e.config.Logger.Desugar()),
	}

	// 启动 http server
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.config.Logger.Infow("http server is ready", "address", e.endpoint)
		e.httpSvr.SetKeepAlivesEnabled(true) // 默认开启 keepalive， grpc/http 共享
		if err := e.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.config.Logger.Fatalw("failed to start http server", "error", err)
		}
	}()

	// 标记为已启动
	e.lock.Lock()
	e.running = true
	e.lock.Unlock()

	// 重置关闭信号
	e.once = sync.Once{}
}

func (e *Engine) Stop() {
	e.once.Do(func() {
		e.lock.Lock()
		e.running = false
		e.lock.Unlock()
		e.cancel()
		e.wg.Wait()
		// 关闭 http server
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

func (e *Engine) RegisterService(svc Service) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.services = append(e.services, svc)
	}
}

func (e *Engine) RegisterMiddleware(handler gin.HandlerFunc) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.running {
		e.handlers = append(e.handlers, handler)
	}
}
