package noncommutative

import (
	"bytes"
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
)

func (this *Bytes) HeaderSize() uint32 {
	return 3 * codec.UINT32_LEN
}

func (this *Bytes) Size() uint32 {
	return this.HeaderSize() + this.MemSize()
}

func (this *Bytes) Encode() []byte {
	byteset := [][]byte{
		codec.Bool(this.placeholder).Encode(),
		this.value,
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Bytes) EncodeToBuffer(buffer []byte) int {
	offset := codec.Bool(this.placeholder).EncodeToBuffer(buffer)
	return offset + codec.Bytes(this.value).EncodeToBuffer(buffer[offset:])
}

func (this *Bytes) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}

	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	return &Bytes{
		placeholder: bool(codec.Bool(true).Decode(fields[0]).(codec.Bool)),
		value:       bytes.Clone(fields[1]),
	}
}

// func (this *Bytes) Encode() []byte {
// 	return this.Encode()
// }

// func (this *Bytes) DecodeCompact(bytes []byte) interface{} {
// 	return this.Decode(bytes)
// }

func (this *Bytes) Reset() {}

func (this *Bytes) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

func (this *Bytes) Print() {
	fmt.Println(*this)
	fmt.Println()
}
