package cache

import (
	"strconv"
	"testing"
)

func TestSet(t *testing.T) {
	C.Set([]byte("k"), []byte("v"))
}

func TestGet(t *testing.T) {
	C.Set([]byte("k"), []byte("v"))
	val := C.Get([]byte("k"))
	if val == nil {
		t.Error("应该包含 key = k, value = v")
	}
}

func TestDel(t *testing.T) {
	C.Set([]byte("k"), []byte("v"))
	val := C.Get([]byte("k"))
	if val == nil {
		t.Error("应该包含 key = k, value = v")
	}
	C.Del([]byte("k"))
	val = C.Get([]byte("k"))
	if val != nil {
		t.Error("不应该包含 key = k, value = v")
	}
}

func BenchmarkSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := []byte(strconv.Itoa(i))
		C.Set(c, c)
	}
}
