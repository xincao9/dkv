package appendfile

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"math"
)

var (
	byteOrder = binary.BigEndian
)

type keyValue struct {
	Crc32     uint32
	KeySize   uint8
	ValueSize uint16
	Key       []byte
	Value     []byte
}

func NewKeyValue(key []byte, value []byte) (*keyValue, error) {
	kl := len(key)
	vl := len(value)
	if kl > math.MaxUint8 || vl > math.MaxUint16 {
		return nil, errors.New("key or value length exceeds maximum")
	}
	kv := &keyValue{
		KeySize:   uint8(kl),
		ValueSize: uint16(vl),
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
	b := make([]byte, int(kv.KeySize)+int(kv.ValueSize)+7)
	byteOrder.PutUint32(b[0:4], kv.Crc32)
	b[4] = kv.KeySize
	byteOrder.PutUint16(b[5:7], kv.ValueSize)
	s, e := 7, 7+int(kv.KeySize)
	copy(b[s:e], kv.Key)
	s, e = 7+int(kv.KeySize), 7+int(kv.KeySize)+int(kv.ValueSize)
	copy(b[s:e], kv.Value)
	return b
}

func Decode(b []byte) (*keyValue, error) {
	if len(b) < 9 {
		return nil, errors.New("packet exception")
	}
	kv := &keyValue{
		Crc32:     byteOrder.Uint32(b[:4]),
		KeySize:   b[4],
		ValueSize: byteOrder.Uint16(b[5:7]),
	}
	s := int(kv.KeySize) + int(kv.ValueSize)
	if len(b) < s+7 {
		return nil, errors.New("packet exception")
	}
	s, e := 7, 7+int(kv.KeySize)
	kv.Key = b[s:e]
	s, e = 7+int(kv.KeySize), 7+int(kv.KeySize)+int(kv.ValueSize)
	kv.Value = b[s:e]
	crc := crc32Checksum(*kv)
	if crc != kv.Crc32 {
		return nil, errors.New("packet exception")
	}
	return kv, nil
}

func DecodeHeader(b []byte) (*keyValue, error) {
	if len(b) < 7 {
		return nil, errors.New("packet exception")
	}
	kv := &keyValue{
		Crc32:     byteOrder.Uint32(b[:4]),
		KeySize:   b[4],
		ValueSize: byteOrder.Uint16(b[5:7]),
	}
	return kv, nil
}
