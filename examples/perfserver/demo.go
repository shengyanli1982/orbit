package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/shengyanli1982/orbit"
	"github.com/shengyanli1982/orbit/utils/log"
)

type benchService struct{}

func (s *benchService) RegisterGroup(g *gin.RouterGroup) {
	g.GET("/bench", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	g.GET("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
		})
	})
}

func noopLogEvent(_ *logr.Logger, _ *log.LogEvent) {}

func main() {
	conf := orbit.NewConfig().
		WithAddress("127.0.0.1").
		WithPort(18080).
		WithRelease().
		WithAccessLogEventFunc(noopLogEvent).
		WithRecoveryLogEventFunc(noopLogEvent)

	opts := orbit.NewOptions().
		EnablePProf().
		EnableHealthCheck()

	engine := orbit.NewEngine(conf, opts)
	engine.RegisterService(&benchService{})
	engine.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	engine.Stop()
}
