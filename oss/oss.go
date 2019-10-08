package oss

import (
	"crypto/md5"
	"dkv/store"
	"dkv/store/appendfile"
	"dkv/store/logger"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func Route(engine *gin.Engine) {

	// 获取对象　http://localhost:9090/oss/116a71ebd837470652f063028127c5cd
	engine.GET("/oss/:oid", func(c *gin.Context) {
		oid := c.Param("oid")
		if oid == "" {
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
		val, err := store.D.Get([]byte(oid))
		if err == nil {
			c.Data(http.StatusOK, "application/octet-stream", val)
			return
		}
		if err == appendfile.KeyNotFound {
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		logger.D.Errorf("oid = %s, err = %v\n", oid, err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	})

	// 对象上传　 curl -X POST 'http://localhost:9090/oss' -F "file[]=@config.yaml" -H 'content-type:multipart/form-data' -i
	engine.POST("/oss", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "参数错误",
			})
			return
		}
		files := form.File["file[]"]
		items := make([]struct {
			Filename string `json:"filename"`
			Oid      string `json:"oid"`
			status   bool   `json:"status"`
		}, len(files))
		for i, file := range files {
			f, err := file.Open()
			items[i].Filename = file.Filename
			items[i].status = true
			if err != nil {
				items[i].status = false
				continue
			}
			val, err := ioutil.ReadAll(f)
			if err != nil {
				items[i].status = false
				continue
			}
			h := md5.New()
			h.Write(val)
			key := []byte(hex.EncodeToString(h.Sum(nil)))
			_, err = store.D.Get(key)
			if err == appendfile.KeyNotFound {
				err = store.D.Put(key, val)
			}
			if err != nil {
				items[i].status = false
			}
			items[i].Oid = string(key)
		}
		c.JSON(http.StatusOK,
			gin.H{
				"code":    http.StatusOK,
				"message": "成功",
				"items":  items,
			})
	})
}
