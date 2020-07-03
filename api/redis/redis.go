package redis

import (
	"dkv/component/constant"
	"dkv/component/logger"
	"dkv/store"
	"fmt"
	"github.com/tidwall/redcon"
	"strings"
)

func run() {
	addr := fmt.Sprintf(":%d", constant.RedisPort)
	logger.L.Infof("Redis listening and serving on : %s", addr)
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
				err := store.S.Put(cmd.Args[1], cmd.Args[2])
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
				val, err := store.S.Get(cmd.Args[1])
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
				err := store.S.Delete(cmd.Args[1])
				if err != nil {
					conn.WriteInt(0)
					return
				}
				conn.WriteInt(1)
			}
		},
		func(conn redcon.Conn) bool {
			// use this function to accept or deny the connection.
			logger.L.Infof("accept: %s\n", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// this is called when the connection has been closed
			logger.L.Infof("closed: %s, err: %v\n", conn.RemoteAddr(), err)
		},
	)
	if err != nil {
		logger.L.Fatalf("Fatal error redis: %v\n", err)
	}
}

func ListenAndServe() {
	go run()
}
