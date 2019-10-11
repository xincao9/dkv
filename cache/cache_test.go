package cache

import (
	"testing"
	"time"
)

// ristretto set 操作异步生效
func TestSet(t *testing.T) {
	C.Set("k", "v", 0)
}

func TestGet(t *testing.T) {
	C.Set("k", "v", 0)
	time.Sleep(time.Second)
	_, state := C.Get("k")
	if state == false {
		t.Error("应该包含 key = k, value = v")
	}
}

// ristretto del 操作异步生效
func TestDel(t *testing.T) {
	C.Set("k", "v", 0)
	time.Sleep(time.Second)
	_, state := C.Get("k")
	if state == false {
		t.Error("应该包含 key = k, value = v")
	}
	C.Del("k")
	time.Sleep(time.Second)
	_, state = C.Get("k")
	if state {
		t.Error("不应该包含 key = k, value = v")
	}

}

func BenchmarkSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		C.Set(i, i, 0)
	}
}
