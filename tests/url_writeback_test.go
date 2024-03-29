package ccurltest

import (
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
)

func TestAddAndDelete(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	path := commutative.NewPath()
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	acctTrans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	url.Init(store)

	// Delete an non-existing entry, should NOT appear in the transitions
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil); err == nil {
		t.Error("Deleting an non-existing entry should've flaged an error", err)
	}

	raw := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	if acctTrans := raw; len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestRecursiveDeletionSameBatch(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path := commutative.NewPath()
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)
	// _, addPath := url.Export(indexer.Sorter)
	addPath := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(addPath).Encode()).(indexer.Univalues))
	// url.Import(url.Decode(indexer.Univalues(addPath).Encode()))
	url.Sort()
	url.Commit([]uint32{1})

	url.Init(store)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewInt64(1))
	// _, addTrans := url.Export(indexer.Sorter)
	addTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	// url.Import(url.Decode(indexer.Univalues(addTrans).Encode()))
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(addTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	if v, _ := url.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v == nil {
		t.Error("Error: Failed to read the key !")
	}

	url2 := ccurl.NewConcurrentUrl(store)
	if v, _ := url2.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v.(int64) != 1 {
		t.Error("Error: Failed to read the key !")
	}

	url2.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", nil)
	// _, deleteTrans := url2.Export(indexer.Sorter)
	deleteTrans := indexer.Univalues(common.Clone(url2.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	if v, _ := url2.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v != nil {
		t.Error("Error: Failed to read the key !")
	}

	url3 := ccurl.NewConcurrentUrl(store)
	url3.Import(append(addTrans, deleteTrans...))
	url3.Sort()
	url3.Commit([]uint32{1, 2})

	if v, _ := url3.Read(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.Int64)); v != nil {
		t.Error("Error: Failed to delete the entry !")
	}
}

func TestApplyingTransitionsFromMulitpleBatches(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(acctTrans)
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	path := commutative.NewPath()
	_, err := url.Write(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path)

	if err != nil {
		t.Error("error")
	}

	acctTrans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	url.Init(store)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil)

	if acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{}); len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestRecursiveDeletionDifferentBatch(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	in := indexer.Univalues(acctTrans).Encode()
	out := indexer.Univalues{}.Decode(in).(indexer.Univalues)
	// url.Import(url.Decode(indexer.Univalues(out).Encode()))
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(out).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	// create a path
	path := commutative.NewPath()
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewString("1"))
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewString("2"))

	acctTrans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	in = indexer.Univalues(acctTrans).Encode()
	out = indexer.Univalues{}.Decode(in).(indexer.Univalues)
	// url.Import(url.Decode(indexer.Univalues(out).Encode()))
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(out).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	_1, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", new(noncommutative.String))
	if _1 != "1" {
		t.Error("Error: Not match")
	}

	url.Init(store)
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewString("3"))
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewString("4"))

	outpath, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys := outpath.(*orderedset.OrderedSet).Keys()
	if reflect.DeepEqual(keys, []string{"1", "2", "3", "4"}) {
		t.Error("Error: Not match")
	}

	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil) // delete the path
	if acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{}); len(acctTrans) != 3 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestStateUpdate(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	// _, initTrans := url.Export(indexer.Sorter)
	initTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})

	// url.Import(url.Decode(indexer.Univalues(initTrans).Encode()))
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(initTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	tx0bytes, trans, err := Create_Ctrn_0(alice, store)
	if err != nil {
		t.Error(err)
	}
	tx0Out := indexer.Univalues{}.Decode(tx0bytes).(indexer.Univalues)
	tx0Out = trans
	tx1bytes, err := Create_Ctrn_1(alice, store)
	if err != nil {
		t.Error(err)
	}

	tx1Out := indexer.Univalues{}.Decode(tx1bytes).(indexer.Univalues)

	// url.Import(url.Decode(indexer.Univalues(append(tx0Out, tx1Out...)).Encode()))
	url.Import((append(tx0Out, tx1Out...)))
	url.Sort()
	url.Commit([]uint32{0, 1})
	//need to encode delta only now it encodes everything

	if err := CheckPaths(alice, url); err != nil {
		t.Error(err)
	}

	v, _ := url.Read(9, "blcc://eth1.0/account/"+alice+"/storage/", &commutative.Path{}) //system doesn't generate sub paths for /storage/
	// if v.(*commutative.Path).CommittedLength() != 2 {
	// 	t.Error("Error: Wrong sub paths")
	// }

	// if !reflect.DeepEqual(v.([]string), []string{"ctrn-0/", "ctrn-1/"}) {
	// 	t.Error("Error: Didn't find the subpath!")
	// }

	v, _ = url.Read(9, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys := v.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		t.Error("Error: Keys don't match !")
	}

	// Delete the container-0
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil {
		t.Error("Error: Cann't delete a path twice !")
	}

	if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil {
		t.Error("Error: The path should be gone already !")
	}

	transitions := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	out := indexer.Univalues{}.Decode(indexer.Univalues(transitions).Encode()).(indexer.Univalues)

	url.Import(out)
	url.Sort()
	url.Commit([]uint32{1})

	if v, _ := url.Read(ccurlcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil {
		t.Error("Error: Should be gone already !")
	}
}
