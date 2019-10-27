package constant

import (
	"dkv/config"
	"encoding/binary"
	"errors"
)

const (
	Older        = 1
	Active       = 2
	DeleteFlag   = "d#f"
	MetaFn       = "meta.json"
	DefaultDir   = "/tmp/dkv"
	Master       = 1
	Slave        = 2
	Idle         = 0
	Running      = 1
	MaxValueSize = 1 << 26 // 64M
)

var (
	EOF          = []byte("E#O#F")
	ByteOrder    = binary.BigEndian
	Dir          = config.D.GetString("data.dir")
	InvalidIndex = config.D.GetBool("data.invalidIndex")
	Compress     = config.D.GetBool("data.compress")
	Cache        = config.D.GetBool("data.cache")
	Mode         = config.D.GetString("server.mode")
	Port         = config.D.GetInt("server.port")
	RedisPort    = config.D.GetInt("server.redis.port")
	Sequence     = config.D.GetBool("server.sequence")
	LoggerLevel  = config.D.GetString("logger.level")
	MSRole       = config.D.GetInt("ms.role")
	MSMPort      = config.D.GetInt("ms.m.port")
	MSSAddr      = config.D.GetString("ms.s.addr")
	KeyNotFound  = errors.New("key is not found")
)
