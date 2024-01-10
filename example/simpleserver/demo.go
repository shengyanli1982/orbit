package main

import (
	"time"

	"github.com/shengyanli1982/orbit"
)

func main() {
	// Create a new orbit configuration.
	config := orbit.NewConfig()

	// Create a new orbit feature options.
	opts := orbit.NewOptions()

	// Create a new orbit engine.
	engine := orbit.NewEngine(config, opts)

	// Start the engine.
	engine.Run()

	// Wait for 30 seconds.
	time.Sleep(30 * time.Second)

	// Stop the engine.
	engine.Stop()
}
