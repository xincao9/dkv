package main

import (
	"dkv/store"
	"dkv/store/appendfile"
	"dkv/store/meta"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type KV struct {
	K string `json:"k"`
	V string `json:"v"`
}

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/dkv/")
	viper.AddConfigPath("$HOME/.dkv")
	viper.AddConfigPath(".")
	viper.SetDefault("data.dir", meta.DefaultDir)
	viper.SetDefault("server.port", ":8080")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func main() {
	store, err := store.New(viper.GetString("data.dir"))
	defer store.Close()
	if err != nil {
		log.Fatalln(err)
	}
	gin.SetMode(viper.GetString("server.mode"))
	r := gin.Default()
	r.GET("/kv/:key", func(c *gin.Context) {
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
		c.JSON(500,
			gin.H{
				"code":    500,
				"message": "服务端错误",
			})
	})
	r.POST("/kv", func(c *gin.Context) {
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
	r.DELETE("/kv/:key", func(c *gin.Context) {
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
		c.JSON(500,
			gin.H{
				"code":    500,
				"message": "服务端错误",
			})
	})
	if err := r.Run(viper.GetString("server.port")); err != nil {
		log.Fatalln(err)
	}
}