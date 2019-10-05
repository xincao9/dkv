package store

import (
	"dkv/store/appendfile"
	"dkv/store/meta"
	"os"
	"testing"
)

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
