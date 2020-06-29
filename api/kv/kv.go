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

const (
	InvalidArgument = "invalid argument"
	Ok              = "ok"
	InternalError   = "internal error"
	KeyNotFound     = "key not found"
)

func Route(engine *gin.Engine) {
	engine.GET("/kv/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			response(c, http.StatusBadRequest, InvalidArgument)
			metrics.GetCount.WithLabelValues("failure", "bad_request").Inc()
			return
		}
		val := cache.C.Get([]byte(key))
		if val != nil {
			val = compress.C.Decode(val)
			responseKV(c, http.StatusOK, Ok, &KV{K: key, V: string(val)})
			metrics.GetCount.WithLabelValues("success", "memory").Inc()
			return
		}
		val, err := store.S.Get([]byte(key))
		if err == nil {
			cache.C.Set([]byte(key), val)
			val = compress.C.Decode(val)
			responseKV(c, http.StatusOK, Ok, &KV{K: key, V: string(val)})
			metrics.GetCount.WithLabelValues("success", "disk").Inc()
			return
		}
		if err == constant.KeyNotFound {
			responseKV(c, http.StatusNotFound, KeyNotFound, &KV{K: key, V: ""})
			metrics.GetCount.WithLabelValues("failure", "not_found").Inc()
			return
		}
		logger.L.Errorf("method:get path:/kv/%s err=%s\n", key, err)
		response(c, http.StatusInternalServerError, InternalError)
		metrics.GetCount.WithLabelValues("failure", "server_error").Inc()
	})
	engine.PUT("/kv", func(c *gin.Context) {
		var kv KV
		if err := c.ShouldBindJSON(&kv); err != nil {
			response(c, http.StatusBadRequest, InvalidArgument)
			metrics.PutCount.WithLabelValues("failure").Inc()
			return
		}
		val := compress.C.Encode([]byte(kv.V))
		err := store.S.Put([]byte(kv.K), val)
		if err != nil {
			logger.L.Errorf("method:post path:/kv/ body:%v err=%s\n", kv, err)
			response(c, http.StatusInternalServerError, InternalError)
			metrics.PutCount.WithLabelValues("failure").Inc()
			return
		}
		cache.C.Del([]byte(kv.K))
		response(c, http.StatusOK, Ok)
		metrics.PutCount.WithLabelValues("success").Inc()
	})
	engine.DELETE("/kv/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			response(c, http.StatusBadRequest, InvalidArgument)
			return
		}
		err := store.S.Delete([]byte(key))
		if err == nil {
			cache.C.Del([]byte(key))
			response(c, http.StatusOK, Ok)
			return
		}
		if err == constant.KeyNotFound {
			response(c, http.StatusNotFound, KeyNotFound)
			return
		}
		logger.L.Errorf("method:delete path:/kv/%s err=%s\n", key, err)
		response(c, http.StatusInternalServerError, InternalError)
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
