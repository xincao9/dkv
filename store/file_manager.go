package store

import (
	"fmt"
	"time"
)

type keyInfo struct {
	f      int32
	offset int32
	size   int32
}

type fileManager struct {
	dir      string
	cur      int32
	activeAF *appendFile
	olderAF  []*appendFile
	index    map[string]*keyInfo
}

func NewFileManager(dir string, cur int32) (*fileManager, error) {
	if cur < 0 {
		return nil, fmt.Errorf("cur　{%d} is invalid argument", cur)
	}
	activeAF, err := NewAppendFile(fmt.Sprintf("%s/%d", dir, cur), Active)
	if err != nil {
		return nil, err
	}
	olderAF := make([]*appendFile, 0)
	for no := 0; no < int(cur); no++ {
		af, err := NewAppendFile(fmt.Sprintf("%s/%d", dir, no), Older)
		if err != nil {
			return nil, err
		}
		olderAF = append(olderAF, af)
	}
	olderAF = append(olderAF, activeAF)
	fm := &fileManager{
		dir:      dir,
		cur:      cur,
		activeAF: activeAF,
		olderAF:  olderAF,
		index:    make(map[string]*keyInfo, 0),
	}
	go func() {
		for range time.Tick(time.Second) {
			s, err := fm.activeAF.Size()
			if err != nil {
				continue
			}
			if s > 1024 * 1024 {
				af, err := NewAppendFile(fmt.Sprintf("%s/%d", dir, fm.cur+1), Active)
				if err != nil {
					continue
				}
				fm.olderAF = append(fm.olderAF, af)
				fm.activeAF = af
				fm.cur++
			}
		}
	}()
	return fm, nil
}

func (fm fileManager) Write(key string, b []byte) error {
	off, err := fm.activeAF.Write(b)
	if err != nil {
		return err
	}
	fm.index[key] = &keyInfo{
		f:      fm.cur,
		offset: off,
		size:   int32(len(b)),
	}
	return nil
}

func (fm fileManager) Read(key string) ([]byte, error) {
	ki, state := fm.index[key]
	if state == false {
		return nil, fmt.Errorf("key {%s} is not found", key)
	}
	b := make([]byte, ki.size)
	fm.olderAF[ki.f].Read(int64(ki.offset), b)
	return b, nil
}

func (fm fileManager) Close () {
	for _, af := range fm.olderAF {
		af.Close()
	}
}
