package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
	"github.com/shengyanli1982/orbit/utils/httptool"
)

func customMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Get the response body buffer from the context
		for i := 0; i < 20; i++ {
			body, _ := httptool.GenerateResponseBody(c)
			fmt.Printf("# %d, %s\n", i, string(body))
		}
	}
}

type service struct{}

func (s *service) RegisterGroup(g *gin.RouterGroup) {
	g.GET("/demo", func(c *gin.Context) {
		c.String(http.StatusOK, "demo")
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

	// Simulate a request.
	resp, _ := http.Get("http://localhost:8080/demo")
	defer resp.Body.Close()

	// Wait for 30 seconds.
	time.Sleep(30 * time.Second)

	// Stop the engine.
	engine.Stop()
}
