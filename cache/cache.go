package cache

import (
	"dkv/constant"
	"github.com/VictoriaMetrics/fastcache"
	"path/filepath"
)

var (
	C    *fastcache.Cache
	open bool
)

func init() {
	C = fastcache.LoadFromFileOrNew(filepath.Join(constant.Dir, "cache"), 1<<30)
	open = constant.Cache
}

func Get(key []byte) []byte {
	if open == false {
		return nil
	}
	return C.Get(nil, key)
}

func Set(key []byte, val []byte) {
	if open == false {
		return
	}
	C.Set(key, val)
}

func Del(key []byte) {
	if open == false {
		return
	}
	C.Del(key)
}

func Close() {
	C.SaveToFile(filepath.Join(constant.Dir, "cache"))
}
