package kv

import (
	"dkv/store"
	"dkv/store/appendfile"
	"dkv/store/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

type KV struct {
	K string `json:"k"`
	V string `json:"v"`
}

func Route(engine *gin.Engine) {
	engine.GET("/kv/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "参数错误",
			})
			return
		}
		val, err := store.D.Get([]byte(key))
		if err == nil {
			c.JSON(http.StatusOK,
				gin.H{
					"code":    http.StatusOK,
					"message": "成功",
					"kv": &KV{
						K: key,
						V: string(val),
					},
				})
			return
		}
		if err == appendfile.KeyNotFound {
			c.JSON(http.StatusOK,
				gin.H{
					"code":    http.StatusOK,
					"message": "没有找到",
					"kv": &KV{
						K: key,
						V: "",
					},
				})
			return
		}
		logger.D.Errorf("method:get path:/kv/%s err=%s\n", key, err)
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code":    http.StatusInternalServerError,
				"message": "服务端错误",
			})
	})
	engine.PUT("/kv", func(c *gin.Context) {
		var kv KV
		if err := c.ShouldBindJSON(&kv); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "参数错误",
			})
			return
		}
		err := store.D.Put([]byte(kv.K), []byte(kv.V))
		if err != nil {
			logger.D.Errorf("method:post path:/kv/ body:%v err=%s\n", kv, err)
			c.JSON(http.StatusInternalServerError,
				gin.H{
					"code":    http.StatusInternalServerError,
					"message": "服务端错误",
				})
			return
		}
		c.JSON(http.StatusOK,
			gin.H{
				"code":    http.StatusOK,
				"message": "成功",
			})
	})
	engine.DELETE("/kv/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "参数错误",
			})
			return
		}
		err := store.D.Delete([]byte(key))
		if err == nil {
			c.JSON(http.StatusOK,
				gin.H{
					"code":    http.StatusOK,
					"message": "成功",
				})
			return
		}
		if err == appendfile.KeyNotFound {
			c.JSON(http.StatusOK,
				gin.H{
					"code":    http.StatusOK,
					"message": "没有找到",
				})
			return
		}
		logger.D.Errorf("method:delete path:/kv/%s err=%s\n", key, err)
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code":    http.StatusInternalServerError,
				"message": "服务端错误",
			})
	})
}
