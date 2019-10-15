package compress

import (
	"bytes"
	"compress/zlib"
	"dkv/config"
	"github.com/golang/snappy"
	"io"
)

func Encode(d []byte) []byte {
	if config.D.GetBool("data.compress") == false {
		return d
	}
	return snappyEncode(d)
}

func Decode(d []byte) []byte {
	if config.D.GetBool("data.compress") == false {
		return d
	}
	return snappyDecode(d)
}

func zlibEncode(d []byte) []byte {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(d)
	w.Close()
	return b.Bytes()
}

func zlibDecode(d []byte) []byte {
	r, err := zlib.NewReader(bytes.NewReader(d))
	if err != nil {
		return d
	}
	var b bytes.Buffer
	io.Copy(&b, r)
	r.Close()
	return b.Bytes()
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
