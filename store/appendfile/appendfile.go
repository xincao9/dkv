package appendfile

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync"
)

const (
	Older  = 1
	Active = 2
)

type appendFile struct {
	fn     string
	offset int64
	role   int
	f      *os.File
	fid    int64
	sync.Mutex
}

func NewAppendFile(fn string, role int, fid int64) (*appendFile, error) {
	if role != Older && role != Active {
		return nil, fmt.Errorf("role {%d} is not found", role)
	}
	af := &appendFile{
		fn:     fn,
		offset: 0,
		role:   role,
		fid:    fid,
	}
	var err error
	if role == Active {
		af.f, err = os.OpenFile(fn, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		af.offset, err = af.Size()
		if err != nil {
			debug.PrintStack()
			return nil, err
		}
		return af, nil
	}
	af.f, err = os.OpenFile(fn, os.O_RDONLY, 0644)
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	return af, nil
}

func (af *appendFile) Write(b []byte) (int64, error) {
	if af.role == Older {
		return -1, fmt.Errorf("write operations are not supported, %v\n", af)
	}
	af.Lock()
	defer af.Unlock()
	off := af.offset
	n, err := af.f.Write(b)
	if err != nil {
		return -1, err
	}
	if n != len(b) {
		af.f.Seek(off, 0)
		return -1, fmt.Errorf("write %d bytes, actually write %d bytes", len(b), n)
	}
	af.offset += int64(n)
	return off, nil
}

func (af *appendFile) Read(offset int64, b []byte) (int, error) {
	af.Lock()
	defer af.Unlock()
	return af.f.ReadAt(b, offset)
}

func (af *appendFile) Size() (int64, error) {
	fi, err := af.f.Stat()
	if err != nil {
		return -1, err
	}
	return fi.Size(), nil
}

func (af *appendFile) SetOlder() {
	af.role = Older
	af.Sync()
}

func (af *appendFile) Close() {
	if af.f != nil {
		af.Sync()
		af.f.Close()
	}
}

func (af *appendFile) Sync() {
	af.f.Sync()
}
