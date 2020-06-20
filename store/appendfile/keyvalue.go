package appendfile

import (
	"dkv/component/constant"
	"errors"
	"hash/crc32"
	"math"
)

type keyValue struct {
	Crc32     uint32
	KeySize   uint8
	ValueSize uint32
	Key       []byte
	Value     []byte
}

func NewKeyValue(key []byte, value []byte) (*keyValue, error) {
	kl := len(key)
	vl := len(value)
	if kl > math.MaxUint8 || vl > constant.MaxValueSize {
		return nil, errors.New("key or value length exceeds maximum")
	}
	kv := &keyValue{
		KeySize:   uint8(kl),
		ValueSize: uint32(vl),
		Key:       key,
		Value:     value,
	}
	kv.Crc32 = crc32Checksum(*kv)
	return kv, nil
}

func crc32Checksum(kv keyValue) uint32 {
	kv.Crc32 = 0
	return crc32.ChecksumIEEE(Encode(&kv))
}

func Encode(kv *keyValue) []byte {
	b := make([]byte, int(kv.KeySize)+int(kv.ValueSize)+9)
	constant.ByteOrder.PutUint32(b[0:4], kv.Crc32)
	b[4] = kv.KeySize
	constant.ByteOrder.PutUint32(b[5:9], kv.ValueSize)
	s, e := 9, 9+int(kv.KeySize)
	copy(b[s:e], kv.Key)
	s, e = 9+int(kv.KeySize), 9+int(kv.KeySize)+int(kv.ValueSize)
	copy(b[s:e], kv.Value)
	return b
}

func Decode(b []byte) (*keyValue, error) {
	if len(b) < 9 {
		return nil, errors.New("packet header exception")
	}
	kv := &keyValue{
		Crc32:     constant.ByteOrder.Uint32(b[:4]),
		KeySize:   b[4],
		ValueSize: constant.ByteOrder.Uint32(b[5:9]),
	}
	s := int(kv.KeySize) + int(kv.ValueSize)
	if len(b) < s+9 {
		return nil, errors.New("packet size exception")
	}
	s, e := 9, 9+int(kv.KeySize)
	kv.Key = b[s:e]
	s, e = 9+int(kv.KeySize), 9+int(kv.KeySize)+int(kv.ValueSize)
	kv.Value = b[s:e]
	crc := crc32Checksum(*kv)
	if crc != kv.Crc32 {
		return nil, errors.New("packet crc32 exception")
	}
	return kv, nil
}

func DecodeHeader(b []byte) (*keyValue, error) {
	if len(b) < 9 {
		return nil, errors.New("packet header exception")
	}
	kv := &keyValue{
		Crc32:     constant.ByteOrder.Uint32(b[:4]),
		KeySize:   b[4],
		ValueSize: constant.ByteOrder.Uint32(b[5:9]),
	}
	return kv, nil
}
