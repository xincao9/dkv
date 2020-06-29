package kv

import (
	"dkv/component/cache"
	"dkv/component/compress"
	"dkv/component/constant"
	"dkv/component/logger"
	"dkv/component/metrics"
	"dkv/store"
	"net/http"

	"github.com/gin-gonic/gin"
)

type KV struct {
	K string `json:"k"`
	V string `json:"v"`
}

func Route(engine *gin.Engine) {
	engine.GET("/kv/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			response(c, http.StatusBadRequest, constant.InvalidArgument)
			metrics.GetCount.WithLabelValues("failure", "bad_request").Inc()
			return
		}
		val := cache.C.Get([]byte(key))
		if val != nil {
			val = compress.C.Decode(val)
			responseKV(c, http.StatusOK, constant.Ok, &KV{K: key, V: string(val)})
			metrics.GetCount.WithLabelValues("success", "memory").Inc()
			return
		}
		val, err := store.S.Get([]byte(key))
		if err == nil {
			cache.C.Set([]byte(key), val)
			val = compress.C.Decode(val)
			responseKV(c, http.StatusOK, constant.Ok, &KV{K: key, V: string(val)})
			metrics.GetCount.WithLabelValues("success", "disk").Inc()
			return
		}
		if err == constant.KeyNotFoundError {
			responseKV(c, http.StatusNotFound, constant.KeyNotFound, &KV{K: key, V: ""})
			metrics.GetCount.WithLabelValues("failure", "not_found").Inc()
			return
		}
		logger.L.Errorf("method:get path:/kv/%s err=%s\n", key, err)
		response(c, http.StatusInternalServerError, constant.InternalError)
		metrics.GetCount.WithLabelValues("failure", "server_error").Inc()
	})
	engine.PUT("/kv", func(c *gin.Context) {
		var kv KV
		if err := c.ShouldBindJSON(&kv); err != nil {
			response(c, http.StatusBadRequest, constant.InvalidArgument)
			metrics.PutCount.WithLabelValues("failure").Inc()
			return
		}
		val := compress.C.Encode([]byte(kv.V))
		err := store.S.Put([]byte(kv.K), val)
		if err != nil {
			logger.L.Errorf("method:post path:/kv/ body:%v err=%s\n", kv, err)
			response(c, http.StatusInternalServerError, constant.InternalError)
			metrics.PutCount.WithLabelValues("failure").Inc()
			return
		}
		cache.C.Del([]byte(kv.K))
		response(c, http.StatusOK, constant.Ok)
		metrics.PutCount.WithLabelValues("success").Inc()
	})
	engine.DELETE("/kv/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			response(c, http.StatusBadRequest, constant.InvalidArgument)
			return
		}
		err := store.S.Delete([]byte(key))
		if err == nil {
			cache.C.Del([]byte(key))
			response(c, http.StatusOK, constant.Ok)
			return
		}
		if err == constant.KeyNotFoundError {
			response(c, http.StatusNotFound, constant.KeyNotFound)
			return
		}
		logger.L.Errorf("method:delete path:/kv/%s err=%s\n", key, err)
		response(c, http.StatusInternalServerError, constant.InternalError)
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
