package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
	"github.com/shengyanli1982/orbit/utils/httptool"
)

type service struct{}

func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// /demo
	g.POST("/demo", func(c *gin.Context) {
		// Repeat the read request body content 20 times.
		for i := 0; i < 20; i++ {
			if body, err := httptool.GenerateRequestBody(c); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
			} else {
				c.String(http.StatusOK, fmt.Sprintf(">> %d, %s\n", i, string(body)))
			}
		}
	})
}

func main() {
	// Create a new orbit configuration.
	config := orbit.NewConfig()

	// Create a new orbit feature options.
	opts := orbit.NewOptions()

	// Create a new orbit engine.
	engine := orbit.NewEngine(config, opts)

	// Register a custom router group.
	engine.RegisterService(&service{})

	// Start the engine.
	engine.Run()

	// Simulate a request.
	resp, _ := http.Post("http://localhost:8080/demo", "text/plain", io.Reader(bytes.NewBuffer([]byte("demo"))))
	defer resp.Body.Close()

	// Print the response body.
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	fmt.Println(buf.String())

	// Wait for 30 seconds.
	time.Sleep(30 * time.Second)

	// Stop the engine.
	engine.Stop()
}
