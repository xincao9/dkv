package kv

import (
	"dkv/cache"
	"dkv/compress"
	"dkv/logger"
	"dkv/metrics"
	"dkv/store"
	"dkv/store/appendfile"
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
			response(c, http.StatusBadRequest, "参数错误")
			metrics.GetCount.WithLabelValues("failure", "bad_request").Inc()
			return
		}
		val := cache.Get([]byte(key))
		if val != nil {
			val = compress.Decode(val)
			responseKV(c, http.StatusOK, "成功", &KV{K: key, V: string(val)})
			metrics.GetCount.WithLabelValues("success", "memory").Inc()
			return
		}
		val, err := store.D.Get([]byte(key))
		if err == nil {
			cache.Set([]byte(key), val)
			val = compress.Decode(val)
			responseKV(c, http.StatusOK, "成功", &KV{K: key, V: string(val)})
			metrics.GetCount.WithLabelValues("success", "disk").Inc()
			return
		}
		if err == appendfile.KeyNotFound {
			responseKV(c, http.StatusOK, "没有找到", &KV{K: key, V: ""})
			metrics.GetCount.WithLabelValues("failure", "not_found").Inc()
			return
		}
		logger.D.Errorf("method:get path:/kv/%s err=%s\n", key, err)
		response(c, http.StatusInternalServerError, "服务端错误")
		metrics.GetCount.WithLabelValues("failure", "server_error").Inc()
	})
	engine.PUT("/kv", func(c *gin.Context) {
		var kv KV
		if err := c.ShouldBindJSON(&kv); err != nil {
			response(c, http.StatusBadRequest, "参数错误")
			metrics.PutCount.WithLabelValues("failure").Inc()
			return
		}
		val := compress.Encode([]byte(kv.V))
		err := store.D.Put([]byte(kv.K), val)
		if err != nil {
			logger.D.Errorf("method:post path:/kv/ body:%v err=%s\n", kv, err)
			response(c, http.StatusInternalServerError, "服务端错误")
			metrics.PutCount.WithLabelValues("failure").Inc()
			return
		}
		cache.Del([]byte(kv.K))
		response(c, http.StatusOK, "成功")
		metrics.PutCount.WithLabelValues("success").Inc()
	})
	engine.DELETE("/kv/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			response(c, http.StatusBadRequest, "参数错误")
			return
		}
		err := store.D.Delete([]byte(key))
		if err == nil {
			cache.Del([]byte(key))
			response(c, http.StatusOK, "成功")
			return
		}
		if err == appendfile.KeyNotFound {
			response(c, http.StatusOK, "没有找到")
			return
		}
		logger.D.Errorf("method:delete path:/kv/%s err=%s\n", key, err)
		response(c, http.StatusInternalServerError, "服务端错误")
	})
}

func response(c *gin.Context, code int, message string) {
	c.JSON(code,
		gin.H{
			"code":    code,
			"message": message,
		})
}

func responseKV(c *gin.Context, code int, message string, kv *KV) {
	c.JSON(code,
		gin.H{
			"code":    code,
			"message": message,
			"kv":      kv,
		})

}
