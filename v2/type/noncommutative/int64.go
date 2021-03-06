package noncommutative

import (
	"fmt"

	"github.com/arcology-network/common-lib/codec"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	"github.com/elliotchance/orderedmap"
)

type Int64 int64

func NewInt64(v int64) interface{} {
	var this Int64 = Int64(v)
	return &this
}

func (this *Int64) TypeID() uint8 { return ccurlcommon.NoncommutativeInt64 }

// create a new path
func (this *Int64) Deepcopy() interface{} {
	value := *this
	return (*Int64)(&value)
}

func (this *Int64) Value() interface{} {
	return this
}

func (this *Int64) ToAccess() interface{} {
	return nil
}

func (this *Int64) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *Int64) Peek(source interface{}) interface{} {
	return this
}

func (this *Int64) Delta(source interface{}) interface{} {
	return this
}

func (this *Int64) Set(tx uint32, path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value != nil {
		*this = Int64(*(value.(*Int64)))
	}
	return 0, 1, nil
}

func (this *Int64) ApplyDelta(tx uint32, v interface{}) ccurlcommon.TypeInterface {
	for iter := v.(*orderedmap.Element); iter != nil; iter = iter.Next() {
		if iter.Value == nil {
			continue
		}

		if v := iter.Value.(ccurlcommon.UnivalueInterface).Value(); v != nil {
			this.Set(tx, "", v.(*Int64), nil)
		} else {
			this = nil
		}
	}

	if this == nil {
		return nil
	}
	return this
}

func (this *Int64) Composite() bool { return false }

func (this *Int64) Encode() []byte {
	return codec.Int64(int64(*this)).Encode()
}

func (*Int64) Decode(bytes []byte) interface{} {
	this := Int64(codec.Int64(0).Decode(bytes))
	return &this
}

func (this *Int64) EncodeCompact() []byte {
	return this.Encode()
}

func (this *Int64) DecodeCompact(bytes []byte) interface{} {
	return this.Decode(bytes)
}

func (*Int64) Purge() {}

func (this *Int64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *Int64) Print() {
	fmt.Println(*this)
	fmt.Println()
}
