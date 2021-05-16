package client

import (
	"sync/atomic"
)

var (
	B *balancer
)

func init() {
	B = new()
}

type balancer struct {
	nodes   []string
	counter uint64
	l       uint64
}

func new() *balancer {
	return &balancer{nodes: make([]string, 0, 10)}
}

func (lb *balancer) Register(node string) {
	lb.nodes = append(lb.nodes, node)
	lb.l = uint64(len(lb.nodes))
}

func (lb *balancer) Increment() {
	atomic.AddUint64(&lb.counter, 1)
}

func (lb *balancer) Choose() string {
	return lb.nodes[lb.counter%lb.l]
}
