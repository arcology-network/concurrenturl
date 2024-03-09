/*
 *   Copyright (c) 2023 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package committertest

import (
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/slice"
	cache "github.com/arcology-network/eu/cache"
	stgcommitter "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/commutative"
	importer "github.com/arcology-network/storage-committer/importer"
	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	platform "github.com/arcology-network/storage-committer/platform"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

func TestEmptyNodeSet(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore(nil, nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)
	committer := stgcommitter.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit()

	// acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	committer.Import(univalue.Univalues{})
	committer.Sort()
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit()

	committer.Import(univalue.Univalues{})
	committer.Sort()
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit()
}
func TestAddAndDelete(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore(nil, nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)
	committer := stgcommitter.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit()

	// acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	committer.Import(univalue.Univalues{})
	committer.Sort()
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit()

	committer.Init(store)
	path := commutative.NewPath()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{1})
	committer.Commit()

	committer.Init(store)
	writeCache.Reset(writeCache)

	// Delete an non-existing entry, should NOT appear in the transitions
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil); err == nil {
		t.Error("Deleting an non-existing entry should've flaged an error", err)
	}

	raw := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	if acctTrans := raw; len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}

	// Delete an non-existing entry, should NOT appear in the transitions
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", noncommutative.NewString("124")); err != nil {
		t.Error("Failed to write", err)
	}

	if v, _, err := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil); v != "124" {
		t.Error("Wrong return value", err)
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", new(noncommutative.String)); v != "124" {
		t.Error("Error: Wrong return value")
	}
}

func TestRecursiveDeletionSameBatch(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore(nil, nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)
	committer := stgcommitter.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit()

	committer.Init(store)
	// create a path
	path := commutative.NewPath()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)
	// _, addPath := writeCache.Export(importer.Sorter)
	addPath := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(addPath).Encode()).(univalue.Univalues))
	// committer.Import(committer.Decode(univalue.Univalues(addPath).Encode()))
	committer.Sort()
	committer.Precommit([]uint32{1})
	committer.Commit()
	committer.Init(store)

	writeCache.Reset(writeCache)

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewInt64(1))
	// _, addTrans := writeCache.Export(importer.Sorter)
	addTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	// committer.Import(committer.Decode(univalue.Univalues(addTrans).Encode()))
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(addTrans).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{1})
	committer.Commit()
	writeCache.Reset(writeCache)

	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v == nil {
		t.Error("Error: Failed to read the key !")
	}

	// url2 := stgcommitter.NewStorageCommitter(store)
	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v.(int64) != 1 {
		t.Error("Error: Failed to read the key !")
	}

	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", nil)
	// _, deleteTrans := url2.WriteCache().Export(importer.Sorter)
	deleteTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v != nil {
		t.Error("Error: Failed to read the key !")
	}

	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import(append(addTrans, deleteTrans...))
	committer.Sort()
	committer.Precommit([]uint32{1, 2})
	committer.Commit()
	writeCache.Reset(writeCache)

	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v != nil {
		t.Error("Error: Failed to delete the entry !")
	}
}

func TestApplyingTransitionsFromMulitpleBatches(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore(nil, nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	committer := stgcommitter.NewStorageCommitter(store)
	committer.Import(acctTrans)
	committer.Sort()
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit()

	committer.Init(store)
	path := commutative.NewPath()
	_, err := writeCache.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path)

	if err != nil {
		t.Error("error")
	}

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{1})
	committer.Commit()

	committer.Init(store)

	writeCache.Reset(writeCache)
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil)

	if acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{}); len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestRecursiveDeletionDifferentBatch(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore(nil, nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	in := univalue.Univalues(acctTrans).Encode()
	out := univalue.Univalues{}.Decode(in).(univalue.Univalues)
	// committer.Import(committer.Decode(univalue.Univalues(out).Encode()))

	committer := stgcommitter.NewStorageCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(out).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit()

	committer.Init(store)
	// create a path
	path := commutative.NewPath()
	writeCache.Reset(writeCache)

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewString("1"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewString("2"))

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	in = univalue.Univalues(acctTrans).Encode()
	out = univalue.Univalues{}.Decode(in).(univalue.Univalues)
	// committer.Import(committer.Decode(univalue.Univalues(out).Encode()))
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(out).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{1})
	committer.Commit()

	_1, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.String))
	if _1 != "1" {
		t.Error("Error: Not match")
	}

	committer.Init(store)
	writeCache.Reset(writeCache)

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewString("3"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewString("4"))

	outpath, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys := outpath.(*deltaset.DeltaSet[string]).Elements()
	if reflect.DeepEqual(keys, []string{"1", "2", "3", "4"}) {
		t.Error("Error: Not match")
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil) // delete the path
	if acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{}); len(acctTrans) != 3 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestStateUpdate(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore(nil, nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	// _, initTrans := writeCache.Export(importer.Sorter)
	initTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})

	// committer.Import(committer.Decode(univalue.Univalues(initTrans).Encode()))
	committer := stgcommitter.NewStorageCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(initTrans).Encode()).(univalue.Univalues))
	committer.Sort()
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit()
	committer.Init(store)

	writeCache.Reset(writeCache)
	tx0bytes, trans, err := Create_Ctrn_0(alice, store)
	if err != nil {
		t.Error(err)
	}
	tx0Out := univalue.Univalues{}.Decode(tx0bytes).(univalue.Univalues)
	tx0Out = trans
	tx1bytes, err := Create_Ctrn_1(alice, store)
	if err != nil {
		t.Error(err)
	}

	tx1Out := univalue.Univalues{}.Decode(tx1bytes).(univalue.Univalues)

	// committer.Import(committer.Decode(univalue.Univalues(append(tx0Out, tx1Out...)).Encode()))
	committer.Import((append(tx0Out, tx1Out...)))
	committer.Sort()
	committer.Precommit([]uint32{0, 1})
	committer.Commit()
	//need to encode delta only now it encodes everything

	writeCache.Reset(writeCache)
	if err := CheckPaths(alice, writeCache); err != nil {
		t.Error(err)
	}

	v, _, _ := writeCache.Read(9, "blcc://eth1.0/account/"+alice+"/storage/", &commutative.Path{}) //system doesn't generate sub paths for /storage/
	// if v.(*commutative.Path).CommittedLength() != 2 {
	// 	t.Error("Error: Wrong sub paths")
	// }

	// if !reflect.DeepEqual(v.([]string), []string{"ctrn-0/", "ctrn-1/"}) {
	// 	t.Error("Error: Didn't find the subpath!")
	// }

	v, _, _ = writeCache.Read(9, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys := v.(*deltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		t.Error("Error: Keys don't match !")
	}

	// Delete the container-0
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil {
		t.Error("Error: Cann't delete a path twice !")
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil {
		t.Error("Error: The path should be gone already !")
	}

	transitions := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	out := univalue.Univalues{}.Decode(univalue.Univalues(transitions).Encode()).(univalue.Univalues)

	committer.Import(out)
	committer.Sort()
	committer.Precommit([]uint32{1})
	committer.Commit()

	writeCache.Reset(writeCache)
	if v, _, _ := writeCache.Read(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil {
		t.Error("Error: Should be gone already !")
	}
}