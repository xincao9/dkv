package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zsais/go-gin-prometheus"
)

var (
	ObjectCurrentCount prometheus.Gauge
	PutCount           *prometheus.GaugeVec
	GetCount           *prometheus.GaugeVec
)

func init() {
	ObjectCurrentCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dkv_object_current_count",
		Help: "当前的对象数",
	})
	PutCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dkv_put_count",
		Help: "put操作数",
	}, []string{"status"})
	GetCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dkv_get_count",
		Help: "get操作数",
	}, []string{"status", "source"})
	prometheus.MustRegister(ObjectCurrentCount, PutCount, GetCount)
}

func Use(engine *gin.Engine) {
	p := ginprometheus.NewPrometheus("dkv")
	p.Use(engine)
}
