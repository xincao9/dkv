package store

import "dkv/store/appendfile"

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

type store struct {
	fm       *appendfile.FileManager
	rop      chan *ROps
	wop      chan *WOps
	shutdown chan bool
}

const (
	sequence = false
)

func New(dir string) (*store, error) {
	fm, err := appendfile.NewFileManager(dir)
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
	return &store{fm: fm, rop: rop, wop: wop, shutdown: shutdown}, nil
}

func (s *store) Get(k []byte) ([]byte, error) {
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

func (s *store) Put(k, v []byte) error {
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

func (s *store) Delete(k []byte) error {
	_, err := s.Get(k)
	if err != nil {
		return err
	}
	return s.Put(k, []byte(appendfile.DeleteFlag))
}

func (s *store) Close() {
	s.shutdown <- true
	s.fm.Close()
}
