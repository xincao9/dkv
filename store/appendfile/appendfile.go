package appendfile

import (
    "dkv/component/logger"
    "dkv/constant"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "sync"
)

type appendFile struct {
	fn     string
	offset int64
	role   int
	fo     *os.File
	fid    int64
	sync.Mutex
}

func NewAppendFile(dir string, role int, fid int64) (*appendFile, error) {
	if role != constant.Older && role != constant.Active {
		return nil, fmt.Errorf("role {%d} is not found", role)
	}
	af := &appendFile{
		fn:     filepath.Join(dir, strconv.FormatInt(fid, 10)),
		offset: 0,
		role:   role,
		fid:    fid,
	}
	var err error
	if role == constant.Active {
		af.fo, err = os.OpenFile(af.fn, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		af.offset, err = af.Size()
		if err != nil {
			return nil, err
		}
		return af, nil
	}
	af.fo, err = os.OpenFile(af.fn, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return af, nil
}

func (af *appendFile) Write(b []byte) (int64, error) {
	if af.role == constant.Older {
		return -1, fmt.Errorf("write operations are not supported, %v\n", af.fn)
	}
	af.Lock()
	defer af.Unlock()
	off := af.offset
	n, err := af.fo.Write(b)
	if err != nil {
		return -1, err
	}
	if n != len(b) {
		af.fo.Seek(off, 0)
		return -1, fmt.Errorf("write %d bytes, actually write %d bytes", len(b), n)
	}
	af.offset += int64(n)
	return off, nil
}

func (af *appendFile) Read(offset int64, b []byte) (int, error) {
	return af.fo.ReadAt(b, offset)
}

func (af *appendFile) Size() (int64, error) {
	fi, err := af.fo.Stat()
	if err != nil {
		return -1, err
	}
	return fi.Size(), nil
}

func (af *appendFile) SetOlder() {
	af.role = constant.Older
	af.Sync()
}

func (af *appendFile) Close() {
	af.Sync()
	if af.fo != nil {
		af.fo.Close()
	}
}

func (af *appendFile) Sync() {
	if af.fo != nil {
		err := af.fo.Sync()
		if err != nil {
			logger.L.Errorf("sync: %v\n", err)
		}
	}
}
