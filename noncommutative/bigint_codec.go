package noncommutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/evm/rlp"
)

func (this *Bigint) Size() uint32 {
	v := codec.Bigint(*this)
	return v.Size()
}

func (this *Bigint) Encode() []byte {
	v := codec.Bigint(*this)
	return v.Encode()
}

func (this *Bigint) EncodeToBuffer(buffer []byte) int {
	v := codec.Bigint(*this)
	return v.EncodeToBuffer(buffer)
}

func (this *Bigint) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}
	this = (*Bigint)((&codec.Bigint{}).Decode(buffer).(*codec.Bigint))
	return this
}

// func (this *Bigint) Encode() []byte {
// 	return this.Encode()
// }

// func (this *Bigint) DecodeCompact(bytes []byte) interface{} {
// 	return this.Decode(bytes)
// }

func (this *Bigint) StorageEncode() []byte {
	buffer, _ := rlp.EncodeToBytes(*this)
	return buffer
}

func (this *Bigint) StorageDecode(buffer []byte) interface{} {
	rlp.DecodeBytes(buffer, this)
	return this
}

func (this *Bigint) Reset() {}

func (this *Bigint) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

func (this *Bigint) Print() {
	fmt.Println(*this)
	fmt.Println()
}
