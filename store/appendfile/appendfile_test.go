package appendfile

import (
	"dkv/constant"
	"os"
	"testing"
)

func TestNewAppendFile(t *testing.T) {
	os.Remove("/tmp/0")
	af, err := NewAppendFile("/tmp/0", constant.Active, 0)
	if err != nil {
		t.Error(err)
	}
	af.Close()
	af, err = NewAppendFile("/tmp/0", constant.Older, 0)
	if err != nil {
		t.Error(err)
	}
	af.Close()
	os.Remove("/tmp/0")
}

func TestAppendFile_Write(t *testing.T) {
	os.Remove("/tmp/0")
	af, err := NewAppendFile("/tmp/0", constant.Active, 0)
	if err != nil {
		t.Error(err)
	}
	af.Write([]byte("hello world!"))
	af.Close()
	os.Remove("/tmp/0")
}

func TestAppendFile_Read(t *testing.T) {
	os.Remove("/tmp/0")
	af, err := NewAppendFile("/tmp/0", constant.Active, 0)
	if err != nil {
		t.Error(err)
	}
	off, err := af.Write([]byte("hello world!"))
	af.Close()
	if err != nil {
		t.Error(err)
	}
	af, err = NewAppendFile("/tmp/0", constant.Older, 0)
	if err != nil {
		t.Error(err)
	}
	defer af.Close()
	b := make([]byte, len("hello world!"))
	af.Read(int64(off), b)
	t.Logf("read %s\n", b)
	os.Remove("/tmp/0")
}
