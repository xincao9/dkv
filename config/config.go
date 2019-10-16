package config

import (
	"dkv/store/meta"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
)

var (
	D *viper.Viper
)

func init() {
	// 配置文件设置
	D = viper.New()
	if os.Getenv("ENV") == "prod" {
		D.SetConfigName("config-prod")
	} else {
		D.SetConfigName("config")
	}
	D.SetConfigType("yaml")
	D.AddConfigPath("/etc/dkv/")
	D.AddConfigPath("$HOME/.dkv")
	D.AddConfigPath(".")
	D.AddConfigPath("/usr/local/dkv/")
	D.SetDefault("data.dir", meta.DefaultDir)
	D.SetDefault("data.invalidIndex", false)
	D.SetDefault("data.compress", false)
	D.SetDefault("data.cache", true)
	D.SetDefault("server.mode", "debug")
	D.SetDefault("server.port", ":9090")
	D.SetDefault("server.sequence", true)
	D.SetDefault("logger.level", "debug")
	err := D.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config : %v\n", err)
	}
}

func Route(engine *gin.Engine) {
	engine.GET("/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, D.AllSettings())
	})
}
