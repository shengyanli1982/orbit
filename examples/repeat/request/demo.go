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

// 定义 service 结构体
// Define the service struct
type service struct{}

// RegisterGroup 方法将路由组注册到 service
// The RegisterGroup method registers a router group to the service
func (s *service) RegisterGroup(g *gin.RouterGroup) {
	// 在 "/demo" 路径上注册一个 POST 方法的处理函数
	// Register a POST method handler function on the "/demo" path
	g.POST("/demo", func(c *gin.Context) {
		// 重复读取请求体内容 20 次
		// Repeat the read request body content 20 times
		for i := 0; i < 20; i++ {
			// 生成请求体
			// Generate the request body
			if body, err := httptool.GenerateRequestBody(c); err != nil {
				// 如果生成请求体出错，返回 HTTP 状态码 500 和错误信息
				// If there is an error generating the request body, return HTTP status code 500 and the error message
				c.String(http.StatusInternalServerError, err.Error())
			} else {
				// 如果生成请求体成功，返回 HTTP 状态码 200 和请求体内容
				// If the request body is successfully generated, return HTTP status code 200 and the request body content
				c.String(http.StatusOK, fmt.Sprintf(">> %d, %s\n", i, string(body)))
			}
		}
	})
}

func main() {
	// 创建一个新的 Orbit 配置
	// Create a new Orbit configuration
	config := orbit.NewConfig()

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

	// 模拟一个请求
	// Simulate a request
	resp, _ := http.Post("http://localhost:8080/demo", "text/plain", io.Reader(bytes.NewBuffer([]byte("demo"))))
	defer resp.Body.Close()

	// 打印响应体
	// Print the response body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	fmt.Println(buf.String())

	// 等待 30 秒
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// 停止引擎
	// Stop the engine
	engine.Stop()
}
