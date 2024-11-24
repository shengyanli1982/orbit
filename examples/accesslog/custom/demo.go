package main

import (
	"net/http"
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
	// 在 "/demo" 路径上注册一个 GET 方法的处理函数
	// Register a GET method handler function on the "/demo" path
	g.GET("/demo", func(c *gin.Context) {
		// 返回 HTTP 状态码 200 和 "demo" 字符串
		// Return HTTP status code 200 and the string "demo"
		c.String(http.StatusOK, "demo")
	})
}

// customAccessLogger 函数定义了一个自定义的访问日志记录器
// The customAccessLogger function defines a custom access logger
func customAccessLogger(logger *logr.Logger, event *log.LogEvent) {
	// 记录访问日志，包括路径和方法
	// Log the access, including the path and method
	logger.Info("access log", "path", event.Path, "method", event.Method)
}

func main() {
	// 创建一个新的 Orbit 配置，并设置访问日志事件函数
	// Create a new Orbit configuration and set the access log event function
	config := orbit.NewConfig().WithAccessLogEventFunc(customAccessLogger)

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
