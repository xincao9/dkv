package ui

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func Route(engine *gin.Engine) {
	engine.Static("/css", "./static/css")
	engine.Static("/js", "./static/js")
	engine.Static("/image", "./static/image")
    engine.Static("/html", "./static/html")
	engine.LoadHTMLGlob("templates/*")
	engine.GET("/admin", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "dkv website",
		})
	})
}
