package orbit

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	com "github.com/shengyanli1982/orbit/common"
	wrap "github.com/shengyanli1982/orbit/utils/wrapper"
	swag "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

// pprofService registers the pprof handlers to the given router group.
func pprofService(group *gin.RouterGroup) {
	// Get
	group.GET(com.EmptyURLPath, wrap.WrapHandlerFuncToGin(pprof.Index))                            // Get the pprof index page
	group.GET("/cmdline", wrap.WrapHandlerFuncToGin(pprof.Cmdline))                                // Get the command line arguments
	group.GET("/profile", wrap.WrapHandlerFuncToGin(pprof.Profile))                                // Get the profiling goroutine stack traces
	group.GET("/symbol", wrap.WrapHandlerFuncToGin(pprof.Symbol))                                  // Get the symbol table
	group.GET("/trace", wrap.WrapHandlerFuncToGin(pprof.Trace))                                    // Get the execution trace
	group.GET("/allocs", wrap.WrapHandlerFuncToGin(pprof.Handler("allocs").ServeHTTP))             // Get the heap allocations
	group.GET("/block", wrap.WrapHandlerFuncToGin(pprof.Handler("block").ServeHTTP))               // Get the goroutine blocking profile
	group.GET("/goroutine", wrap.WrapHandlerFuncToGin(pprof.Handler("goroutine").ServeHTTP))       // Get the goroutine profile
	group.GET("/heap", wrap.WrapHandlerFuncToGin(pprof.Handler("heap").ServeHTTP))                 // Get the heap profile
	group.GET("/mutex", wrap.WrapHandlerFuncToGin(pprof.Handler("mutex").ServeHTTP))               // Get the mutex profile
	group.GET("/threadcreate", wrap.WrapHandlerFuncToGin(pprof.Handler("threadcreate").ServeHTTP)) // Get the thread creation profile

	// Post
	group.POST("/pprof/symbol", wrap.WrapHandlerFuncToGin(pprof.Symbol)) // Get the symbol table
}

// metricService registers the prometheus metrics handlers to the given router group.
func metricService(group *gin.RouterGroup) {
	group.GET(com.EmptyURLPath, wrap.WrapHandlerToGin(promhttp.Handler()))
}

// swaggerService registers the swagger handlers to the given router group.
func swaggerService(group *gin.RouterGroup) {
	group.GET("/*any", gs.WrapHandler(swag.Handler))
}

// healthcheckService registers the healthcheck handlers to the given router group.
func healthcheckService(group *gin.RouterGroup) {
	group.GET(com.EmptyURLPath, func(c *gin.Context) {
		c.String(http.StatusOK, com.RequestOK)
	})
}
