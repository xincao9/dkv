package appendfile

import (
	"dkv/config"
	"dkv/logger"
	"dkv/metrics"
	"dkv/store/meta"
	"errors"
	"fmt"
	"golang.org/x/exp/mmap"
	"io"
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

type AppendFileManager struct {
	meta     *meta.Meta
	activeAF *appendFile
	olderAF  []*appendFile
	index    sync.Map
	afmap    sync.Map
}

func NewAppendFileManager(dir string) (*AppendFileManager, error) {
	m, err := meta.NewMeta(dir)
	if err != nil {
		return nil, err
	}
	if m.ActiveFid == 0 {
		m.ActiveFid = time.Now().UnixNano()
		m.Save()
	}
	afmap := sync.Map{}
	activeAF, err := NewAppendFile(fmt.Sprintf("%s/%d", m.Dir, m.ActiveFid), Active, m.ActiveFid)
	if err != nil {
		return nil, err
	}
	afmap.Store(m.ActiveFid, activeAF)
	olderAF := make([]*appendFile, 0)
	for _, fid := range m.OlderFids {
		af, err := NewAppendFile(fmt.Sprintf("%s/%d", m.Dir, fid), Older, fid)
		if err != nil {
			return nil, err
		}
		olderAF = append(olderAF, af)
		afmap.Store(fid, af)
	}
	fm := &AppendFileManager{
		meta:     m,
		activeAF: activeAF,
		olderAF:  olderAF,
		index:    sync.Map{},
		afmap:    afmap,
	}
	go func() {
		for range time.Tick(time.Second) {
			err := func() error {
				defer func() {
					if err := recover(); err != nil {
						logger.D.Errorf("定时任务异常 %v\n", err)
					}
				}()
				s, err := fm.activeAF.Size()
				if err != nil {
					return err
				}
				if s > 1024*1024*1024 {
					fid := time.Now().UnixNano()
					af, err := NewAppendFile(fmt.Sprintf("%s/%d", m.Dir, fid), Active, fid)
					if err != nil {
						return err
					}
					afmap.Store(fid, af)
					oaf := fm.activeAF
					fm.olderAF = append(fm.olderAF, oaf)
					fm.activeAF = af
					oaf.SetOlder()
					fm.meta.ActiveFid = fid
					fm.meta.OlderFids = append(fm.meta.OlderFids, oaf.fid)
					fm.meta.Save()
					fm.IndexSave()
				}
				return nil
			}()
			if err != nil {
				logger.D.Errorf("定时任务异常 %v\n", err)
			}
		}
	}()
	err = fm.Load()
	if err != nil {
		return nil, err
	}
	return fm, nil
}

func (fm *AppendFileManager) Write(k []byte, v []byte) error {
	kv, err := NewKeyValue(k, v)
	if err != nil {
		metrics.PutCount.WithLabelValues("failure").Inc()
		return err
	}
	b := Encode(kv)
	off, err := fm.activeAF.Write(b)
	if err != nil {
		metrics.PutCount.WithLabelValues("failure").Inc()
		return err
	}
	fm.index.Store(string(k), &Item{
		fid:    fm.activeAF.fid,
		offset: off,
		size:   int32(len(b)),
	})
	metrics.PutCount.WithLabelValues("success").Inc()
	return nil
}

var KeyNotFound = errors.New("key is not found")

func (fm *AppendFileManager) Read(k []byte) ([]byte, error) {
	v, state := fm.index.Load(string(k))
	if state == false {
		metrics.GetCount.WithLabelValues("failure").Inc()
		return nil, KeyNotFound
	}
	item := v.(*Item)
	b := make([]byte, item.size)
	af, ok := fm.afmap.Load(item.fid)
	if ok == false {
		metrics.GetCount.WithLabelValues("failure").Inc()
		return nil, fmt.Errorf("item = %v is exception", *item)
	}
	af.(*appendFile).Read(int64(item.offset), b)
	kv, err := Decode(b)
	if err != nil {
		metrics.GetCount.WithLabelValues("failure").Inc()
		return nil, err
	}
	if string(kv.Value) == DeleteFlag {
		metrics.GetCount.WithLabelValues("failure").Inc()
		fm.index.Delete(string(kv.Value))
		return nil, KeyNotFound
	}
	metrics.GetCount.WithLabelValues("success").Inc()
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

func (fm *AppendFileManager) Load() error {
	startTime := time.Now()
	logger.D.Infof("开始加载索引")
	if config.D.GetBool("data.invalidIndex") {
		if fm.meta.OlderFids != nil {
			sort.Sort(i64(fm.meta.OlderFids))
			for _, fid := range fm.meta.OlderFids {
				af, _ := fm.afmap.Load(fid)
				err := fm.loadAppendFile(af.(*appendFile))
				if err != nil {
					return err
				}
			}
		}
	} else {
		fm.IndexLoad()
	}
	if fm.meta.ActiveFid != 0 {
		af, _ := fm.afmap.Load(fm.meta.ActiveFid)
		err := fm.loadAppendFile(af.(*appendFile))
		if err != nil {
			return err
		}
	}
	logger.D.Infof("加载索引完成，耗时 %.2f 秒\n", time.Since(startTime).Seconds())
	return nil
}

func (fm *AppendFileManager) loadAppendFile(af *appendFile) error {
	b := make([]byte, 7)
	off := int64(0)
	var err error
	n := 0
	for {
		n, err = af.Read(off, b)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		if n < 7 {
			break
		}
		kv, err := DecodeHeader(b)
		if err != nil {
			return err
		}
		s := int(kv.KeySize) + int(kv.ValueSize)
		d := make([]byte, 7+s)
		n, err = af.Read(off, d)
		kv, err = Decode(d)
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

func (fm *AppendFileManager) Close() {
	fm.activeAF.Close()
	for _, af := range fm.olderAF {
		af.Close()
	}
	fm.meta.Save()
}

func (fm *AppendFileManager) IndexSave() {
	startTime := time.Now()
	defer func() {
		logger.D.Infof("index save 耗时: %.2f 秒\n", time.Since(startTime).Seconds())
	}()
	fn := filepath.Join(fm.meta.Dir, "idx")
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return
	}
	defer f.Close()
	os := 0
	fm.index.Range(func(key, value interface{}) bool {
		os++
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
			logger.D.Errorf("index save key = %s, item = %v, err = %v\n", k, i, err)
		}
		return true
	})
	metrics.ObjectCurrentCount.Set(float64(os))
}

func (fm *AppendFileManager) IndexLoad() error {
	fn := filepath.Join(fm.meta.Dir, "idx")
	f, err := mmap.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
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
			fid:    int64(byteOrder.Uint64(d[2:10])),
			offset: int32(byteOrder.Uint32(d[10:14])),
			size:   int32(byteOrder.Uint32(d[14:18])),
		}
		key := d[18:]
		fm.index.Store(string(key), item)
		off = off + int64(s)
	}
	return nil
}
