package metrics

import "github.com/prometheus/client_golang/prometheus"

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
	},
		[]string{"status"})
	GetCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dkv_get_count",
		Help: "get操作数",
	}, []string{"status"})
	prometheus.MustRegister(ObjectCurrentCount, PutCount, GetCount)
}
