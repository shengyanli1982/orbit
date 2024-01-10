package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
	ocom "github.com/shengyanli1982/orbit/common"
)

type service struct{}

func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// Register a custom router group.
	g = g.Group("/demo")

	// /demo
	g.GET(ocom.EmptyURLPath, func(c *gin.Context) {
		c.String(http.StatusOK, "demo")
	})

	// /demo/test
	g.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})
}

func main() {
	// Create a new orbit configuration.
	config := orbit.NewConfig()

	// Create a new orbit feature options.
	opts := orbit.NewOptions().EnableMetric()

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
