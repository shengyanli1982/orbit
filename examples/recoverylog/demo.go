package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/shengyanli1982/orbit"
	"github.com/shengyanli1982/orbit/utils/log"
)

// 定义 service 结构体
// Define the service struct
type service struct{}

// RegisterGroup 方法将路由组注册到 service
// The RegisterGroup method registers a router group to the service
func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// 在 "/demo" 路径上注册一个 GET 方法的处理函数，该函数会触发 panic
	// Register a GET method handler function on the "/demo" path, this function will trigger a panic
	g.GET("/demo", func(c *gin.Context) {
		panic("demo")
	})
}

// customRecoveryLogger 函数定义了一个自定义的恢复日志记录器
// The customRecoveryLogger function defines a custom recovery logger
func customRecoveryLogger(logger *logr.Logger, event *log.LogEvent) {
	// 记录恢复日志，包括路径、方法、错误和错误堆栈
	// Log the recovery, including the path, method, error, and error stack
	logger.Info("recovery log", "path", event.Path, "method", event.Method, "error", event.Error, "errorStack", event.ErrorStack)
}

func main() {
	// 创建一个新的 Orbit 配置，并设置恢复日志事件函数
	// Create a new Orbit configuration and set the recovery log event function
	config := orbit.NewConfig().WithRecoveryLogEventFunc(customRecoveryLogger)

	// 创建一个新的 Orbit 功能选项
	// Create a new Orbit feature options
	opts := orbit.NewOptions()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(config, opts)

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
