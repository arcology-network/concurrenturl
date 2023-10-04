package concurrenturl

import (
	common "github.com/arcology-network/common-lib/common"
	interfaces "github.com/arcology-network/concurrenturl/interfaces"
	// ethparams "github.com/arcology-network/evm/params"
)

const (
	READ_NONEXIST          = uint64(3)
	READ_COMMITTED_FROM_DB = uint64(1000) // Read for the state db
)

type Fee struct{}

func (Fee) Reader(v interface{}) uint64 { // Call this before setting the value attribute to nil
	if v == nil {
		return READ_NONEXIST
	}

	typedv := v.(interfaces.Univalue).Value()
	dataSize := common.IfThenDo1st(typedv != nil, func() uint64 { return uint64(typedv.(interfaces.Type).MemSize()) }, 0)
	return common.IfThen(
		v.(interfaces.Univalue).Reads() > 1, // Is hot loaded
		common.Max(dataSize/32, 1)*3,
		(dataSize/32)*5000,
	)
}

func (Fee) Writer(key string, v interface{}, writecache interfaces.WriteCache) int64 { // May get refunds sometimes
	committedv := writecache.RetriveShallow(key)
	committedSize := common.IfThenDo1st(committedv != nil, func() uint64 { return uint64(committedv.(interfaces.Type).MemSize()) }, 0)

	dataSize := common.IfThenDo1st(v != nil, func() uint64 { return uint64(v.(interfaces.Type).Size()) }, 0)
	return common.IfThen(
		committedSize > 0,
		common.Max(int64(dataSize/32), 1)*20000,
		common.IfThen(
			int64(dataSize) > int64(committedSize),
			int64(dataSize),
			((int64(committedSize)-int64(dataSize))/32)*20000,
		),
	)
}