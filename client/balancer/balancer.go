package balancer

import (
	"math"
	"sync"
	"sync/atomic"
)

var (
	D *balancer
)

func init() {
	D = New()
}

type balancer struct {
	ct *sync.Map
}

func New() *balancer {
	return &balancer{ct: &sync.Map{}}
}

func (lb *balancer) Register(node string) {
	c := uint64(0)
	lb.ct.Store(node, &c)
}

func (lb *balancer) Add(node string, v uint64) {
	val, ok := lb.ct.Load(node)
	if ok == false {
		return
	}
	c, _ := val.(*uint64)
	n := atomic.AddUint64(c, v)
	if n >= math.MaxUint64 {
		lb.ct.Range(func(key, value interface{}) bool {
			rn, _ := key.(string)
			lb.Register(rn)
			return true
		})
	}
}

func (lb *balancer) Choose() string {
	min := uint64(math.MaxUint64)
	cn := ""
	lb.ct.Range(func(key, value interface{}) bool {
		c, _ := value.(uint64)
		if c < min {
			min = c
			cn, _ = key.(string)
		}
		return true
	})
	return cn
}
