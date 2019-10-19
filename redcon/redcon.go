package redcon

import (
	"dkv/config"
	"dkv/logger"
	"dkv/store"
	"github.com/tidwall/redcon"
	"log"
	"strings"
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
				store.D.Put(cmd.Args[1], cmd.Args[2])
				conn.WriteString("OK")
			case "get":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				val, err := store.D.Get(cmd.Args[1])
				if err == nil {
					conn.WriteBulk(val)
				} else {
					conn.WriteNull()
				}
			case "del":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				err := store.D.Delete(cmd.Args[1])
				if err != nil {
					conn.WriteInt(0)
				} else {
					conn.WriteInt(1)
				}
			}
		},
		func(conn redcon.Conn) bool {
			// use this function to accept or deny the connection.
			// log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// this is called when the connection has been closed
			// log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		},
	)
	if err != nil {
		log.Fatalf("fatal err redisconn : %v\n", err)
	}
}

func ListenAndServe() {
	go run()
}
