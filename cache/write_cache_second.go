/*
 *   Copyright (c) 2023 Arcology Network
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cache

import (
	"errors"
	"math"

	common "github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
	"github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/interfaces"
)

func (this *WriteCache) IndexOf(tx uint32, path string, key interface{}, T any) (uint64, uint64) {
	if !common.IsPath(path) {
		return math.MaxUint64, READ_NONEXIST //, errors.New("Error: Not a path!!!")
	}

	getter := func(v interface{}) (uint32, uint32, uint32, interface{}) { return 1, 0, 0, v }
	if v, err := this.Do(tx, path, getter, T); err == nil {
		pathInfo := v.(interfaces.Univalue).Value()
		if common.IsType[*commutative.Path](pathInfo) && common.IsType[string](key) {
			return pathInfo.(*commutative.Path).View().IdxOf(key.(string)), 0
		}
	}
	return math.MaxUint64, READ_NONEXIST
}

func (this *WriteCache) KeyAt(tx uint32, path string, index interface{}, T any) (string, uint64) {
	if !common.IsPath(path) {
		return "", READ_NONEXIST //, errors.New("Error: Not a path!!!")
	}

	getter := func(v interface{}) (uint32, uint32, uint32, interface{}) { return 1, 0, 0, v }
	if v, err := this.Do(tx, path, getter, T); err == nil {
		pathInfo := v.(interfaces.Univalue).Value()
		if common.IsType[*commutative.Path](pathInfo) && common.IsType[uint64](index) {
			return pathInfo.(*commutative.Path).View().KeyAt(index.(uint64)), 0
		}
	}
	return "", READ_NONEXIST
}

// func (this *Write) Peek(path string, T any) (interface{}, uint64) {
// 	typedv, univ := this.writeCache.Peek(path, T)
// 	var v interface{}
// 	if typedv != nil {
// 		v, _, _ = typedv.(interfaces.Type).Get()
// 	}
// 	return v, Fee{}.Reader(univ.(interfaces.Univalue))
// }

func (this *WriteCache) PeekCommitted(path string, T any) (interface{}, uint64) {
	v, _ := this.store.Retrive(path, T)
	return v, READ_COMMITTED_FROM_DB
}

// func (this *WriteCache) Read(tx uint32, path string, T any) (interface{}, uint64) {
// 	typedv, univ := this.writeCache.Read(tx, path, T)
// 	// fmt.Println("Read: ", path, "|", typedv)
// 	return typedv, Fee{}.Reader(univ.(interfaces.Univalue))
// }

// func (this *WriteCache) Write(tx uint32, path string, value interface{}) (int64, error) {
// 	// fmt.Println("Write: ", path, "|", value)
// 	fee := int64(0) //Fee{}.Writer(path, value, this.writeCache)
// 	if value == nil || (value != nil && value.(interfaces.Type).TypeID() != uint8(reflect.Invalid)) {
// 		return fee, common.FilterSecond(this.writeCache.Write(tx, path, value))
// 	}

// 	return fee, errors.New("Error: Unknown data type !")
// }

// func (this *WriteCache) Do(tx uint32, path string, doer interface{}, T any) (interface{}, error) {
// 	return this.writeCache.Do(tx, path, doer, T), nil
// }

// Read th Nth element under a path
// func (this *WriteCache) getKeyByIdx(tx uint32, path string, idx uint64) (interface{}, uint64, error) {
// 	if !common.IsPath(path) {
// 		return nil, READ_NONEXIST, errors.New("Error: Not a path!!!")
// 	}

// 	meta, readFee := this.Read(tx, path, new(commutative.Path)) // read the container meta
// 	return common.IfThen(meta == nil,
// 		meta,
// 		common.IfThenDo1st(idx < uint64(len(meta.(*orderedset.OrderedSet).Keys())), func() interface{} { return path + meta.(*orderedset.OrderedSet).Keys()[idx] }, nil),
// 	), readFee, nil
// }

// func (this *WriteCache) ReadAt(tx uint32, path string, idx uint64, T any) (interface{}, uint64, error) {
// 	if key, Fee, err := this.getKeyByIdx(tx, path, idx); err == nil && key != nil {
// 		v, Fee := this.Read(tx, key.(string), T)
// 		return v, Fee, nil
// 	} else {
// 		return key, Fee, err
// 	}
// }

// Read th Nth element under a path
func (this *WriteCache) DoAt(tx uint32, path string, idx uint64, do interface{}, T any) (interface{}, uint64, error) {
	if key, Fee, err := this.getKeyByIdx(tx, path, idx); err == nil && key != nil {
		v, err := this.Do(tx, key.(string), do, T)
		return v, Fee, err
	} else {
		return key, Fee, err
	}
}

// Read th Nth element under a path
func (this *WriteCache) PopBack(tx uint32, path string, T any) (interface{}, int64, error) {
	if !common.IsPath(path) {
		return nil, int64(READ_NONEXIST), errors.New("Error: Not a path!!!")
	}
	pathDecoder := T

	meta, Fee := this.ReadEx(tx, path, pathDecoder) // read the container meta

	subkeys := meta.(*orderedset.OrderedSet).Keys()
	if subkeys == nil || len(subkeys) == 0 {
		return nil, int64(Fee), errors.New("Error: The path is either empty or doesn't exist")
	}

	key := path + subkeys[len(subkeys)-1]

	value, Fee := this.ReadEx(tx, key, pathDecoder)
	if value == nil {
		return nil, int64(Fee), errors.New("Error: Empty container!")
	}

	writeFee, err := this.Write(tx, key, nil)
	return value, writeFee, err
}

// Read th Nth element under a path
// func (this *WriteCache) WriteAt(tx uint32, path string, idx uint64, T any) (int64, error) {
// 	if !common.IsPath(path) {
// 		return int64(READ_NONEXIST), errors.New("Error: Not a path!!!")
// 	}

// 	if key, Fee, err := this.getKeyByIdx(tx, path, idx); err == nil {
// 		return this.Write(tx, key.(string), T)
// 	} else {
// 		return int64(Fee), err
// 	}
// }
