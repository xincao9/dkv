package synchronous

import (
    "dkv/component/constant"
    "dkv/component/logger"
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
	role        int
	connections sync.Map
}

var (
	D *Synchronous
)

func init() {
	var err error
	D, err = New()
	if err != nil {
		logger.L.Fatalf("Fatal error synchronous: %v\n", err)
	}
}

func New() (*Synchronous, error) {
	role := constant.MSRole
	if role != constant.Master && role != constant.Slave {
		return nil, nil
	}
	s := &Synchronous{}
	s.role = role
	if role == constant.Master {
		addr := fmt.Sprintf(":%d", constant.MSMPort)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}
		logger.L.Infof("Synchronous new listen port: %s\n", addr)
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					logger.L.Errorf("Synchronous new accept: %v\n", err)
					continue
				}
				go func(c net.Conn) {
					i := strings.LastIndex(c.RemoteAddr().String(), ":")
					host := string([]byte(c.RemoteAddr().String())[:i])
					logger.L.Infof("Synchronous new handler host: %s\n", host)
					s.connections.Store(host, conn)
					state := true
					for state {
						_, state = s.connections.Load(host)
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
	addr := constant.MSSAddr
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	go func(c net.Conn) {
		logger.L.Infof("Synchronous new connection addr: %s\n", addr)
		b := make([]byte, 1024)
		for {
			n, err := c.Read(b)
			if err != nil {
				netErr, ok := err.(net.Error)
				if ok && netErr.Timeout() {
					continue
				}
				logger.L.Errorf("Synchronous new read: %v\n", err)
				break
			}
			err = store.S.FM.WriteRaw(b[:n])
			if err != nil {
				logger.L.Errorf("Synchronous new store write : %v\n", err)
			}
		}
		logger.L.Infof("Synchronous new close connection addr: %s\n", addr)
		c.Close()
	}(conn)
	return s, nil
}

func (s *Synchronous) handler(host string) {
	c, _ := s.connections.Load(host)
	conn := c.(net.Conn)
	sI, state := meta.M.GetSalveInfoByHost(host)
	var cFid, cOff int64
	if state {
		cFid = sI.Fid
		cOff = sI.Off
	}
	fids := meta.M.GetFids()
	for _, fid := range fids {
		fn := filepath.Join(constant.Dir, strconv.FormatInt(fid, 10))
		if fid < cFid {
			continue
		}
		if fid > cFid {
			cOff = 0
			if cFid != 0 {
				conn.Write(constant.EOF)
			}
		}
		fo, err := os.OpenFile(fn, os.O_RDONLY, 0644)
		if err != nil {
			logger.L.Errorf("Synchronous handler openfile fn: %s, err: %v\n", fn, err)
			continue
		}
		fi, err := fo.Stat()
		if err == nil {
			if fid == cFid && fi.Size() <= cOff {
				fo.Close()
				continue
			}
		}
		cFid = fid
		b := make([]byte, 1024)
		for {
			n, err := fo.ReadAt(b, cOff)
			if n > 0 {
				cOff = cOff + int64(n)
				_, err = conn.Write(b[:n])
				if err != nil {
					logger.L.Errorf("Synchronous handler connection write fn = %s, err = %v\n", fn, err)
					s.close(host)
				}
			}
			if err == io.EOF {
				meta.M.SaveSlaveInfo(host, cFid, cOff)
				logger.L.Infof("Synchronous handler write fn = %s finish\n", fn)
				fo.Close()
				break
			} else if err != nil {
				meta.M.SaveSlaveInfo(host, cFid, cOff)
				logger.L.Errorf("Synchronous handler read fn = %s, err = %v\n", fn, err)
				fo.Close()
				break
			}
		}
	}
}

func (s *Synchronous) close(addr string) {
	val, state := s.connections.Load(addr)
	if state {
		val.(net.Conn).Close()
		s.connections.Delete(addr)
	}
}
