package store

import (
	"crypto/md5"
	"dkv/config"
	"dkv/logger"
	"dkv/store/appendfile"
	"encoding/hex"
	"math"
)

var (
	D *Store
)

func init() {
	// 启动存储引擎
	var err error
	D, err = NewStore(config.D.GetString("data.dir"))
	if err != nil {
		logger.D.Fatalf("Fatal error store: %v\n", err)
	}
}

type KV struct {
	K   []byte
	V   []byte
	Err error
}

type ROps struct {
	kv   *KV
	resp chan bool
}

type WOps struct {
	kv   *KV
	resp chan bool
}

type Store struct {
	fm       *appendfile.AppendFileManager
	rop      chan *ROps
	wop      chan *WOps
	shutdown chan bool
}

var (
	sequence bool
)

func NewStore(dir string) (*Store, error) {
	sequence = config.D.GetBool("server.sequence")
	fm, err := appendfile.NewAppendFileManager(dir)
	if err != nil {
		return nil, err
	}
	rop := make(chan *ROps)
	wop := make(chan *WOps)
	shutdown := make(chan bool)
	if sequence {
		go func() {
			for {
				select {
				case r := <-rop:
					{
						r.kv.V, r.kv.Err = fm.Read(r.kv.K)
						r.resp <- true
					}
				case w := <-wop:
					{
						w.kv.Err = fm.Write(w.kv.K, w.kv.V)
						w.resp <- true
					}
				case <-shutdown:
					{
						return
					}
				}
			}
		}()
	}
	return &Store{fm: fm, rop: rop, wop: wop, shutdown: shutdown}, nil
}

func (s *Store) Get(k []byte) ([]byte, error) {
	if len(k) >= math.MaxUint8 {
		h := md5.New()
		k = []byte(hex.EncodeToString(h.Sum(k)))
	}
	if sequence {
		r := &ROps{
			kv:   &KV{K: k},
			resp: make(chan bool),
		}
		s.rop <- r
		<-r.resp
		return r.kv.V, r.kv.Err
	}
	return s.fm.Read(k)
}

func (s *Store) Put(k, v []byte) error {
	if len(k) >= math.MaxUint8 {
		h := md5.New()
		k = []byte(hex.EncodeToString(h.Sum(k)))
	}
	if sequence {
		w := &WOps{
			kv:   &KV{K: k, V: v},
			resp: make(chan bool),
		}
		s.wop <- w
		<-w.resp
		return w.kv.Err
	}
	return s.fm.Write(k, v)
}

func (s *Store) Delete(k []byte) error {
	_, err := s.Get(k)
	if err != nil {
		return err
	}
	return s.Put(k, []byte(appendfile.DeleteFlag))
}

// 用于数据文件同步
func (s *Store) WriteRaw (d []byte) error {
	return s.fm.WriteRaw(d)
}

func (s *Store) Close() {
	s.shutdown <- true
	s.fm.Close()
}

func (s *Store) GetAppendFiles() []string {
	return s.fm.GetAppendFiles()
}