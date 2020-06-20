package balancer

import (
	"testing"
)

func TestBalancer_Choose(t *testing.T) {
	nodes := map[string]int64{"a": 0, "b": 0, "c": 0}
	for node, _ := range nodes {
		B.Register(node)
	}
	for i := 0; i < 100000; i++ {
		node := B.Choose()
		nodes[node]++
		B.Increment()
	}
	t.Logf("%v\n", nodes)
}

func BenchmarkBalancer_Choose(b *testing.B) {
	nodes := []string{"a", "b", "c"}
	for _, node := range nodes {
		B.Register(node)
	}
	for i := 0; i < b.N; i++ {
		node := B.Choose()
		b.Logf("%s\n", node)
		B.Increment()
	}
}
