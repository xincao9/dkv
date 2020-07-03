package store

import (
	"crypto/md5"
	"dkv/component/constant"
	"dkv/component/logger"
	"dkv/store/appendfile"
	"encoding/hex"
	"math"
	"sync"
)

var (
	S *Store
)

// 启动存储引擎
func init() {
	var err error
	S, err = NewStore()
	if err != nil {
		logger.L.Fatalf("Fatal error store: %v\n", err)
	}
}

type (
	kv struct {
		k   []byte
		v   []byte
		err error
	}
	ops struct {
		kv   *kv
		resp chan bool
	}
	kvObjectPool struct {
		pool *sync.Pool
	}
	Store struct {
		FM       *appendfile.AppendFileManager
		rop      chan *ops
		wop      chan *ops
		shutdown chan bool
		kvop     *kvObjectPool
		sequence bool
	}
)

func NewStore() (*Store, error) {
	fm, err := appendfile.NewAppendFileManager()
	if err != nil {
		return nil, err
	}
	if constant.Sequence == false {
		return &Store{
			FM:       fm,
			sequence: constant.Sequence,
		}, nil
	}
	rop := make(chan *ops)
	wop := make(chan *ops)
	shutdown := make(chan bool)
	go func() {
		for {
			select {
			case r := <-rop:
				{
					r.kv.v, r.kv.err = fm.Read(r.kv.k)
					r.resp <- true
				}
			case w := <-wop:
				{
					w.kv.err = fm.Write(w.kv.k, w.kv.v)
					w.resp <- true
				}
			case <-shutdown:
				{
					return
				}
			}
		}
	}()
	return &Store{
		FM:       fm,
		rop:      rop,
		wop:      wop,
		shutdown: shutdown,
		kvop:     newOpsObjectPool(),
		sequence: constant.Sequence,
	}, nil
}

func (s *Store) Get(k []byte) ([]byte, error) {
	if len(k) >= math.MaxUint8 {
		h := md5.New()
		k = []byte(hex.EncodeToString(h.Sum(k)))
	}
	if s.sequence {
		r := s.kvop.get()
		r.kv.k = k
		defer s.kvop.put(r)
		s.rop <- r
		<-r.resp
		return r.kv.v, r.kv.err
	}
	return s.FM.Read(k)
}

func (s *Store) Put(k, v []byte) error {
	if len(k) >= math.MaxUint8 {
		h := md5.New()
		k = []byte(hex.EncodeToString(h.Sum(k)))
	}
	if s.sequence {
		w := s.kvop.get()
		w.kv.k = k
		w.kv.v = v
		defer s.kvop.put(w)
		s.wop <- w
		<-w.resp
		return w.kv.err
	}
	return s.FM.Write(k, v)
}

func (s *Store) Delete(k []byte) error {
	_, err := s.Get(k)
	if err != nil {
		return err
	}
	return s.Put(k, []byte(constant.DeleteFlag))
}

func (s *Store) Close() {
	if s.sequence {
		s.shutdown <- true
	}
	s.FM.Close()
}

func newOpsObjectPool() *kvObjectPool {
	pool := &sync.Pool{New: func() interface{} {
		return &ops{
			kv: &kv{
				k:   nil,
				v:   nil,
				err: nil,
			},
			resp: make(chan bool),
		}
	}}
	return &kvObjectPool{pool: pool}
}

func (k *kvObjectPool) get() *ops {
	v := k.pool.Get()
	return v.(*ops)
}

func (k *kvObjectPool) put(o *ops) {
	o.kv.k = nil
	o.kv.v = nil
	o.kv.err = nil
	k.pool.Put(o)
}
