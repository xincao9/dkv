package synchronous

import (
	"bufio"
	"dkv/config"
	"dkv/logger"
	"dkv/store"
	"dkv/store/appendfile"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	master          = 1
	slave           = 2
	slaveInfoSuffix = "%s.slaveInfoSuffix"
)

type slaveInfo struct {
	Fid int64 `json:"fid"`
	Off int64 `json:"off"`
}

type Synchronous struct {
	role  int
	conns sync.Map
}

var (
	D *Synchronous
)

func init() {
	var err error
	D, err = New()
	if err != nil {
		logger.D.Fatalf("Fatal error synchronous: %v\n", err)
	}
}

func New() (*Synchronous, error) {
	role := config.D.GetInt("ms.role")
	if role != master && role != slave {
		return nil, nil
	}
	s := &Synchronous{}
	s.role = role
	if role == master {
		ln, err := net.Listen("tcp", config.D.GetString("ms.m.port"))
		if err != nil {
			return nil, err
		}
		logger.D.Infof("Synchronous new listen port: %s\n", config.D.GetString("ms.m.port"))
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					logger.D.Errorf("Synchronous new accept %v\n", err)
					continue
				}
				go func(c net.Conn) {
					addr := c.RemoteAddr().String()
					logger.D.Infof("Synchronous new handler %s\n", addr)
					s.conns.Store(addr, conn)
					state := true
					for state {
						_, state = s.conns.Load(addr)
						if state {
							s.handler(addr)
						}
					}
				}(conn)
			}
		}()
	} else {
		conn, err := net.Dial("tcp", config.D.GetString("ms.s.addr"))
		if err != nil {
			return nil, err
		}
		go func(c net.Conn) {
			logger.D.Infof("Synchronous new conn addr: %s\n", config.D.GetString("ms.s.addr"))
			fn := filepath.Join(config.D.GetString("data.dir"), strconv.FormatInt(time.Now().UnixNano(), 10))
			f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				logger.D.Errorf("Synchronous handler fn = %s, err = %v\n", fn, err)
				c.Close()
				return
			}
			scanner := bufio.NewScanner(c)
			for scanner.Scan() {

				_, err = f.Write(scanner.Bytes())
				if err != nil {
					logger.D.Errorf("Synchronous handler fn = %s, err = %v\n", fn, err)
				}
			}
			logger.D.Infof("Synchronous new close conn addr: %s\n", config.D.GetString("ms.s.addr"))
			c.Close()
		}(conn)
	}
	return s, nil
}

func (s *Synchronous) handler(addr string) {
	c, state := s.conns.Load(addr)
	if state == false {
		logger.D.Errorf("Synchronous handler addr = %s not found conn\n", addr)
		return
	}
	conn := c.(net.Conn)
	var sI slaveInfo
	val, err := store.D.Get([]byte(fmt.Sprintf(slaveInfoSuffix, addr)))
	if err == nil {
		err = json.Unmarshal(val, &sI)
		if err != nil {
			logger.D.Errorf("Synchronous handler %v\n", err)
			s.close(addr)
			return
		}
	} else if err != appendfile.KeyNotFound {
		logger.D.Errorf("Synchronous handler %v\n", err)
		s.close(addr)
		return
	}
	fns := store.D.GetAppendFiles()
	for _, fn := range fns {
		i := strings.LastIndex(fn, "/")
		if i == -1 || len(fn) <= i+1 {
			logger.D.Errorf("Synchronous handler fn = %s\n", fn)
			s.close(addr)
			return
		}
		ofid, err := strconv.ParseInt(fn[i+1:], 10, 64)
		if err != nil {
			logger.D.Errorf("Synchronous handler fn = %s\n", fn)
			s.close(addr)
			return
		}
		if sI.Fid != 0 && ofid < sI.Fid {
			continue
		}
		f, err := os.OpenFile(fn, os.O_RDONLY, 0644)
		if err != nil {
			logger.D.Errorf("Synchronous handler fn = %s, err = %v\n", fn, err)
			continue
		}
		b := make([]byte, 1024)
		sI.Fid = ofid
		sI.Off = 0
		for {
			n, err := f.ReadAt(b, sI.Off)
			if n > 0 {
				sI.Off = sI.Off + int64(n)
				val, _ = json.Marshal(&sI)
				err = store.D.Put([]byte(fmt.Sprintf(slaveInfoSuffix, addr)), val)
				if err != nil {
					logger.D.Errorf("Synchronous handler fn = %s, err = %v\n", fn, err)
				}
				_, err = conn.Write(b[:n])
				if err != nil {
					logger.D.Errorf("Synchronous handler fn = %s, err = %v\n", fn, err)
					s.close(addr)
				}
			}
			if err == io.EOF {
				logger.D.Infof("Synchronous handler fn = %s finish\n", fn)
				f.Close()
				break
			} else if err != nil {
				logger.D.Errorf("Synchronous handler fn = %s, err = %v\n", fn, err)
				f.Close()
				break
			}
		}
	}
}

func (s *Synchronous) close(addr string) {
	val, state := s.conns.Load(addr)
	if state {
		val.(net.Conn).Close()
		s.conns.Delete(addr)
	}
}

// 主节点广播
func (s *Synchronous) Broadcast(b []byte) {
	s.conns.Range(func(addr, value interface{}) bool {
		_, err := value.(net.Conn).Write(b)
		if err != nil {
			logger.D.Errorf("Synchronous Broadcast addr = %s, err = %v\n", addr, err)
		}
		return true
	})
}
