package appendfile

import (
	"dkv/config"
	"dkv/logger"
	"dkv/metrics"
	"dkv/store/meta"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DeleteFlag = "S_D"
	idle       = 0
	running    = 1
)

var (
	KeyNotFound = errors.New("key is not found")
)

type (
	Item struct {
		fid    int64
		offset int64
		size   int32
	}
	AppendFileManager struct {
		meta     *meta.Meta
		activeAF *appendFile
		olderAF  []*appendFile
		index    sync.Map
		afmap    sync.Map
		counter  sync.Map
		rot      int64 // Recent operation time
		sistate  int32 // save index state
	}
	i64 []int64
)

func (i i64) Len() int {
	return len(i)
}
func (i i64) Swap(x, y int) {
	i[x], i[y] = i[y], i[x]
}

func (i i64) Less(x, y int) bool {
	return i[x] < i[y]
}

func NewAppendFileManager(dir string) (*AppendFileManager, error) {
	var err error
	m, err := meta.NewMeta(dir)
	if err != nil {
		return nil, err
	}
	if m.ActiveFid == 0 {
		m.ActiveFid = time.Now().UnixNano()
		m.Save()
	}
	afmap := sync.Map{}
	activeAF, err := NewAppendFile(filepath.Join(m.Dir, strconv.FormatInt(m.ActiveFid, 10)), Active, m.ActiveFid)
	if err != nil {
		logger.D.Errorf("open active fid=%d, err=%v\n", m.ActiveFid, err)
		return nil, err
	} else {
		logger.D.Infof("open active fid=%d, success\n", m.ActiveFid)
	}
	afmap.Store(m.ActiveFid, activeAF)
	olderAF := make([]*appendFile, 0)
	for _, fid := range m.OlderFids {
		af, err := NewAppendFile(filepath.Join(m.Dir, strconv.FormatInt(fid, 10)), Older, fid)
		if err != nil {
			logger.D.Errorf("open older fid=%d, err=%v\n", fid, err)
			return nil, err
		} else {
			logger.D.Infof("open older fid=%d, success\n", fid)
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
						logger.D.Errorf("文件回滚定时任务异常 %v\n", err)
					}
				}()
				fm.activeAF.Sync()
				s, err := fm.activeAF.Size()
				if err != nil {
					return err
				}
				if s < 1024*1024*1024 {
					return nil
				}
				fid := time.Now().UnixNano()
				af, err := NewAppendFile(filepath.Join(m.Dir, strconv.FormatInt(fid, 10)), Active, fid)
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
				if atomic.CompareAndSwapInt32(&fm.sistate, idle, running) {
					go func() { fm.IndexSave() }()
				}
				return nil
			}()
			if err != nil {
				logger.D.Errorf("文件回滚定时任务异常 %v\n", err)
			}
		}
	}()
	go func() {
		for range time.Tick(time.Minute) {
			if time.Now().Unix()-fm.rot <= 300 {
				continue
			}
			err := func() error {
				defer func() {
					if err := recover(); err != nil {
						logger.D.Errorf("文件merge定时任务异常 %v\n", err)
					}
				}()
				fm.counter.Range(func(fid, value interface{}) bool {
					count := value.(*uint32)
					if *count <= 200 {
						return true
					}
					af, state := fm.afmap.Load(fid)
					if state == false {
						return true
					}
					if err = fm.Merge(af.(*appendFile)); err != nil {
						logger.D.Errorf("fid = %d merge failure, err = %v\n", fid.(int64), err)
					} else {
						fm.Remove(af.(*appendFile))
						fm.counter.Delete(fid)
						logger.D.Infof("fid = %d merge success\n", fid.(int64))
					}
					return true
				})
				return nil
			}()
			if err != nil {
				logger.D.Errorf("文件merge定时任务异常 %v\n", err)
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
	fm.rot = time.Now().Unix()
	kv, err := NewKeyValue(k, v)
	if err != nil {
		return err
	}
	b := Encode(kv)
	af := fm.activeAF
	off, err := af.Write(b)
	if err != nil {
		return fmt.Errorf("write(%s, %s) %v", k, v, err)
	}
	val, state := fm.index.Load(string(k))
	if state {
		item := val.(*Item)
		c, state := fm.counter.Load(item.fid)
		if state {
			atomic.AddUint32(c.(*uint32), 1)
		} else {
			count := uint32(1)
			fm.counter.Store(item.fid, &count)
		}
	}
	fm.index.Store(string(k), &Item{
		fid:    af.fid,
		offset: off,
		size:   int32(len(b)),
	})
	return nil
}

func (fm *AppendFileManager) Read(k []byte) ([]byte, error) {
	fm.rot = time.Now().Unix()
	v, state := fm.index.Load(string(k))
	if state == false {
		return nil, KeyNotFound
	}
	item := v.(*Item)
	b := make([]byte, item.size)
	af, ok := fm.afmap.Load(item.fid)
	if ok == false {
		return nil, fmt.Errorf("read(%v) fid = %d not found", *item, item.fid)
	}
	n, err := af.(*appendFile).Read(item.offset, b)
	if err != nil {
		return nil, fmt.Errorf("read(%v) read %v", *item, err)
	}
	if n != int(item.size) {
		return nil, fmt.Errorf("read(%v) read 期望读取 %d 字节，实际读取 %d 字节", *item, item.size, n)
	}
	kv, err := Decode(b)
	if err != nil {
		return nil, fmt.Errorf("read(%v) decode %v", *item, err)
	}
	if string(kv.Value) == DeleteFlag {
		fm.index.Delete(string(kv.Value))
		return nil, KeyNotFound
	}
	return kv.Value, nil
}

func (fm *AppendFileManager) GetAppendFiles() []string {
	var fns []string
	if fm.meta.OlderFids != nil {
		sort.Sort(i64(fm.meta.OlderFids))
		for _, fid := range fm.meta.OlderFids {
			if fid != 0 {
				fn := filepath.Join(fm.meta.Dir, strconv.FormatInt(fid, 10))
				fns = append(fns, fn)
			}
		}
	}
	if fm.meta.ActiveFid != 0 {
		fn := filepath.Join(fm.meta.Dir, strconv.FormatInt(fm.meta.ActiveFid, 10))
		fns = append(fns, fn)
	}
	return fns
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
	b := make([]byte, 9)
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
		if n < 9 {
			return nil
		}
		kv, err := DecodeHeader(b)
		if err != nil {
			return err
		}
		s := int(kv.KeySize) + int(kv.ValueSize)
		d := make([]byte, 9+s)
		n, err = af.Read(off, d)
		if err != nil {
			return err
		}
		if n < 9+s {
			return nil
		}
		kv, err = Decode(d)
		if err != nil {
			return err
		}
		fm.index.Store(string(kv.Key), &Item{
			fid:    af.fid,
			offset: off,
			size:   int32(9 + s),
		})
		off = off + int64(9) + int64(s)
	}
}

func (fm *AppendFileManager) Close() {
	fm.activeAF.Close()
	for _, af := range fm.olderAF {
		af.Close()
	}
	fm.meta.Save()
}

func (fm *AppendFileManager) IndexSave() {
	defer func() {
		atomic.CompareAndSwapInt32(&fm.sistate, running, idle)
	}()
	startTime := time.Now()
	defer func() {
		logger.D.Infof("index save 耗时: %.2f 秒\n", time.Since(startTime).Seconds())
	}()
	fn := filepath.Join(fm.meta.Dir, "idx")
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0644)
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
		b := make([]byte, kl+22)
		byteOrder.PutUint16(b[0:2], uint16(kl+22))
		byteOrder.PutUint64(b[2:10], uint64(i.fid))
		byteOrder.PutUint64(b[10:18], uint64(i.offset))
		byteOrder.PutUint32(b[18:22], uint32(i.size))
		copy(b[22:kl+22], k)
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
	f, err := os.OpenFile(fn, os.O_RDONLY, 0644)
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
			offset: int64(byteOrder.Uint64(d[10:18])),
			size:   int32(byteOrder.Uint32(d[18:22])),
		}
		key := d[22:]
		fm.index.Store(string(key), item)
		off = off + int64(s)
	}
	return nil
}

func (fm *AppendFileManager) Merge(af *appendFile) error {
	b := make([]byte, 9)
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
		if n < 9 {
			break
		}
		kv, err := DecodeHeader(b)
		if err != nil {
			return err
		}
		s := int(kv.KeySize) + int(kv.ValueSize)
		d := make([]byte, 9+s)
		n, err = af.Read(off, d)
		if err != nil {
			return err
		}
		if n < 9+s {
			return nil
		}
		kv, err = Decode(d)
		if err != nil {
			return err
		}
		v, state := fm.index.Load(string(kv.Key))
		if state {
			item := v.(*Item)
			if af.fid == item.fid && off == item.offset && int32(9+s) == item.size {
				if string(kv.Value) != DeleteFlag {
					err = fm.Write(kv.Key, kv.Value)
					if err != nil {
						return err
					}
				}
			}
		}
		off = off + int64(9) + int64(s)
	}
	return nil
}

func (fm *AppendFileManager) Remove(af *appendFile) {
	if af.role != Older {
		return
	}
	fm.afmap.Delete(af.fid)
	af.Close()
	os.Remove(af.fn)
	ofids := make([]int64, 0)
	for _, fid := range fm.meta.OlderFids {
		if fid != af.fid {
			ofids = append(ofids, fid)
		}
	}
	fm.meta.OlderFids = ofids
	fm.meta.Save()
}
