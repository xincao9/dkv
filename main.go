package main

import (
	"dkv/kv"
	"dkv/oss"
	"dkv/store"
	"dkv/store/config"
	"dkv/store/ginpprof"
	"dkv/store/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"log"
)

func main() {
	// 启动存储引擎
	defer store.D.Close()

	// 启动http服务
	gin.SetMode(config.D.GetString("server.mode"))
	engine := gin.New()
	engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: logger.D.WriterLevel(logrus.DebugLevel)}), gin.Recovery())
	kv.Route(engine)                              // 注册KV服务接口
	oss.Route(engine)                             // 注册OSS服务接口
	engine.GET("/metrics", func(c *gin.Context) { // 注册普罗米修斯接口
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})
	ginpprof.Wrap(engine) // 注册pprof接口
	if err := engine.Run(config.D.GetString("server.port")); err != nil {
		log.Fatalf("Fatal error gin: %v\n", err)
	}
}
