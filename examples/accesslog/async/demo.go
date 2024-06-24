package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shengyanli1982/law"
	"github.com/shengyanli1982/orbit"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	// 使用 os.Stdout 和配置创建一个新的 WriteAsyncer 实例
	// Create a new WriteAsyncer instance using os.Stdout and the configuration
	w := law.NewWriteAsyncer(os.Stdout, nil)

	// 使用 defer 语句确保在 main 函数退出时停止 WriteAsyncer
	// Use a defer statement to ensure that WriteAsyncer is stopped when the main function exits
	defer w.Stop()

	// 创建一个 zapcore.EncoderConfig 实例，用于配置 zap 的编码器
	// Create a zapcore.EncoderConfig instance to configure the encoder of zap
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",                         // 消息的键名
		LevelKey:       "level",                       // 级别的键名
		NameKey:        "logger",                      // 记录器名的键名
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 级别的编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // 时间的编码器
		EncodeDuration: zapcore.StringDurationEncoder, // 持续时间的编码器
	}

	// 使用 WriteAsyncer 创建一个 zapcore.WriteSyncer 实例
	// Create a zapcore.WriteSyncer instance using WriteAsyncer
	zapAsyncWriter := zapcore.AddSync(w)

	// 使用编码器配置和 WriteSyncer 创建一个 zapcore.Core 实例
	// Create a zapcore.Core instance using the encoder configuration and WriteSyncer
	zapCore := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), zapAsyncWriter, zapcore.DebugLevel)

	// 使用 Core 创建一个 zap.Logger 实例
	// Create a zap.Logger instance using Core
	zapLogger := zap.New(zapCore)

	// 创建一个新的 Orbit 配置，并设置访问日志事件函数
	// Create a new Orbit configuration and set the access log event function
	config := orbit.NewConfig().WithLogger(zapLogger)

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
