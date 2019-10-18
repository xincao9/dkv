package agent

import (
	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
		Timeout: 0,
	}
)

func Route(engine *gin.Engine) {
	engine.Any("/*c", func(c *gin.Context) {
		uri := c.Request.RequestURI
		if !strings.HasPrefix(uri, "http") {
			uri = fmt.Sprintf("http://%s", uri)
		}
		request, err := http.NewRequest(c.Request.Method, uri, c.Request.Body)
		request.Header = c.Request.Header
		response, err := client.Do(request)
		if err != nil {
			log.Printf("method = %s, uri = %s, body = %s, err = %v", c.Request.Method, uri, c.Request.Body, err)
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


 func Bootstrap () {
 	engine := gin.Default()
 	Route(engine)
	 if err := engine.Run(":12306"); err != nil {
		 log.Fatalf("Fatal error gin: %v\n", err)
	 }
 }