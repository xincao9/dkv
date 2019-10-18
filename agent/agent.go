package agent

import (
	"dkv/logger"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
)

var (
	client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
		Timeout: 0,
	}
)

func Route(engine *gin.Engine) {
	engine.Any("/*c", func(c *gin.Context) {
		request, err := http.NewRequest(c.Request.Method, c.Request.RequestURI, c.Request.Body)
		request.Header = c.Request.Header
		response, err := client.Do(request)
		if err != nil {
			logger.D.Errorf("method = %s, uri = %s, body = %s", c.Request.Method, c.Request.RequestURI, c.Request.Body, err)
			c.Writer.WriteHeader(http.StatusBadGateway)
			return
		}
		defer response.Body.Close()
		c.Writer.WriteHeader(response.StatusCode)
		if response.Header != nil {
			for name, values := range response.Header {
				for _, value := range values {
					c.Writer.Header().Add(name, value)
				}
			}
		}
		io.Copy(c.Writer, response.Body)
	})
}