package balancer

import (
    "testing"
)

func TestBalancer_Choose(t *testing.T) {
	nodes := map[string]int64{"a": 0, "b": 0, "c": 0}
	for node, _ := range nodes {
		D.Register(node)
	}
	for i := 0; i < 100000; i++ {
		node := D.Choose()
		nodes[node]++
		D.Increment()
	}
    t.Logf("%v\n", nodes)
}

func BenchmarkBalancer_Choose(b *testing.B) {
	nodes := []string{"a", "b", "c"}
	for _, node := range nodes {
		D.Register(node)
	}
	for i := 0; i < b.N; i++ {
		node := D.Choose()
		b.Logf("%s\n", node)
		D.Increment()
	}
}
