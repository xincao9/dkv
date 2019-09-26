package appendfile

import (
	"dkv/store/meta"
	"os"
	"strconv"
	"testing"
)

func TestNewFileManager(t *testing.T) {
	os.RemoveAll(meta.DefaultDir)
	fm, err := NewFileManager("")
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
	os.RemoveAll(meta.DefaultDir)
}

func BenchmarkFileManager_Write(b *testing.B) {
	os.RemoveAll(meta.DefaultDir)
	fm, err := NewFileManager("")
	if err != nil {
		b.Error(err)
	}
	defer fm.Close()
	for i := 0; i < b.N; i++ {
		v := []byte(strconv.Itoa(i))
		fm.Write(v, v)
	}
	os.RemoveAll(meta.DefaultDir)
}

func TestFileManager_Load(t *testing.T) {
	os.RemoveAll(meta.DefaultDir)
	fm, err := NewFileManager("")
	if err != nil {
		t.Error(err)
	}
	err = fm.Write([]byte("k"), []byte("v"))
	if err != nil {
		t.Error(err)
	}
	fm.Close()
	fm, err = NewFileManager("")
	if err != nil {
		t.Error(err)
	}
	val, err := fm.Read([]byte("k"))
	if err != nil {
		t.Error(err)
	}
	t.Logf("%s\n", val)
	os.RemoveAll(meta.DefaultDir)
}