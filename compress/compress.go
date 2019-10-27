package compress

import (
	"dkv/constant"
	"github.com/golang/snappy"
)

func Encode(d []byte) []byte {
	if constant.Compress == false {
		return d
	}
	return snappyEncode(d)
}

func Decode(d []byte) []byte {
	if constant.Compress == false {
		return d
	}
	return snappyDecode(d)
}

func snappyEncode(d []byte) []byte {
	return snappy.Encode(nil, d)
}

func snappyDecode(d []byte) []byte {
	v, err := snappy.Decode(nil, d)
	if err != nil {
		return d
	}
	return v
}
