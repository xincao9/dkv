package cache

import (
	"dkv/constant"
	"github.com/VictoriaMetrics/fastcache"
	"path/filepath"
)

var (
	C *cache
)

func init() {
	C = new(filepath.Join(constant.Dir, "cache"), constant.Cache)
}

type cache struct {
	file string
	c    *fastcache.Cache
	open bool
}

func new(file string, open bool) *cache {
	fc := fastcache.LoadFromFileOrNew(file, 1<<30)
	return &cache{
		file: file,
		c:    fc,
		open: open,
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
