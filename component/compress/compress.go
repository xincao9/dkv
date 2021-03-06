package compress

import (
	"dkv/component/constant"
	"github.com/golang/snappy"
)

var (
	C *compress
)

func init() {
	C = new(constant.CompressOpen)
}

func new(open bool) *compress {
	return &compress{
		open: open,
	}
}

type compress struct {
	open bool
}

func (c *compress) Encode(d []byte) []byte {
	if c.open == false {
		return d
	}
	return snappy.Encode(nil, d)
}

func (c *compress) Decode(d []byte) []byte {
	if c.open == false {
		return d
	}
	v, err := snappy.Decode(nil, d)
	if err != nil {
		return d
	}
	return v
}
