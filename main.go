package main

import (
	"dkv/store"
	"dkv/store/appendfile"
	"dkv/store/ginpprof"
	"dkv/store/meta"
	logrus "github.com/sirupsen/logrus"
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

type KV struct {
	K string `json:"k"`
	V string `json:"v"`
}

var (
	logger *logrus.Logger
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/dkv/")
	viper.AddConfigPath("$HOME/.dkv")
	viper.AddConfigPath(".")
	viper.SetDefault("data.dir", meta.DefaultDir)
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("logger.level", "debug")
	viper.SetDefault("server.sequence", false)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %v\n", err)
	}
	logger = logrus.New()
	level, err := logrus.ParseLevel(viper.GetString("logger.level"))
	if err != nil {
		log.Fatalf("Fatal error config file logger.level: %v\n", err)
	}
	fn := filepath.Join(viper.GetString("data.dir"), "server.log")
	logger.Out = &lumberjack.Logger{
		Filename:   fn,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
	logger.SetLevel(level)
	logger.Formatter = &logrus.JSONFormatter{}
	log.SetOutput(logger.WriterLevel(logrus.InfoLevel))
}

func main() {
	store, err := store.New(viper.GetString("data.dir"))
	defer store.Close()
	if err != nil {
		log.Fatalf("Fatal error store: %v\n", err)
	}
	gin.SetMode(viper.GetString("server.mode"))
	engine := gin.New()
	engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: logger.WriterLevel(logrus.DebugLevel)}), gin.Recovery())
	engine.GET("/kv/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.JSON(400, gin.H{
				"code":    400,
				"message": "参数错误",
			})
			return
		}
		val, err := store.Get([]byte(key))
		if err == nil {
			c.JSON(200,
				gin.H{
					"code":    200,
					"message": "成功",
					"kv": &KV{
						K: key,
						V: string(val),
					},
				})
			return
		}
		if err == appendfile.KeyNotFound {
			c.JSON(200,
				gin.H{
					"code":    200,
					"message": "没有找到",
					"kv": &KV{
						K: key,
						V: "",
					},
				})
			return
		}
		logger.Errorf("method:get path:/kv/%s err=%s\n", key, err)
		c.JSON(500,
			gin.H{
				"code":    500,
				"message": "服务端错误",
			})
	})
	engine.POST("/kv", func(c *gin.Context) {
		var kv KV
		if err := c.ShouldBindJSON(&kv); err != nil {
			c.JSON(400, gin.H{
				"code":    400,
				"message": "参数错误",
			})
			return
		}
		err := store.Put([]byte(kv.K), []byte(kv.V))
		if err != nil {
			logger.Errorf("method:post path:/kv/ body:%v err=%s\n", kv, err)
			c.JSON(500,
				gin.H{
					"code":    500,
					"message": "服务端错误",
				})
			return
		}
		c.JSON(200,
			gin.H{
				"code":    200,
				"message": "成功",
			})
	})
	engine.DELETE("/kv/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.JSON(400, gin.H{
				"code":    400,
				"message": "参数错误",
			})
			return
		}
		err := store.Delete([]byte(key))
		if err == nil {
			c.JSON(200,
				gin.H{
					"code":    200,
					"message": "成功",
				})
			return
		}
		if err == appendfile.KeyNotFound {
			c.JSON(200,
				gin.H{
					"code":    200,
					"message": "没有找到",
				})
			return
		}
		logger.Errorf("method:delete path:/kv/%s err=%s\n", key, err)
		c.JSON(500,
			gin.H{
				"code":    500,
				"message": "服务端错误",
			})
	})
	engine.GET("/metrics", func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})
	ginpprof.Wrap(engine)
	if err := engine.Run(viper.GetString("server.port")); err != nil {
		log.Fatalf("Fatal error gin: %v\n", err)
	}
}
