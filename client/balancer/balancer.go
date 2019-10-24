package balancer

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

var (
	D *balancer
)

func init () {
	D = New()
}

type balancer struct {
	counter *sync.Map
	oc      *sync.Map
	tps     *sync.Map
}

func New() *balancer {
	b := &balancer{counter: &sync.Map{}, oc: &sync.Map{}, tps: &sync.Map{}}
	go func() {
		for range time.Tick(time.Second) {
			b.counter.Range(func(label, v interface{}) bool {
				ov, _ := b.oc.Load(label)
				c, _ := v.(*uint64)
				oc, _ := ov.(*uint64)
				b.tps.Store(label, (*c-*oc)/30)
				*oc = *c
				b.oc.Store(label, oc)
				return true
			})
		}
	}()
	return b
}

func (lb *balancer) Register (node string) {
	c := uint64(0)
	lb.counter.Store(node, &c)
	c = uint64(0)
	lb.oc.Store(node, &c)
}

func (lb *balancer) Increase(node string, v uint64) {
	val, ok := lb.counter.Load(node)
	if ok == false {
		return
	}
	c, _ := val.(*uint64)
	atomic.AddUint64(c, v)
}

func (lb *balancer) Choose() string {
	min := uint64(math.MaxUint64)
	cn := ""
	lb.tps.Range(func(key, value interface{}) bool {
		c, _ := value.(uint64)
		if c < min {
			min = c
			cn, _ = key.(string)
		}
		return true
	})
	return cn
}
