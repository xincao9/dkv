package appendfile

import (
	"bytes"
	"dkv/component/constant"
	"dkv/component/logger"
	"dkv/component/metrics"
	"dkv/store/meta"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type (
	Item struct {
		fid    int64
		offset int64
		size   int32
	}
	AppendFileManager struct {
		activeAF *appendFile
		olderAF  []*appendFile
		index    sync.Map
		afmap    sync.Map
		counter  sync.Map
		rot      int64 // Recent operation time
		sistate  int32 // save index state
		loadtime time.Time
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

func NewAppendFileManager() (*AppendFileManager, error) {
	if meta.M.ActiveFid == 0 {
		meta.M.ActiveFid = time.Now().UnixNano()
		meta.M.Save()
	}
	afmap := sync.Map{}
	activeAF, err := NewAppendFile(constant.Dir, constant.Active, meta.M.ActiveFid)
	if err != nil {
		logger.L.Errorf("open active fid=%d, err=%v\n", meta.M.ActiveFid, err)
		return nil, err
	} else {
		logger.L.Infof("open active fid=%d, success\n", meta.M.ActiveFid)
	}
	afmap.Store(meta.M.ActiveFid, activeAF)
	olderAF := make([]*appendFile, 0)
	for _, fid := range meta.M.OlderFids {
		af, err := NewAppendFile(constant.Dir, constant.Older, fid)
		if err != nil {
			logger.L.Errorf("open older fid=%d, err=%v\n", fid, err)
			return nil, err
		} else {
			logger.L.Infof("open older fid=%d, success\n", fid)
		}
		olderAF = append(olderAF, af)
		afmap.Store(fid, af)
	}
	fm := &AppendFileManager{
		activeAF: activeAF,
		olderAF:  olderAF,
		index:    sync.Map{},
		afmap:    afmap,
		loadtime: time.Now(),
	}
	if constant.MSRole != constant.Slave {
		go func() {
			for range time.Tick(time.Second) {
				err := func() error {
					defer func() {
						if err := recover(); err != nil {
							logger.L.Errorf("文件回滚定时任务异常 %v\n", err)
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
					af, err := NewAppendFile(constant.Dir, constant.Active, fid)
					if err != nil {
						return err
					}
					afmap.Store(fid, af)
					oaf := fm.activeAF
					fm.olderAF = append(fm.olderAF, oaf)
					fm.activeAF = af
					oaf.SetOlder()
					meta.M.ActiveFid = fid
					meta.M.OlderFids = append(meta.M.OlderFids, oaf.fid)
					meta.M.Save()
					if atomic.CompareAndSwapInt32(&fm.sistate, constant.Idle, constant.Running) {
						go func() { fm.IndexSave() }()
					}
					return nil
				}()
				if err != nil {
					logger.L.Errorf("文件回滚定时任务异常 %v\n", err)
				}
			}
		}()
	} else {
		go func() {
			for range time.Tick(time.Second) {
				if time.Since(fm.loadtime).Seconds() > 60 {
					fm.loadtime = time.Now()
					err = fm.loadAppendFile(fm.activeAF)
					if err != nil {
						logger.L.Errorf("loadAppendFile fid = %d %v\n", fm.activeAF.fid, err)
					}
				}
			}
		}()
	}
	go func() {
		for range time.Tick(time.Minute) {
			if time.Now().Unix()-fm.rot <= 300 {
				continue
			}
			err := func() error {
				defer func() {
					if err := recover(); err != nil {
						logger.L.Errorf("文件merge定时任务异常 %v\n", err)
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
						logger.L.Errorf("fid = %d merge failure, err = %v\n", fid.(int64), err)
					} else {
						fm.Remove(af.(*appendFile))
						fm.counter.Delete(fid)
						logger.L.Infof("fid = %d merge success\n", fid.(int64))
					}
					return true
				})
				return nil
			}()
			if err != nil {
				logger.L.Errorf("文件merge定时任务异常 %v\n", err)
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

// 用于数据文件同步
func (fm *AppendFileManager) WriteRaw(d []byte) error {
	i := bytes.Index(d, constant.EOF)
	if i == -1 {
		_, err := fm.activeAF.Write(d)
		if err != nil {
			return err
		}
		return nil
	}
	if i != 0 {
		_, err := fm.activeAF.Write(d[:i])
		if err != nil {
			return err
		}
	}
	fid := time.Now().UnixNano()
	af, err := NewAppendFile(constant.Dir, constant.Active, fid)
	if err != nil {
		return err
	}
	fm.afmap.Store(fid, af)
	oaf := fm.activeAF
	fm.olderAF = append(fm.olderAF, oaf)
	fm.activeAF = af
	oaf.SetOlder()
	meta.M.ActiveFid = fid
	meta.M.OlderFids = append(meta.M.OlderFids, oaf.fid)
	meta.M.Save()
	fm.loadtime = time.Now()
	err = fm.loadAppendFile(oaf)
	if err != nil {
		logger.L.Errorf("loadAppendFile fid = %d %v\n", fm.activeAF.fid, err)
	}
	if atomic.CompareAndSwapInt32(&fm.sistate, constant.Idle, constant.Running) {
		go func() { fm.IndexSave() }()
	}
	if len(d) > i+6 {
		_, err := fm.activeAF.Write(d[i+6:])
		if err != nil {
			return err
		}
	}
	return nil
}

func (fm *AppendFileManager) Read(k []byte) ([]byte, error) {
	fm.rot = time.Now().Unix()
	v, state := fm.index.Load(string(k))
	if state == false {
		return nil, constant.KeyNotFound
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
	if string(kv.Value) == constant.DeleteFlag {
		fm.index.Delete(string(kv.Value))
		return nil, constant.KeyNotFound
	}
	return kv.Value, nil
}

func (fm *AppendFileManager) Load() error {
	startTime := time.Now()
	logger.L.Infof("开始加载索引")
	if constant.InvalidIndex {
		if meta.M.OlderFids != nil {
			sort.Sort(i64(meta.M.OlderFids))
			for _, fid := range meta.M.OlderFids {
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
	if meta.M.ActiveFid != 0 {
		af, _ := fm.afmap.Load(meta.M.ActiveFid)
		err := fm.loadAppendFile(af.(*appendFile))
		if err != nil {
			return err
		}
	}
	logger.L.Infof("加载索引完成，耗时 %.2f 秒\n", time.Since(startTime).Seconds())
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
		if err == io.EOF {
			return nil
		} else if err != nil {
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
	meta.M.Save()
}

func (fm *AppendFileManager) IndexSave() {
	defer func() {
		atomic.CompareAndSwapInt32(&fm.sistate, constant.Running, constant.Idle)
	}()
	startTime := time.Now()
	defer func() {
		logger.L.Infof("index save 耗时: %.2f 秒\n", time.Since(startTime).Seconds())
	}()
	fn := filepath.Join(constant.Dir, "idx")
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
		constant.ByteOrder.PutUint16(b[0:2], uint16(kl+22))
		constant.ByteOrder.PutUint64(b[2:10], uint64(i.fid))
		constant.ByteOrder.PutUint64(b[10:18], uint64(i.offset))
		constant.ByteOrder.PutUint32(b[18:22], uint32(i.size))
		copy(b[22:kl+22], k)
		_, err := f.Write(b)
		if err != nil {
			logger.L.Errorf("index save key = %s, item = %v, err = %v\n", k, i, err)
		}
		return true
	})
	metrics.ObjectCurrentCount.Set(float64(os))
}

func (fm *AppendFileManager) IndexLoad() error {
	fn := filepath.Join(constant.Dir, "idx")
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
		s := constant.ByteOrder.Uint16(b)
		d := make([]byte, s)
		n, err = f.ReadAt(d, off)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		item := &Item{
			fid:    int64(constant.ByteOrder.Uint64(d[2:10])),
			offset: int64(constant.ByteOrder.Uint64(d[10:18])),
			size:   int32(constant.ByteOrder.Uint32(d[18:22])),
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
				if string(kv.Value) != constant.DeleteFlag {
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
	if af.role != constant.Older {
		return
	}
	fm.afmap.Delete(af.fid)
	af.Close()
	os.Remove(af.fn)
	ofids := make([]int64, 0)
	for _, fid := range meta.M.OlderFids {
		if fid != af.fid {
			ofids = append(ofids, fid)
		}
	}
	meta.M.OlderFids = ofids
	meta.M.Save()
}
