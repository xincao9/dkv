package appendfile

import (
	"strconv"
	"testing"
)

func TestNewFileManager(t *testing.T) {
	fm, err := NewAppendFileManager()
	if err != nil {
		t.Error(err)
	}
	defer fm.Close()
	err = fm.Write([]byte("k"), []byte("v"))
	if err != nil {
		t.Error(err)
	}
	val, err := fm.Read([]byte("k"))
	if err != nil {
		t.Error(err)
	}
	t.Logf("%s\n", val)
}

func BenchmarkFileManager_Write(b *testing.B) {
	fm, err := NewAppendFileManager()
	if err != nil {
		b.Error(err)
	}
	defer fm.Close()
	for i := 0; i < b.N; i++ {
		v := []byte(strconv.Itoa(i))
		fm.Write(v, v)
	}
}

func TestFileManager_Load(t *testing.T) {
	fm, err := NewAppendFileManager()
	if err != nil {
		t.Error(err)
	}
	err = fm.Write([]byte("k"), []byte("v"))
	if err != nil {
		t.Error(err)
	}
	fm.Close()
	fm, err = NewAppendFileManager()
	if err != nil {
		t.Error(err)
	}
	val, err := fm.Read([]byte("k"))
	if err != nil {
		t.Error(err)
	}
	t.Logf("%s\n", val)
}
