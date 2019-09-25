package appendfile

import (
	"dkv/store/keyvalue"
	"dkv/store/meta"
	"errors"
	"fmt"
	"io"
	"time"
)

type Item struct {
	fid    int64
	offset int32
	size   int32
}

type FileManager struct {
	meta     *meta.Meta
	activeAF *appendFile
	olderAF  []*appendFile
	index    map[string]*Item
	afmap    map[int64]*appendFile
}

func NewFileManager(dir string) (*FileManager, error) {
	m, err := meta.NewMeta(dir)
	if err != nil {
		return nil, err
	}
	if m.ActiveFid == 0 {
		m.ActiveFid = time.Now().UnixNano()
	}
	afmap := make(map[int64]*appendFile)
	activeAF, err := NewAppendFile(fmt.Sprintf("%s/%d", m.Dir, m.ActiveFid), Active, m.ActiveFid)
	if err != nil {
		return nil, err
	}
	afmap[m.ActiveFid] = activeAF
	olderAF := make([]*appendFile, 0)
	for _, fid := range m.OlderFids {
		af, err := NewAppendFile(fmt.Sprintf("%s/%d", m.Dir, fid), Older, fid)
		if err != nil {
			return nil, err
		}
		olderAF = append(olderAF, af)
		afmap[fid] = af
	}
	fm := &FileManager{
		meta:     m,
		activeAF: activeAF,
		olderAF:  olderAF,
		index:    make(map[string]*Item, 0),
		afmap:    afmap,
	}
	go func() {
		for range time.Tick(time.Second) {
			s, err := fm.activeAF.Size()
			if err != nil {
				continue
			}
			if s > 1024*1024*100 {
				fid := time.Now().UnixNano()
				af, err := NewAppendFile(fmt.Sprintf("%s/%d", m.Dir, fid), Active, fid)
				if err != nil {
					continue
				}
				oaf := fm.activeAF
				fm.olderAF = append(fm.olderAF, fm.activeAF)
				fm.activeAF = af
				oaf.role = Older
				fm.meta.ActiveFid = fid
				fm.meta.OlderFids = append(fm.meta.OlderFids, oaf.fid)
			}
			fm.meta.Save()
		}
	}()
	err = fm.Load()
	if err != nil {
		return nil, err
	}
	return fm, nil
}

func (fm *FileManager) Write(k []byte, v []byte) error {
	kv, err := keyvalue.NewKeyValue(k, v)
	if err != nil {
		return err
	}
	b := keyvalue.Encode(kv)
	off, err := fm.activeAF.Write(b)
	if err != nil {
		return err
	}
	fm.index[string(k)] = &Item{
		fid:    fm.activeAF.fid,
		offset: off,
		size:   int32(len(b)),
	}
	return nil
}

var KeyNotFound = errors.New("key is not found")

func (fm *FileManager) Read(k []byte) ([]byte, error) {
	item, state := fm.index[string(k)]
	if state == false {
		return nil, KeyNotFound
	}
	b := make([]byte, item.size)
	fm.afmap[item.fid].Read(int64(item.offset), b)
	kv, err := keyvalue.Decode(b)
	if err != nil {
		return nil, err
	}
	return kv.Value, nil
}

func (fm *FileManager) Load() error {
	if fm.meta.OlderFids != nil {
		for _, fid := range fm.meta.OlderFids {
			af := fm.afmap[fid]
			err := fm.loadAppendFile(af)
			if err != nil {
				return err
			}
		}
	}
	if fm.meta.ActiveFid != 0 {
		af := fm.afmap[fm.meta.ActiveFid]
		err := fm.loadAppendFile(af)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fm *FileManager) loadAppendFile(af *appendFile) error {
	b := make([]byte, 7)
	off := int64(0)
	var err error
	n := 0
	for {
		n, err = af.f.ReadAt(b, off)
		if err == io.EOF {
			return nil
		}
		if n < 7 {
			break
		}
		kv, err := keyvalue.DecodeHeader(b)
		if err != nil {
			return err
		}
		s := int(kv.KeySize) + int(kv.ValueSize)
		d := make([]byte, 7+s)
		n, err = af.f.ReadAt(d, off)
		kv, err = keyvalue.Decode(d)
		if err != nil {
			return err
		}
		fm.index[string(kv.Key)] = &Item{
			fid:    af.fid,
			offset: int32(off),
			size:   int32(7 + s),
		}
		off = off + int64(7) + int64(s)
	}
	return nil
}

func (fm *FileManager) Close() {
	fm.activeAF.Close()
	for _, af := range fm.olderAF {
		af.Close()
	}
	fm.meta.Save()
}
