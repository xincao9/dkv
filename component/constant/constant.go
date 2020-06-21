package constant

import (
	"dkv/component/config"
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
	Dir          = config.C.GetString("data.dir")
	InvalidIndex = config.C.GetBool("data.invalidIndex")
	Compress     = config.C.GetBool("data.compress")
	Cache        = config.C.GetBool("data.cache")
	CacheSize    = config.C.GetInt("data.cache.size")
	Mode         = config.C.GetString("server.mode")
	Port         = config.C.GetInt("server.port")
	RedisPort    = config.C.GetInt("server.redis.port")
	Sequence     = config.C.GetBool("server.sequence")
	LoggerLevel  = config.C.GetString("logger.level")
	LoggerDir    = config.C.GetString("logger.dir")
	MSRole       = config.C.GetInt("ms.role")
	MSMPort      = config.C.GetInt("ms.m.port")
	MSSAddr      = config.C.GetString("ms.s.addr")
	KeyNotFound  = errors.New("key is not found")
)
