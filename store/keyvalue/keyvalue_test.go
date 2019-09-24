package keyvalue

import (
	"encoding/hex"
	"testing"
)

func TestNewKeyValue(t *testing.T) {
	kv, err := NewKeyValue("k", "v")
	if err != nil {
		t.Error(err)
	}
	t.Log(kv)
}

func TestEncode(t *testing.T) {
	kv, err := NewKeyValue("k", "v")
	if err != nil {
		t.Error(err)
	}
	t.Log(hex.Dump(Encode(kv)))
}

func TestDecode(t *testing.T) {
	kv, err := NewKeyValue("k", "v")
	if err != nil {
		t.Error(err)
	}
	b := Encode(kv)
	nkv, err := Decode(b)
	if err != nil {
		t.Error(err)
	}
	t.Log(nkv)
}