package appendfile

import (
	"dkv/store/keyvalue"
	"dkv/store/meta"
	"fmt"
	"time"
)

type Item struct {
	fid    int64
	offset int32
	size   int32
}

type fileManager struct {
	meta     *meta.Meta
	activeAF *appendFile
	olderAF  []*appendFile
	index    map[string]*Item
	afmap    map[int64]*appendFile
}

func NewFileManager(dir string) (*fileManager, error) {
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
	fm := &fileManager{
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
	return fm, nil
}

func (fm fileManager) Write(k []byte, v []byte) error {
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

func (fm fileManager) Read(k []byte) ([]byte, error) {
	item, state := fm.index[string(k)]
	if state == false {
		return nil, fmt.Errorf("key {%s} is not found", k)
	}
	b := make([]byte, item.size)
	fm.afmap[item.fid].Read(int64(item.offset), b)
	kv, err := keyvalue.Decode(b)
	if err != nil {
		return nil, err
	}
	return kv.Value, nil
}

func (fm fileManager) Close() {
	fm.activeAF.Close()
	for _, af := range fm.olderAF {
		af.Close()
	}
	fm.meta.Save()
}
