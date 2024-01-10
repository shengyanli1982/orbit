package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
	"github.com/shengyanli1982/orbit/utils/log"
	"go.uber.org/zap"
)

type service struct{}

func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// /demo
	g.GET("/demo", func(c *gin.Context) {
		panic("demo")
	})
}

func customRecoveryLogger(logger *zap.SugaredLogger, event *log.LogEvent) {
	logger.Infow("recovery log", "path", event.Path, "method", event.Method, "error", event.Error, "errorStack", event.ErrorStack)
}

func main() {
	// Create a new orbit configuration.
	config := orbit.NewConfig().WithRecoveryLogEventFunc(customRecoveryLogger)

	// Create a new orbit feature options.
	opts := orbit.NewOptions()

	// Create a new orbit engine.
	engine := orbit.NewEngine(config, opts)

	// Register a custom router group.
	engine.RegisterService(&service{})

	// Start the engine.
	engine.Run()

	// Wait for 30 seconds.
	time.Sleep(30 * time.Second)

	// Stop the engine.
	engine.Stop()
}
