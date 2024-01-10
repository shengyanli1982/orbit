package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
)

func customMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("xxxx")
		c.Next()
	}
}

type service struct{}

func (s *service) RegisterGroup(g *gin.RouterGroup) {
	g.GET("/demo", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}

func main() {
	// Create a new orbit configuration.
	config := orbit.NewConfig()

	// Create a new orbit feature options.
	opts := orbit.NewOptions().EnableMetric()

	// Create a new orbit engine.
	engine := orbit.NewEngine(config, opts)

	// Register a custom middleware.
	engine.RegisterMiddleware(customMiddleware())

	// Register a custom router group.
	engine.RegisterService(&service{})

	// Start the engine.
	engine.Run()

	// Wait for 10 seconds.
	time.Sleep(10 * time.Second)

	// Stop the engine.
	engine.Stop()
}
