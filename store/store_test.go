package store

import (
	"dkv/store/appendfile"
	"dkv/store/meta"
	"os"
	"strconv"
	"testing"
)

var doc = make([]byte, 1024)

func BenchmarkStore_Put(b *testing.B) {
	os.RemoveAll(meta.DefaultDir)
	s, err := NewStore("")
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
	os.RemoveAll(meta.DefaultDir)
}

func BenchmarkStore_Get(b *testing.B) {
	os.RemoveAll(meta.DefaultDir)
	s, err := NewStore("")
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
	os.RemoveAll(meta.DefaultDir)
}

func TestNew(t *testing.T) {
	os.RemoveAll(meta.DefaultDir)
	s, err := NewStore("")
	if err != nil {
		t.Error(err)
	}
	s.Close()
	os.RemoveAll(meta.DefaultDir)
}

func TestStore_Get(t *testing.T) {
	os.RemoveAll(meta.DefaultDir)
	s, err := NewStore("")
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
	os.RemoveAll(meta.DefaultDir)
}

func TestStore_Delete(t *testing.T) {
	os.RemoveAll(meta.DefaultDir)
	s, err := NewStore("")
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
	if err != appendfile.KeyNotFound {
		t.Error(err)
	}
	s.Close()
	os.RemoveAll(meta.DefaultDir)
}
