package synchronous

import (
	"dkv/config"
	"dkv/logger"
	"dkv/store"
	"dkv/store/meta"
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
	if role != meta.Master && role != meta.Slave {
		return nil, nil
	}
	s := &Synchronous{}
	s.role = role
	if role == meta.Master {
		addr := fmt.Sprintf(":%s", config.D.GetString("ms.m.port"))
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}
		logger.D.Infof("Synchronous new listen port: %s\n", addr)
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					logger.D.Errorf("Synchronous new accept: %v\n", err)
					continue
				}
				go func(c net.Conn) {
					i := strings.LastIndex(c.RemoteAddr().String(), ":")
					host := string([]byte(c.RemoteAddr().String())[:i])
					logger.D.Infof("Synchronous new handler host: %s\n", host)
					s.conns.Store(host, conn)
					state := true
					for state {
						_, state = s.conns.Load(host)
						if state {
							time.Sleep(time.Second * 3)
							s.handler(host)
						}
					}
				}(conn)
			}
		}()
		return s, nil
	}
	addr := config.D.GetString("ms.s.addr")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	go func(c net.Conn) {
		logger.D.Infof("Synchronous new connection addr: %s\n", addr)
		b := make([]byte, 1024)
		for {
			n, err := c.Read(b)
			if err != nil {
				netErr, ok := err.(net.Error)
				if ok && netErr.Timeout() {
					continue
				}
				logger.D.Errorf("Synchronous new read: %v\n", err)
				break
			}
			err = store.D.WriteRaw(b[:n])
			if err != nil {
				logger.D.Errorf("Synchronous new store write : %v\n", err)
			}
		}
		logger.D.Infof("Synchronous new close connection addr: %s\n", addr)
		c.Close()
	}(conn)
	return s, nil
}

func (s *Synchronous) handler(host string) {
	c, _ := s.conns.Load(host)
	conn := c.(net.Conn)
	sI, state := store.D.FM.Meta.GetSalveInfoByHost(host)
	if state == false {
		sI = &meta.SlaveInfo{
			Fid: 0,
			Off: 0,
		}
	}
	fids := store.D.FM.GetFids()
	for _, fid := range fids {
		fn := filepath.Join(store.D.FM.Meta.Dir, strconv.FormatInt(fid, 10))
		if fid < sI.Fid {
			continue
		}
		if fid > sI.Fid {
			sI.Off = 0
			if sI.Fid != 0 {
				conn.Write(meta.EOF)
			}
		}
		f, err := os.OpenFile(fn, os.O_RDONLY, 0644)
		if err != nil {
			logger.D.Errorf("Synchronous handler openfile fn: %s, err: %v\n", fn, err)
			continue
		}
		fi, err := f.Stat()
		if err == nil {
			if fid == sI.Fid && fi.Size() <= sI.Off {
				f.Close()
				continue
			}
		}
		sI.Fid = fid
		b := make([]byte, 1024)
		for {
			n, err := f.ReadAt(b, sI.Off)
			if n > 0 {
				sI.Off = sI.Off + int64(n)
				_, err = conn.Write(b[:n])
				if err != nil {
					logger.D.Errorf("Synchronous handler connection write fn = %s, err = %v\n", fn, err)
					s.close(host)
				}
			}
			if err == io.EOF {
				store.D.FM.Meta.SaveSlaveInfo(host, sI)
				logger.D.Infof("Synchronous handler write fn = %s finish\n", fn)
				f.Close()
				break
			} else if err != nil {
				store.D.FM.Meta.SaveSlaveInfo(host, sI)
				logger.D.Errorf("Synchronous handler read fn = %s, err = %v\n", fn, err)
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
