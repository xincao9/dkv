package cache

import (
	"dkv/store/appendfile"
	"github.com/VictoriaMetrics/fastcache"
	"path/filepath"
)

var (
	C *fastcache.Cache
)

const open bool = true

func init() {
	C = fastcache.New (1 << 30)
}

func Get (key []byte) []byte{
	if open == false {
		return nil
	}
	return C.Get(nil, key)
}

func Set (key []byte, val []byte) {
	if open == false {
		return
	}
	C.Set(key, val)
}

func Del (key []byte) {
	if open == false {
		return
	}
	C.Del(key)
}

func Close () {
	C.SaveToFile(filepath.Join(appendfile.M.Dir, "cache"))
}
