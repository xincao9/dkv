package config

import (
    "flag"
    "github.com/gin-gonic/gin"
    "github.com/spf13/viper"
    "log"
    "net/http"
    "strings"
)

var (
	C *viper.Viper
)

func init() {
	c := flag.String("conf", "config.yaml", "configure file")
	flag.Parse()
	// 配置文件设置
	C = viper.New()
	for _, t := range []string{"yaml", "yml"} {
		if strings.HasSuffix(*c, t) {
			i := strings.LastIndex(*c, t)
			*c = string([]byte(*c)[:i-1])
		}
	}
    C.SetConfigName(*c)
	C.SetConfigType("yaml")
	C.AddConfigPath("/tmp/dkv/conf")
	C.SetDefault("data.dir", "/tmp/dkv/data")
	C.SetDefault("data.invalidIndex", false)
	C.SetDefault("data.compress", false)
	C.SetDefault("data.cache", true)
	C.SetDefault("server.mode", "debug")
	C.SetDefault("server.port", 9090)
	C.SetDefault("server.redis.port", 6380)
	C.SetDefault("server.sequence", true)
	C.SetDefault("logger.level", "debug")
	C.SetDefault("ms.role", 0)                  // 0 默认模式，1 主节点 2 从节点
	C.SetDefault("ms.m.port", 7380)             // 主节点监听端口
	C.SetDefault("ms.s.addr", "localhost:7380") // 从同步的主节点地址
	err := C.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config : %v\n", err)
	}
}

func Route(engine *gin.Engine) {
	engine.GET("/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, C.AllSettings())
	})
}
