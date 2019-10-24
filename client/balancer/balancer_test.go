package balancer

import (
	"math/rand"
	"testing"
	"time"
)

func TestBalancer_Choose(t *testing.T) {
	nodes := []string{"a", "b", "c"}
	for _, node := range nodes {
		D.Register(node)
	}
	for i := 0; i < 100000; i++ {
		node := D.Choose()
		t.Logf("%s\n", node)
		D.Increase(node, uint64(rand.Intn(100)))
	}
	time.Sleep(time.Second)
	for i := 0; i < 100000; i++ {
		node := D.Choose()
		t.Logf("%s\n", node)
		D.Increase(node, uint64(rand.Intn(100)))
	}
	time.Sleep(time.Second)
	for i := 0; i < 100000; i++ {
		node := D.Choose()
		t.Logf("%s\n", node)
		D.Increase(node, uint64(rand.Intn(100)))
	}
	time.Sleep(time.Second)
	for i := 0; i < 100000; i++ {
		node := D.Choose()
		t.Logf("%s\n", node)
		D.Increase(node, uint64(rand.Intn(100)))
	}
	time.Sleep(time.Second)
	for i := 0; i < 100000; i++ {
		node := D.Choose()
		t.Logf("%s\n", node)
		D.Increase(node, uint64(rand.Intn(100)))
	}
	time.Sleep(time.Second)
	for i := 0; i < 100000; i++ {
		node := D.Choose()
		t.Logf("%s\n", node)
		D.Increase(node, uint64(rand.Intn(100)))
	}
	time.Sleep(time.Second)
}