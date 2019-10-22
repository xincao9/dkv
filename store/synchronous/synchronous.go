package synchronous

import (
	"dkv/config"
	"dkv/logger"
	"dkv/store"
	"dkv/store/appendfile"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	Master          = 1
	Slave           = 2
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
	D   *Synchronous
	EOF = []byte("CYZEOF")
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
	if role != Master && role != Slave {
		return nil, nil
	}
	s := &Synchronous{}
	s.role = role
	if role == Master {
		ln, err := net.Listen("tcp", config.D.GetString("ms.m.port"))
		if err != nil {
			return nil, err
		}
		logger.D.Infof("Synchronous new listen port: %s\n", config.D.GetString("ms.m.port"))
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
							time.Sleep(time.Second * 10)
							s.handler(host)
						}
					}
				}(conn)
			}
		}()
		return s, nil
	}
	conn, err := net.Dial("tcp", config.D.GetString("ms.s.addr"))
	if err != nil {
		return nil, err
	}
	go func(c net.Conn) {
		logger.D.Infof("Synchronous new connection addr: %s\n", config.D.GetString("ms.s.addr"))
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
		logger.D.Infof("Synchronous new close connection addr: %s\n", config.D.GetString("ms.s.addr"))
		c.Close()
	}(conn)
	return s, nil
}

func saveSlaveInfo (host string, sI *slaveInfo) {
	val, _ := json.Marshal(&sI)
	err := store.D.Put([]byte(fmt.Sprintf(slaveInfoSuffix, host)), val)
	if err != nil {
		logger.D.Errorf("Synchronous saveSlaveInfo store put host: %s, slaveInfo = %v\n", host, sI)
	}
}

func getSalveInfoByHost (host string) (*slaveInfo, error) {
	var sI slaveInfo
	val, err := store.D.Get([]byte(fmt.Sprintf(slaveInfoSuffix, host)))
	if err != nil {
		logger.D.Errorf("Synchronous getSalveInfoByHost store get: %v\n", err)
		return &sI, err
	}
	err = json.Unmarshal(val, &sI)
	if err != nil {
		logger.D.Errorf("Synchronous getSalveInfoByHost store get unmarshal: %v\n", err)
		return &sI, err
	}
	return &sI, nil
}

func (s *Synchronous) handler(host string) {
	c, _ := s.conns.Load(host)
	conn := c.(net.Conn)
	sI, err := getSalveInfoByHost(host)
	if err != nil && err != appendfile.KeyNotFound {
		s.close(host)
		return
	}
	fns := store.D.GetAppendFiles()
	for _, fn := range fns {
		i := strings.LastIndex(fn, "/")
		if i == -1 || len(fn) <= i+1 {
			logger.D.Errorf("Synchronous handler fn: %s\n", fn)
			s.close(host)
			return
		}
		ofid, err := strconv.ParseInt(fn[i+1:], 10, 64)
		if err != nil {
			logger.D.Errorf("Synchronous handler fn: %s\n", fn)
			s.close(host)
			return
		}
		if sI.Fid != 0 && ofid < sI.Fid {
			continue
		}
		if ofid > sI.Fid {
			sI.Off = 0
		}
		f, err := os.OpenFile(fn, os.O_RDONLY, 0644)
		if err != nil {
			logger.D.Errorf("Synchronous handler openfile fn: %s, err: %v\n", fn, err)
			continue
		}
		fi, err := f.Stat()
		if err == nil {
			if ofid == sI.Fid && fi.Size() <= sI.Off {
				continue
			}
		}
		sI.Fid = ofid
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
				saveSlaveInfo(host, sI)
				conn.Write(EOF)
				logger.D.Infof("Synchronous handler write fn = %s finish\n", fn)
				f.Close()
				break
			} else if err != nil {
				saveSlaveInfo(host, sI)
				conn.Write(EOF)
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
