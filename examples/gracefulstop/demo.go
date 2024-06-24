package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/gs"
	"github.com/shengyanli1982/orbit"
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

func main() {
	// 创建一个新的 TerminateSignal 实例
	// Create a new TerminateSignal instance
	sig := gs.NewTerminateSignal()

	// 创建一个新的 Orbit 功能选项
	// Create a new Orbit feature options
	opts := orbit.NewOptions()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(nil, opts)

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 注册需要在终止信号发生时执行的处理函数
	// Register the handle functions to be executed when the termination signal occurs
	sig.RegisterCancelHandles(engine.Stop)

	// 等待所有的异步关闭信号
	// Wait for all asynchronous shutdown signals
	gs.WaitForAsync(sig)
}
