package appendfile

import (
	"fmt"
	"os"
)

const (
	Older  = 1
	Active = 2
)

type appendFile struct {
	fn     string
	offset int32
	role   int
	f      *os.File
	fid    int64
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
	af.f, err = os.OpenFile(fn, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	off, err := af.Size()
	if err != nil {
		return nil, err
	}
	af.offset = int32(off)
	return af, nil
}

func (af *appendFile) Write(b []byte) (int32, error) {
	if af.role == Older {
		return -1, fmt.Errorf("write operations are not supported, %v\n", af)
	}
	off := af.offset
	n, err := af.f.Write(b)
	if err != nil {
		return -1, err
	}
	af.offset += int32(n)
	return off, nil
}

func (af *appendFile) Read(offset int64, b []byte) {
	af.f.Seek(offset, 0)
	af.f.Read(b)
}

func (af *appendFile) Size() (int64, error) {
	fi, err := af.f.Stat()
	if err != nil {
		return -1, err
	}
	return fi.Size(), nil
}

func (af *appendFile) Close() {
	af.f.Close()
}
