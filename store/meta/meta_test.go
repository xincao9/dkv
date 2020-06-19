package meta

import (
    "dkv/constant"
    "os"
	"testing"
	"time"
)

func TestNewMeta(t *testing.T) {
	os.RemoveAll(constant.DefaultDir)
	m, err := New()
	if err != nil {
		t.Error(m)
	}
	os.RemoveAll(constant.DefaultDir)
}

func TestMeta_Save(t *testing.T) {
	os.RemoveAll(constant.DefaultDir)
	m, err := New()
	if err != nil {
		t.Error(m)
	}
	m.OlderFids = []int64{time.Now().UnixNano()}
	m.ActiveFid = time.Now().UnixNano()
	err = m.Save()
	if err != nil {
		t.Error(err)
	}
	os.RemoveAll(constant.DefaultDir)
}
