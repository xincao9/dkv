package appendfile

import (
	"dkv/store/keyvalue"
	"dkv/store/meta"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var DeleteFlag = "S_D"

type Item struct {
	fid    int64
	offset int32
	size   int32
}

type FileManager struct {
	meta     *meta.Meta
	activeAF *appendFile
	olderAF  []*appendFile
	index    sync.Map
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
		index:    sync.Map{},
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
				fm.IndexSave()
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
	fm.index.Store(string(k), &Item{
		fid:    fm.activeAF.fid,
		offset: off,
		size:   int32(len(b)),
	})
	return nil
}

var KeyNotFound = errors.New("key is not found")

func (fm *FileManager) Read(k []byte) ([]byte, error) {
	v, state := fm.index.Load(string(k))
	if state == false {
		return nil, KeyNotFound
	}
	item := v.(*Item)
	b := make([]byte, item.size)
	fm.afmap[item.fid].Read(int64(item.offset), b)
	kv, err := keyvalue.Decode(b)
	if err != nil {
		return nil, err
	}
	if string(kv.Value) == DeleteFlag {
		fm.index.Delete(string(kv.Value))
		return nil, KeyNotFound
	}
	return kv.Value, nil
}

type i64 []int64

func (i i64) Len() int {
	return len(i)
}
func (i i64) Swap(x, y int) {
	i[x], i[y] = i[y], i[x]
}

func (i i64) Less(x, y int) bool {
	return i[x] < i[y]
}

func (fm *FileManager) Load() error {
	startTime := time.Now()
	log.Println("开始加载索引")
	if fm.meta.InvalidIndex {
		if fm.meta.OlderFids != nil {
			sort.Sort(i64(fm.meta.OlderFids))
			for _, fid := range fm.meta.OlderFids {
				af := fm.afmap[fid]
				err := fm.loadAppendFile(af)
				if err != nil {
					return err
				}
			}
		}
	} else {
		fm.IndexLoad()
	}
	if fm.meta.ActiveFid != 0 {
		af := fm.afmap[fm.meta.ActiveFid]
		err := fm.loadAppendFile(af)
		if err != nil {
			return err
		}
	}
	log.Printf("加载索引完成，耗时 %.2f 秒\n", time.Since(startTime).Seconds())
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
		} else if err != nil {
			return err
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
		fm.index.Store(string(kv.Key), &Item{
			fid:    af.fid,
			offset: int32(off),
			size:   int32(7 + s),
		})
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

var (
	byteOrder = binary.BigEndian
)

func (fm *FileManager) IndexSave() {
	fn := filepath.Join(fm.meta.Dir, "idx")
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return
	}
	fm.index.Range(func(key, value interface{}) bool {
		k := key.(string)
		i := value.(*Item)
		kl := len(k)
		b := make([]byte, kl+18)
		byteOrder.PutUint16(b[0:2], uint16(kl+18))
		byteOrder.PutUint64(b[2:10], uint64(i.fid))
		byteOrder.PutUint32(b[10:14], uint32(i.offset))
		byteOrder.PutUint32(b[14:18], uint32(i.size))
		copy(b[18:kl+18], k)
		_, err := f.Write(b)
		if err != nil {
			log.Printf("index save key = %s, item = %v, err = %v\n", k, i, err)
		}
		return true
	})
}

func (fm *FileManager) IndexLoad() error {
	fn := filepath.Join(fm.meta.Dir, "idx")
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	b := make([]byte, 2)
	off := int64(0)
	n := 0
	for {
		n, err = f.ReadAt(b, off)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		if n < 2 {
			break
		}
		s := byteOrder.Uint16(b)
		d := make([]byte, s)
		n, err = f.ReadAt(d, off)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		item := &Item{
			fid:    int64(byteOrder.Uint64(d)),
			offset: int32(byteOrder.Uint32(d)),
			size:   int32(byteOrder.Uint32(d)),
		}
		key := d[18:]
		fm.index.Store(string(key), item)
		off = off + int64(s)
	}
	return nil
}