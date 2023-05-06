package commutative

import (
	"errors"
	"fmt"
	"math/big"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	uint256 "github.com/holiman/uint256"
)

var (
	U256_MIN = (uint256.NewInt(0)) // default limits
	U256_MAX = (uint256.NewInt(0).SetAllOne())

	U256_ZERO = uint256.NewInt(0)
	U256_ONE  = uint256.NewInt(1)
)

type U256 struct {
	value         *codec.Uint256
	delta         *codec.Uint256
	min           *codec.Uint256
	max           *codec.Uint256
	deltaPositive bool
}

func NewU256(limits ...*uint256.Int) interface{} {
	limits = common.IfThen(len(limits) == 0, []*uint256.Int{U256_MIN, U256_MAX}, limits)
	if (limits[1]).Cmp(limits[0]) < 0 {
		return nil
	}

	return &U256{
		value:         (&codec.Uint256{}).NewInt(0),
		delta:         (&codec.Uint256{}).NewInt(0),
		min:           common.IfThen(limits[0] != nil, (*codec.Uint256)(limits[0].Clone()), (*codec.Uint256)(U256_MIN.Clone())),
		max:           common.IfThen(limits[1] != nil, (*codec.Uint256)(limits[1].Clone()), (*codec.Uint256)(U256_MAX.Clone())),
		deltaPositive: true, // positive delta by default
	}
}

func NewU256FromBytes(value []byte, min, max []byte) interface{} {
	this := &U256{
		value:         (&codec.Uint256{}).NewInt(0),
		delta:         (&codec.Uint256{}).NewInt(0),
		deltaPositive: true,
	}
	this.FromBytes(value, min, max)
	return this
}

func NewU256Delta(delta *uint256.Int, deltaPositive bool) interface{} {
	return &U256{
		value:         nil,
		min:           nil,
		max:           nil,
		delta:         (*codec.Uint256)(delta),
		deltaPositive: deltaPositive,
	}
}

func NewU256DeltaFromBigInt(delta *big.Int) (interface{}, bool) {
	sign := delta.Sign()
	deltaV, overflowed := uint256.FromBig(delta.Abs(delta))
	if overflowed {
		return nil, false
	}

	return &U256{
		value:         nil,
		min:           nil,
		max:           nil,
		delta:         (*codec.Uint256)(deltaV),
		deltaPositive: sign != -1, // >= 0
	}, true
}

// For the codec only, don't use it for other purposes
func (this *U256) New(value, delta, sign, min, max interface{}) interface{} {
	return &U256{
		value:         common.IfThenDo1st(value != nil && value.(*codec.Uint256) != nil && !value.(*codec.Uint256).Eq((*codec.Uint256)(U256_ZERO)), func() *codec.Uint256 { return (*codec.Uint256)(value.(*codec.Uint256)) }, nil),
		delta:         common.IfThenDo1st(delta != nil && delta.(*codec.Uint256) != nil && !delta.(*codec.Uint256).Eq((*codec.Uint256)(U256_ZERO)), func() *codec.Uint256 { return (*codec.Uint256)(delta.(*codec.Uint256)) }, nil),
		deltaPositive: common.IfThenDo1st(sign != nil, func() bool { return sign.(bool) }, true),
		min:           common.IfThenDo1st(min != nil && min.(*codec.Uint256) != nil && !min.(*codec.Uint256).Eq((*codec.Uint256)(U256_MIN)), func() *codec.Uint256 { return (*codec.Uint256)(min.(*codec.Uint256)) }, nil),
		max:           common.IfThenDo1st(max != nil && max.(*codec.Uint256) != nil && !max.(*codec.Uint256).Eq((*codec.Uint256)(U256_MAX)), func() *codec.Uint256 { return (*codec.Uint256)(max.(*codec.Uint256)) }, nil),
	}
}

// ReInit the fields with default values that were removed before export.
func (this *U256) ReInit() {
	this.value = common.IfThen(this.value == nil, (*codec.Uint256)(U256_ZERO.Clone()), this.value)
	this.delta = common.IfThen(this.delta == nil, (*codec.Uint256)(U256_ZERO.Clone()), this.delta)
	this.deltaPositive = common.IfThen(this.delta == nil, true, this.deltaPositive)
	this.min = common.IfThen(this.min == nil, (*codec.Uint256)(U256_ZERO.Clone()), this.min)
	this.max = common.IfThen(this.max == nil, (*codec.Uint256)(U256_MAX.Clone()), this.max)
}

func (this *U256) Value() interface{} { return this.value }
func (this *U256) Delta() interface{} { return this.delta }
func (this *U256) Sign() bool         { return this.delta.Cmp((*codec.Uint256)(U256_ZERO)) >= 0 }
func (this *U256) Min() interface{}   { return this.min }
func (this *U256) Max() interface{}   { return this.max }

func (this *U256) MemSize() uint32                                            { return 32*4 + 1 }
func (this *U256) IsSelf(key interface{}) bool                                { return true }
func (this *U256) TypeID() uint8                                              { return UINT256 }
func (this *U256) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *U256) FromBytes(value []byte, min, max []byte) {
	(*uint256.Int)(this.value).SetBytes(value)
	(*uint256.Int)(this.min).SetBytes(min)
	(*uint256.Int)(this.max).SetBytes(max)
	this.deltaPositive = true
}

func (this *U256) Clone() interface{} {
	return &U256{
		value:         common.IfThenDo1st(this.value != nil, func() *codec.Uint256 { return this.value.Clone().(*codec.Uint256) }, nil),
		delta:         common.IfThenDo1st(this.delta != nil, func() *codec.Uint256 { return this.delta.Clone().(*codec.Uint256) }, nil),
		min:           common.IfThenDo1st(this.min != nil, func() *codec.Uint256 { return this.min.Clone().(*codec.Uint256) }, nil),
		max:           common.IfThenDo1st(this.max != nil, func() *codec.Uint256 { return this.max.Clone().(*codec.Uint256) }, nil),
		deltaPositive: this.deltaPositive,
	}
}

func (this *U256) Equal(other interface{}) bool {
	return common.Equal(this.value, other.(*U256).value, func(v *codec.Uint256) bool { return v.Eq((*codec.Uint256)(U256_ZERO)) }) &&
		common.Equal(this.delta, other.(*U256).delta, func(v *codec.Uint256) bool { return v.Eq((*codec.Uint256)(U256_ZERO)) }) &&
		common.Equal(this.min, other.(*U256).min, func(v *codec.Uint256) bool { return v.Eq((*codec.Uint256)(U256_ZERO)) }) &&
		common.Equal(this.max, other.(*U256).max, func(v *codec.Uint256) bool { return v.Eq((*codec.Uint256)(U256_MAX)) })
}

func (this *U256) Get() (interface{}, uint32, uint32) {
	if U256_ZERO.Eq((*uint256.Int)(this.delta)) {
		return (*uint256.Int)(this.value), 1, 0
	}
	return (*uint256.Int)((&codec.Uint256{}).Add(this.value.Clone().(*codec.Uint256), this.delta)), 1, 1
}

func (this *U256) isOverflowed(lhv *codec.Uint256, lhvSign bool, rhv *codec.Uint256, rhvSign bool) (*codec.Uint256, bool) {
	if lhvSign == rhvSign { // Both positive or negative
		summed, overflowed := ((*uint256.Int)(lhv)).AddOverflow((*uint256.Int)(lhv), (*uint256.Int)(rhv))
		if overflowed {
			return nil, true
		}
		return (*codec.Uint256)(summed), lhvSign
	}

	if lhv.Cmp(rhv) < 1 { // v0 <= rhv
		return (&codec.Uint256{}).NewInt(0).Sub(rhv, lhv),
			common.IfThen((*uint256.Int)(rhv).Eq((*uint256.Int)(lhv)), true, rhvSign) // sign is positive when delta values cancel out each other

	}
	return (&codec.Uint256{}).NewInt(0).Sub(lhv, rhv), lhvSign
}

// Set delta
func (this *U256) Set(newDelta interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if newDelta.(*U256).delta.Eq((*codec.Uint256)(U256_ZERO)) {
		return this, 1, 0, 0, nil
	}

	accumDelta, deltaSign := this.isOverflowed(this.delta.Clone().(*codec.Uint256), this.deltaPositive, newDelta.(*U256).delta, newDelta.(*U256).deltaPositive)
	if accumDelta == nil {
		return this, 0, 0, 1, errors.New("Error: Value out of range")
	}

	tempV, possitive := this.isOverflowed(this.value.Clone().(*codec.Uint256), true, accumDelta.Clone().(*codec.Uint256), deltaSign)
	if tempV == nil || !possitive { // Result must be possitive
		return this, 0, 0, 1, errors.New("Error: Value out of range")
	}

	if this.min.Cmp(tempV) < 1 && tempV.Cmp(this.max) < 1 {
		this.delta = accumDelta
		this.deltaPositive = deltaSign
		return this, 0, 0, 1, nil
	}
	return this, 0, 0, 1, errors.New("Error: Value out of range")
}

func (this *U256) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	this.ReInit()
	vec := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*U256)
		}

		if this == nil && v == nil { // Delete a non-existent
			this = nil
		}

		if this != nil && v != nil { // Update an existent
			if _, _, _, _, err := this.Set(v.(*U256), nil); err != nil {
				panic(err)
			}
		}

		if this != nil && v == nil { // Delete an existent
			this = nil
		}
	}

	newValue, _, _ := this.Get()
	this.value = (*codec.Uint256)(newValue.(*uint256.Int))
	(*uint256.Int)(this.delta).Clear()
	return this
}

func (this *U256) Reset() {
	this.delta = (&codec.Uint256{}).NewInt(0)
}

func (this *U256) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

func (this *U256) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}