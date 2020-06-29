package oss

import (
	"crypto/md5"
	"dkv/component/constant"
	"dkv/component/logger"
	"dkv/store"
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
		val, err := store.S.Get([]byte(oid))
		if err == nil {
			c.Data(http.StatusOK, http.DetectContentType(val), val)
			return
		}
		if err == constant.KeyNotFoundError {
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		logger.L.Errorf("oid = %s, err = %v\n", oid, err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
	})

	// 对象上传　 curl -X POST 'http://localhost:9090/oss' -F "file[]=@config.yaml" -H 'content-type:multipart/form-data' -i
	engine.POST("/oss", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": constant.InvalidArgument,
			})
			return
		}
		files := form.File["file[]"]
		items := make([]struct {
			Filename string `json:"filename"`
			Oid      string `json:"oid"`
			Status   bool   `json:"Status"`
		}, len(files))
		for i, file := range files {
			f, err := file.Open()
			items[i].Filename = file.Filename
			items[i].Status = true
			if err != nil {
				items[i].Status = false
				logger.L.Error(err)
				continue
			}
			val, err := ioutil.ReadAll(f)
			if err != nil {
				items[i].Status = false
				logger.L.Error(err)
				continue
			}
			h := md5.New()
			h.Write(val)
			key := []byte(hex.EncodeToString(h.Sum(nil)))
			_, err = store.S.Get(key)
			if err == constant.KeyNotFoundError {
				err = store.S.Put(key, val)
			}
			if err != nil {
				items[i].Status = false
				logger.L.Error(err)
				continue
			}
			items[i].Oid = string(key)
		}
		c.JSON(http.StatusOK,
			gin.H{
				"code":    http.StatusOK,
				"message": constant.Ok,
				"items":   items,
			})
	})
}
