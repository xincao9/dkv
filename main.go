package main

import (
	"dkv/cache"
	"dkv/config"
	"dkv/kv"
	"dkv/logger"
	"dkv/metrics"
	"dkv/oss"
	"dkv/pprof"
	"dkv/store"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
)

func main() {
	// 启动存储引擎
	defer store.D.Close()
	defer cache.Close()

	// 启动http服务
	gin.SetMode(config.D.GetString("server.mode"))
	engine := gin.New()
	engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: logger.D.WriterLevel(logrus.DebugLevel)}), gin.RecoveryWithWriter(logger.D.WriterLevel(logrus.ErrorLevel)))
	kv.Route(engine)      // 注册KV服务接口
	oss.Route(engine)     // 注册OSS服务接口
	metrics.Route(engine) // 注册普罗米修斯接口
	pprof.Wrap(engine)    // 注册pprof接口
	config.Route(engine)  // 配置服务接口
	if err := engine.Run(config.D.GetString("server.port")); err != nil {
		log.Fatalf("Fatal error gin: %v\n", err)
	}
}
