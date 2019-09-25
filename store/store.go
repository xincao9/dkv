package store

import "dkv/store/appendfile"

type readOp struct {
	key  []byte
	resp chan []byte
}

type writeOp struct {
	key  []byte
	val  []byte
	resp chan bool
}

type store struct {
	fm  *appendfile.FileManager
	rop chan *readOp
	wop chan *writeOp
}

func New(dir string) (*store, error) {
	fm, err := appendfile.NewFileManager(dir)
	if err != nil {
		return nil, err
	}
	rop := make(chan *readOp)
	wop := make(chan *writeOp)

	go func() {
		for {
			select {
			case r := <-rop:
				{
					val, _ := fm.Read(r.key)
					r.resp <- val
				}
			case w := <-wop:
				{
					fm.Write(w.key, w.val)
					w.resp <- true
				}
			}
		}
	}()
	return &store{fm: fm, rop: rop, wop: wop}, nil
}

func (s *store) Get(k []byte) ([]byte, error) {
	r := &readOp{
		key:  k,
		resp: make(chan []byte),
	}
	s.rop <- r
	return <-r.resp, nil
}

func (s *store) Put(k, v []byte) error {
	w := &writeOp{
		key:  k,
		val:  v,
		resp: make(chan bool),
	}
	s.wop <- w
	<-w.resp
	return nil
}
