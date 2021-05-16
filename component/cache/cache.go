package cache

import (
	"dkv/component/constant"
	"github.com/VictoriaMetrics/fastcache"
	"path/filepath"
)

const (
	fn = "cache"
)

var (
	C *cache
)

func init() {
	C = new(filepath.Join(constant.Dir, fn), constant.CacheOpen, constant.CacheSize)
}

type cache struct {
	file     string
	c        *fastcache.Cache
	open     bool
	maxBytes int
}

func new(file string, open bool, maxBytes int) *cache {
	fc := fastcache.LoadFromFileOrNew(file, maxBytes)
	return &cache{
		file:     file,
		c:        fc,
		open:     open,
		maxBytes: maxBytes,
	}
}
func (c *cache) Get(key []byte) []byte {
	if c.open == false {
		return nil
	}
	return c.c.Get(nil, key)
}

func (c *cache) Set(key []byte, val []byte) {
	if c.open == false {
		return
	}
	c.c.Set(key, val)
}

func (c *cache) Del(key []byte) {
	if c.open == false {
		return
	}
	c.c.Del(key)
}

func (c *cache) Close() {
	if c.open == false {
		return
	}
	c.c.SaveToFile(c.file)
}
