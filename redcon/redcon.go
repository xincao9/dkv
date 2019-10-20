package redcon

import (
	"dkv/config"
	"dkv/logger"
	"dkv/store"
	"dkv/store/appendfile"
	"encoding/json"
	"fmt"
	"github.com/tidwall/redcon"
	"io"
	"os"
	"strconv"
	"strings"
)

type SlaveInfo struct {
	Fid int64 `fid:"fid"`
	off int64 `off:"off"`
}

const (
	slaveInfoSuffix = "%s.slaveInfoSuffix"
)

func run() {
	addr := config.D.GetString("server.redcon.port")
	logger.D.Infof("Redconn listening and serving HTTP on : %s", addr)
	err := redcon.ListenAndServe(addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			switch strings.ToLower(string(cmd.Args[0])) {
			default:
				conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
			case "ping":
				conn.WriteString("PONG")
			case "quit":
				conn.WriteString("OK")
				conn.Close()
			case "set":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				err := store.D.Put(cmd.Args[1], cmd.Args[2])
				if err != nil {
					conn.WriteNull()
					return
				}
				conn.WriteString("OK")
			case "get":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				val, err := store.D.Get(cmd.Args[1])
				if err == nil {
					conn.WriteBulk(val)
					return
				}
				conn.WriteNull()
			case "del":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				err := store.D.Delete(cmd.Args[1])
				if err != nil {
					conn.WriteInt(0)
					return
				}
				conn.WriteInt(1)
			case "sync":
				if len(cmd.Args) > 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				var fid int64
				var off int64
				var err error
				var salveInfo SlaveInfo
				if len(cmd.Args) == 2 {
					fid, err = strconv.ParseInt(string(cmd.Args[1]), 10, 64)
					if err != nil {
						conn.WriteNull()
						return
					}
					off, err = strconv.ParseInt(string(cmd.Args[2]), 10, 64)
					if err != nil {
						conn.WriteNull()
						return
					}
				} else {
					val, err := store.D.Get([]byte(fmt.Sprintf(slaveInfoSuffix, conn.RemoteAddr())))
					if err == nil {
						err = json.Unmarshal(val, &salveInfo)
						if err != nil {
							conn.WriteNull()
							return
						}
					} else if err != appendfile.KeyNotFound {
						conn.WriteNull()
						return
					}
				}
				if len(cmd.Args) == 0 {
					fid = salveInfo.Fid
					off = salveInfo.off
				} else if len(cmd.Args) == 1 {
					fid, err = strconv.ParseInt(string(cmd.Args[1]), 10, 64)
					if err != nil {
						conn.WriteNull()
						return
					}
					off = salveInfo.off
				}
				fns := store.D.GetAppendFiles()
				start := false
				for _, fn := range fns {
					i := strings.LastIndex(fn, "/")
					if i == -1 || len(fn) <= i+1 {
						logger.D.Errorf("redcon fn = %s\n", fn)
						break
					}
					ofid, err := strconv.ParseInt(fn[i+1:], 10, 64)
					if err != nil {
						logger.D.Errorf("redcon fn = %s\n", fn)
						break
					}
					if fid == 0 || ofid >= fid {
						start = true
					}
					if start {
						f, err := os.OpenFile(fn, os.O_RDONLY, 0644)
						if err != nil {
							logger.D.Errorf("redcon fn = %s, err = %v\n", fn, err)
							continue
						}
						b := make([]byte, 1024)
						for {
							f.Seek(off, 0)
							n, err := f.Read(b)
							if err == io.EOF {
								if n > 0 {
									conn.WriteRaw(b[:n])
								}
								break
							} else if err != nil {
								logger.D.Errorf("redcon fn = %s, err = %v\n", fn, err)
								break
							}
							if n > 0 {
								conn.WriteRaw(b[:n])
							}
						}
					}
				}
				if err != nil {
					conn.WriteNull()
					return
				}
			}
		},
		func(conn redcon.Conn) bool {
			// use this function to accept or deny the connection.
			logger.D.Infof("accept: %s\n", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// this is called when the connection has been closed
			logger.D.Infof("closed: %s, err: %v\n", conn.RemoteAddr(), err)
		},
	)
	if err != nil {
		logger.D.Fatalf("fatal err redisconn : %v\n", err)
	}
}

func ListenAndServe() {
	go run()
}
