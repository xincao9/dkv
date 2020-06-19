package store

import (
    "dkv/component/constant"
    "os"
    "strconv"
    "testing"
)

var doc = make([]byte, 1024)

func BenchmarkStore_Put(b *testing.B) {
	os.RemoveAll(constant.DefaultDir)
	s, err := NewStore()
	if err != nil {
		b.Error(err)
	}
	for i := 0; i < b.N; i++ {
		err = s.Put([]byte(strconv.Itoa(i)), doc)
		if err != nil {
			b.Error(err)
		}
	}
	s.Close()
	os.RemoveAll(constant.DefaultDir)
}

func BenchmarkStore_Get(b *testing.B) {
	os.RemoveAll(constant.DefaultDir)
	s, err := NewStore()
	if err != nil {
		b.Error(err)
	}
	for i := 0; i < b.N; i++ {
		err = s.Put([]byte(strconv.Itoa(i)), doc)
		if err != nil {
			b.Error(err)
		}
		_, err = s.Get([]byte(strconv.Itoa(i)))
		if err != nil {
			b.Error(err)
		}
	}
	s.Close()
	os.RemoveAll(constant.DefaultDir)
}

func TestNew(t *testing.T) {
	os.RemoveAll(constant.DefaultDir)
	s, err := NewStore()
	if err != nil {
		t.Error(err)
	}
	s.Close()
	os.RemoveAll(constant.DefaultDir)
}

func TestStore_Get(t *testing.T) {
	os.RemoveAll(constant.DefaultDir)
	s, err := NewStore()
	if err != nil {
		t.Error(err)
	}
	err = s.Put([]byte("k"), []byte("v"))
	if err != nil {
		t.Error(err)
	}
	val, err := s.Get([]byte("k"))
	if err != nil {
		t.Error(err)
	}
	if string(val) != "v" {
		t.Errorf("value should be v, now %s\n", string(val))
	}
	s.Close()
	os.RemoveAll(constant.DefaultDir)
}

func TestStore_Delete(t *testing.T) {
	os.RemoveAll(constant.DefaultDir)
	s, err := NewStore()
	if err != nil {
		t.Error(err)
	}
	err = s.Put([]byte("k"), []byte("v"))
	if err != nil {
		t.Error(err)
	}
	err = s.Delete([]byte("k"))
	if err != nil {
		t.Error(err)
	}
	_, err = s.Get([]byte("k"))
	if err != constant.KeyNotFound {
		t.Error(err)
	}
	s.Close()
	os.RemoveAll(constant.DefaultDir)
}
