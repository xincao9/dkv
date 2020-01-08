package ms

import (
	"testing"
	"time"
)

func TestPut(t *testing.T) {
	c, err := New("localhost:9090", time.Second)
	if err != nil {
		t.Error(err)
	}
	_, err = c.Put("k", "v")
	if err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	c, err := New("localhost:9090", time.Second)
	if err != nil {
		t.Error(err)
	}
	_, err = c.Put("k", "v")
	if err != nil {
		t.Error(err)
	}
	r, err := c.Get("k")
	if err != nil {
		t.Error(err)
	}
	if r.KV.V != "v" {
		t.Errorf("应该为 v, 实际为 %s", r.KV.V)
	}
}

func TestDelete(t *testing.T) {
	c, err := New("localhost:9090", time.Second)
	if err != nil {
		t.Error(err)
	}
	_, err = c.Put("k", "v")
	if err != nil {
		t.Error(err)
	}
	r, err := c.Delete("k")
	if err != nil {
		t.Error(err)
	}
	r, err = c.Get("v")
	if err != nil {
		t.Error(err)
	}
	if r.KV.V != "" {
		t.Errorf("应该为空字符串, 实际为 %s", r.KV.V)
	}
}
