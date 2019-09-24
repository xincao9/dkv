package meta

import "testing"

func TestNewMeta(t *testing.T) {
	m, err := NewMeta("")
	if err != nil {
		t.Error(m)
	}
}

func TestMeta_Save(t *testing.T) {
	m, err := NewMeta("")
	if err != nil {
		t.Error(m)
	}
	m.Fids = []string{"a, b"}
	m.ActiveFid = "a"
	m.Save()
}
