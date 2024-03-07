package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/orbit"
	"github.com/shengyanli1982/orbit/utils/httptool"
)

// customMiddleware 函数定义了一个自定义的中间件
// The customMiddleware function defines a custom middleware
func customMiddleware() gin.HandlerFunc {
	// 返回一个 Gin 的 HandlerFunc
	// Return a Gin HandlerFunc
	return func(c *gin.Context) {
		// 调用下一个中间件或处理函数
		// Call the next middleware or handler function
		c.Next()

		// 从上下文中获取响应体缓冲区
		// Get the response body buffer from the context
		for i := 0; i < 20; i++ {
			// 生成响应体
			// Generate the response body
			body, _ := httptool.GenerateResponseBody(c)
			// 打印响应体
			// Print the response body
			fmt.Printf("# %d, %s\n", i, string(body))
		}
	}
}

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
	// 创建一个新的 Orbit 配置
	// Create a new Orbit configuration
	config := orbit.NewConfig()

	// 创建一个新的 Orbit 功能选项，并启用 metric
	// Create a new Orbit feature options and enable metric
	opts := orbit.NewOptions().EnableMetric()

	// 创建一个新的 Orbit 引擎
	// Create a new Orbit engine
	engine := orbit.NewEngine(config, opts)

	// 注册一个自定义的中间件
	// Register a custom middleware
	engine.RegisterMiddleware(customMiddleware())

	// 注册一个自定义的路由组
	// Register a custom router group
	engine.RegisterService(&service{})

	// 启动引擎
	// Start the engine
	engine.Run()

	// 模拟一个请求
	// Simulate a request
	resp, _ := http.Get("http://localhost:8080/demo")
	defer resp.Body.Close()

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
