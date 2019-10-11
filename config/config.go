package config

import (
	"dkv/store/meta"
	"github.com/spf13/viper"
	"log"
)

var (
	D *viper.Viper
)

func init() {
	// 配置文件设置
	D = viper.New()
	D.SetConfigName("config")
	D.SetConfigType("yaml")
	D.AddConfigPath("/etc/dkv/")
	D.AddConfigPath("$HOME/.dkv")
	D.AddConfigPath(".")
	D.SetDefault("data.dir", meta.DefaultDir)
	D.SetDefault("data.invalidIndex", false)
	D.SetDefault("server.mode", "debug")
	D.SetDefault("server.port", ":9090")
	D.SetDefault("server.sequence", true)
	D.SetDefault("logger.level", "debug")
	err := D.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config : %v\n", err)
	}
}
