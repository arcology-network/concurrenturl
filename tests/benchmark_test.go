package ccurltest

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/common-lib/common"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	"github.com/arcology-network/common-lib/merkle"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	storage "github.com/arcology-network/concurrenturl/storage"
	univalue "github.com/arcology-network/concurrenturl/univalue"
	orderedmap "github.com/elliotchance/orderedmap"
	"github.com/google/btree"
	// "github.com/google/btree"
	// ehtrlp "github.com/elliotchance/orderedmap"
)

func BenchmarkAccountMerkleImportPerf(b *testing.B) {
	// lut := datacompression.NewCompressionLut()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	// fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	// if err != nil {
	// 	b.Error(err)
	// 	return
	// }
	store := chooseDataStore()

	meta := commutative.NewPath()
	store.Inject((ccurlcommon.ETH10_ACCOUNT_PREFIX), meta)

	url := ccurl.NewConcurrentUrl(store)
	for i := 0; i < 100000; i++ {
		if _, err := url.NewAccount(0, fmt.Sprint(rand.Float64())); err != nil { // Preload account structure {
			b.Error(err)
		}
	}
	acct := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCAccess{})

	t0 := time.Now()
	indexer.Univalues(acct).Encode()
	b.Log("Transition Encoding: ", len(acct), time.Since(t0))
}

func BenchmarkSingleAccountCommit(b *testing.B) {
	// fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	// if err != nil {
	// 	b.Error(err)
	// 	return
	// }
	store := chooseDataStore()

	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath() // create a path
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	for i := 0; i < 1; i++ {
		if _, err := url.Write(0, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"+fmt.Sprint(i), noncommutative.NewString("fmt.Sprint(i)")); err != nil { /* The first Element */
			b.Error(err)
		}
	}

	//t0 = time.Now()
	// _, transitions := url.Export(nil)
	transitions := indexer.Univalues(common.Clone(url.Export())).To(indexer.ITCAccess{})

	// in := indexer.Univalues(transitions).Encode()
	//out := indexer.Univalues{}.Decode(in).(indexer.Univalues)

	//fmt.Println("Export:", time.Since(t0))

	t0 := time.Now()

	url.Import(transitions)
	url.Sort()
	url.Commit([]uint32{0})
	fmt.Println("Account Commit single Total time= :", time.Since(t0))
}

func BenchmarkMultipleAccountCommit(b *testing.B) {
	// fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	// if err != nil {
	// 	b.Error(err)
	// 	return
	// }
	store := chooseDataStore()

	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath() // create a path
	if _, err := url.Write(0, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	t0 := time.Now()
	for i := 0; i < 100000; i++ {
		acct := fmt.Sprint(rand.Int())
		if _, err := url.NewAccount(ccurlcommon.SYSTEM, acct); err != nil { // NewAccount account structure {
			fmt.Println(err)
		}

		path := commutative.NewPath() // create a path
		if _, err := url.Write(0, "blcc://eth1.0/account/"+acct+"/storage/ctrn-0/", path); err != nil {
			b.Error(err)
		}

		for j := 0; j < 4; j++ {
			if _, err := url.Write(0, "blcc://eth1.0/account/"+acct+"/storage/ctrn-0/elem-0"+fmt.Sprint(j), noncommutative.NewString("fmt.Sprint(i)")); err != nil { /* The first Element */
				b.Error(err)
			}
		}
	}
	fmt.Println("Write 2500 accounts in :", time.Since(t0))

	t0 = time.Now()
	trans := indexer.Univalues(common.Clone(url.Export())).To(indexer.ITCTransition{})
	fmt.Println("Export:", time.Since(t0))

	t0 = time.Now()
	url.Import(trans)
	fmt.Println("Import: ", len(trans), " in: ", time.Since(t0))

	t0 = time.Now()
	url.Sort()
	fmt.Println("Sort: ", len(trans), " in: ", time.Since(t0))

	t0 = time.Now()

	url.Commit([]uint32{0})
	fmt.Println("Commit:", time.Since(t0))

	t0 = time.Now()
	// nilHash := merkle.Sha256(nil)
	// fmt.Println("Hash:", nilHash)
	fmt.Println("merkle: ", time.Since(t0))
}

func BenchmarkUrlAddThenDelete(b *testing.B) {
	store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)
	meta := commutative.NewPath()
	url.Write(ccurlcommon.SYSTEM, ccurlcommon.ETH10_ACCOUNT_PREFIX, meta)
	trans := indexer.Univalues(common.Clone(url.Export())).To(indexer.ITCTransition{})

	url.Import(trans)
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath()
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	t0 := time.Now()
	for i := 0; i < 50000; i++ {
		_, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-"+fmt.Sprint(i), noncommutative.NewInt64(int64(i)))
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Write "+fmt.Sprint(50000), time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 50000; i++ {
		if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-"+fmt.Sprint(i), nil); err != nil {
			panic(err)
		}
	}
	fmt.Println("Deleted 50000 keys "+fmt.Sprint(50000), time.Since(t0))
}

func BenchmarkUrlAddThenPop(b *testing.B) {
	store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)
	meta := commutative.NewPath()
	url.Write(ccurlcommon.SYSTEM, ccurlcommon.ETH10_ACCOUNT_PREFIX, meta)

	trans := indexer.Univalues(common.Clone(url.Export())).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(trans).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath()
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	t0 := time.Now()
	for i := 0; i < 50000; i++ {
		v := noncommutative.NewBytes([]byte(fmt.Sprint(rand.Float64())))
		_, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-"+fmt.Sprint(i), v)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Write "+fmt.Sprint(50000), "noncommutative bytes in", time.Since(t0))

	// t0 = time.Now()
	// v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	// for i := 0; i < 50000; i++ {
	// 	key := v.(*commutative.Path).Next()
	// 	url.Write(1, key, nil)
	// }
	// fmt.Println("Pop 50000 noncommutative bytes in", fmt.Sprint(50000), time.Since(t0))
}

// func BenchmarkOrderedMap(b *testing.B) {
// 	m := orderedmap.NewOrderedMap()
// 	alice := AliceAccount()
// 	t0 := time.Now()

// 	for i := 0; i < 100000; i++ {
// 		m.Set("blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"+fmt.Sprint(i))
// 	}
// 	fmt.Println("orderedmap Insertion:", time.Since(t0))

// 	t0 = time.Now()
// 	m2 := make(map[string]bool)
// 	for i := 0; i < 100000; i++ {
// 		m2["blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"+fmt.Sprint(i)] = true
// 	}
// 	fmt.Println("Golang map Insertion:", time.Since(t0))

// 	t0 = time.Now()
// 	m.Keys()
// 	fmt.Println("orderedmap get keys ", time.Since(t0))

// 	t0 = time.Now()
// 	targeStr := make([]string, 100000)
// 	for i := 0; i < 100000; i++ {
// 		targeStr[i] = "blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-0" + fmt.Sprint(i)
// 	}
// 	fmt.Println("Copy keys "+fmt.Sprint(len(targeStr)), time.Since(t0))
// }

// func BenchmarkOrderedMapInit(b *testing.B) {
// 	t0 := time.Now()
// 	orderedMaps := make([]*orderedmap.OrderedMap, 100000)
// 	for i := 0; i < len(orderedMaps); i++ {
// 		orderedMaps[i] = orderedmap.NewOrderedMap()
// 	}
// 	fmt.Println("Initialized  "+fmt.Sprint(len(orderedMaps)), "OrderedMap in", time.Since(t0))
// }

func BenchmarkInsertAndDelete(b *testing.B) {
	m := orderedmap.NewOrderedMap()
	alice := AliceAccount()

	t0 := time.Now()
	for i := 0; i < 100000; i++ {
		m.Set("blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"+fmt.Sprint(i), 0)
	}
	fmt.Println("orderedmap Insertion:", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 100000; i++ {
		//	m.Delete("blcc://eth1.0/account/" + alice +"/storage/ctrn-0/elem-0" + fmt.Sprint(i))
		m.Delete("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-0" + fmt.Sprint(i))
	}

	fmt.Println("Delete then delete keys:", time.Since(t0))
}

func BenchmarkMapInit(b *testing.B) {
	t0 := time.Now()
	orderedMaps := make([]map[string]bool, 100000)
	for i := 0; i < len(orderedMaps); i++ {
		orderedMaps[i] = make(map[string]bool)
	}
	fmt.Println("Initialized  "+fmt.Sprint(len(orderedMaps)), "OrderedMap in", time.Since(t0))
}

func BenchmarkShrinkSlice(b *testing.B) {
	strs := make([]string, 100000)
	alice := AliceAccount()
	for i := 0; i < len(strs); i++ {
		strs[i] = "blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-0" + fmt.Sprint(i)
	}

	t0 := time.Now()
	for i := 0; i < len(strs)/10; i++ {
		idx := rand.Int() % (len(strs) - 2)
		copy(strs[idx:len(strs)-2], strs[idx+1:len(strs)-1])
	}
	fmt.Println("Remove random element from a Slice ", "from", 100000, "to", len(strs), "in", time.Since(t0))
}

func BenchmarkEncodeTransitions(b *testing.B) {
	store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)

	alice := AliceAccount()
	url.NewAccount(ccurlcommon.SYSTEM, alice)
	// acctTrans := indexer.Univalues(common.Clone(url.Export())).To(indexer.ITCAccess{})

	acctTrans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCAccess{})

	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))

	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	path := commutative.NewPath()
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	t0 := time.Now()
	for i := 0; i < 100000; i++ {
		url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-"+fmt.Sprint(i), noncommutative.NewInt64(int64(i)))
	}
	fmt.Println("Write "+fmt.Sprint(10000), time.Since(t0))

	acctTrans = indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCAccess{})

	t0 = time.Now()
	indexer.Univalues(acctTrans).Encode()
	fmt.Println("Encode "+fmt.Sprint(len(acctTrans)), time.Since(t0))

	/* Forward Iter */
	// t0 = time.Now()
	// v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	// for i := 0; i < 100000; i++ {
	// 	v.(*commutative.Path).Next()
	// }
	// fmt.Println("Next "+fmt.Sprint(100000), time.Since(t0))

	// v.(*commutative.Path).ResetIterator()
	// for i := 0; i < 100000; i++ {
	// 	v.(*commutative.Path).Next()
	// }

	// for i := 0; i < 100000; i++ {
	// 	v.(*commutative.Path).Previous()
	// }

	// v.(*commutative.Path).ResetReverseIterator()
	// for i := 0; i < 100000; i++ {
	// 	v.(*commutative.Path).Previous()
	// }
}

func BenchmarkAccountCreationWithMerkle(b *testing.B) {
	// lut := datacompression.NewCompressionLut()
	// fileDB, err := cachedstorage.NewFileDB(ROOT_PATH, 8, 2)
	// if err != nil {
	// 	b.Error(err)
	// 	return
	// }
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	store.Inject((ccurlcommon.ETH10_ACCOUNT_PREFIX), commutative.NewPath())

	t0 := time.Now()
	url := ccurl.NewConcurrentUrl(store)
	for i := 0; i < 10; i++ {
		acct := datacompression.RandomAccount()
		if _, err := url.NewAccount(0, acct); err != nil { // Preload account structure {
			b.Error(err)
		}
	}
	fmt.Println("Write "+fmt.Sprint(100000*9), time.Since(t0))

	t0 = time.Now()
	acctTrans := indexer.Univalues(common.Clone(url.Export())).To(indexer.ITCTransition{})

	fmt.Println("Export "+fmt.Sprint(100000*9), time.Since(t0))

	t0 = time.Now()

	// transitions := indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues)
	url.Import(acctTrans)
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})
	// errs := url.AllInOneCommit(acctTrans, []uint32{0})

	// if len(errs) > 0 {
	// 	fmt.Println(errs)
	// }
	fmt.Println("Commit + Merkle "+fmt.Sprint(100000*9), time.Since(t0))
}

// func TestOrderedMapBasic(t *testing.T) {
// 	om := orderedmap.NewOrderedMap()
// 	om.Set("abc", 1)
// 	om.Set("xyz", 2)
// 	om.Set("uvw", 3)
// 	om.Set("def", 4)
// 	for iter := om.Front(); iter != nil; iter = iter.Next() {
// 		t.Log(iter.Key, iter.Value)
// 	}
// }

// func TestLLRB(t *testing.T) {
// 	tree := llrb.New()

// 	tree.ReplaceOrInsert(llrb.String("abc"))
// 	tree.ReplaceOrInsert(llrb.String("xyz"))
// 	tree.ReplaceOrInsert(llrb.String("uvw"))
// 	tree.ReplaceOrInsert(llrb.String("def"))

// 	tree.AscendGreaterOrEqual(tree.Min(), func(i llrb.Item) bool {
// 		t.Log(i)
// 		return true
// 	})
// }

// func TestPathRepeats(t *testing.T) {
// 	paths := make([]string, 0, 2)
// 	for i := 0; i < 1; i++ {
// 		acct := datacompression.RandomAccount()
// 		for j := 0; j < 10; j++ {
// 			paths = append(paths, (&ccurl.Platform{}).Eth10Account()+acct+"/"+fmt.Sprint(rand.Float64()))
// 		}
// 	}

// 	positions := make([]int, 0, len(paths))
// 	positions = append(positions, 0)
// 	current := paths[0]
// 	for i := 1; i < len(paths); i++ {
// 		p0 := current[:len((&ccurl.Platform{}).Eth10Account())+ccurlcommon.ETH10_ACCOUNT_LENGTH]
// 		p1 := paths[i][:len((&ccurl.Platform{}).Eth10Account())+ccurlcommon.ETH10_ACCOUNT_LENGTH]
// 		if p0 != p1 {
// 			current = paths[i]
// 			positions = append(positions, i)
// 		}
// 	}
// 	positions = append(positions, len(paths))
// }

func BenchmarkStringSort(b *testing.B) {
	paths := make([][]*univalue.Univalue, 100000)
	for i := 0; i < 100000; i++ {
		acct := datacompression.RandomAccount()
		for j := 9; j >= 1; j-- {

			paths[i] = append(paths[i], univalue.NewUnivalue(uint32(j), acct, 0, 0, 0, noncommutative.NewString(fmt.Sprint(rand.Float64())), nil))
		}
	}

	t0 := time.Now()
	sorter := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			sort.SliceStable(paths[i], func(i, j int) bool {
				if paths[i][j].GetTx() == ccurlcommon.SYSTEM {
					return true
				}

				if paths[i][j].GetTx() == ccurlcommon.SYSTEM {
					return false
				}

				return paths[i][j].GetTx() < paths[i][j].GetTx()
			})
		}
	}
	common.ParallelWorker(len(paths), 6, sorter)
	fmt.Println("Path Sort "+fmt.Sprint(100000*9), time.Since(t0))
}

type String string

func (s String) Less(b btree.Item) bool {
	return s < b.(String)
}

// func BenchmarkOrderedMapPerf(b *testing.B) {
// 	N := 1000000
// 	ss := make([]string, N)
// 	for i := 0; i < N; i++ {
// 		ss[i] = "blcc://eth1.0/account/storage/containers/" + fmt.Sprint(rand.Float64())
// 	}

// 	t0 := time.Now()
// 	gomap := make(map[string]string)
// 	for i := 0; i < N; i++ {
// 		gomap[ss[i]] = ss[i]
// 	}
// 	b.Log("time of go map set:", time.Since(t0))

// 	t0 = time.Now()
// 	tlen0 := 0
// 	for i := 0; i < N; i++ {
// 		tlen0 += len(gomap[ss[i]])
// 	}
// 	b.Log("time of go map get:", time.Since(t0))

// 	t0 = time.Now()
// 	omap := orderedmap.NewOrderedMap()
// 	for i := 0; i < N; i++ {
// 		omap.Set(ss[i], ss[i])
// 	}
// 	b.Log("time of orderedmap set:", time.Since(t0))

// 	t0 = time.Now()
// 	tlen1 := 0
// 	for iter := omap.Front(); iter != nil; iter = iter.Next() {
// 		tlen1 += len(iter.Value.(string))
// 	}
// 	b.Log("time of orderedmap get:", time.Since(t0))

// 	t0 = time.Now()
// 	tree := llrb.New()
// 	for i := 0; i < N; i++ {
// 		tree.ReplaceOrInsert(llrb.String(ss[i]))
// 	}
// 	b.Log("time of llrb insert:", time.Since(t0))

// 	t0 = time.Now()
// 	tlen2 := 0
// 	tree.AscendGreaterOrEqual(tree.Min(), func(i llrb.Item) bool {
// 		tlen2 += len(i.(llrb.String))
// 		return true
// 	})
// 	b.Log("time of llrb get:", time.Since(t0))

// 	t0 = time.Now()
// 	btr := btree.New(32)
// 	for i := 0; i < N; i++ {
// 		btr.ReplaceOrInsert(String(ss[i]))
// 	}
// 	b.Log("time of btree insert:", time.Since(t0))

// 	t0 = time.Now()
// 	tlen3 := 0
// 	btr.AscendGreaterOrEqual(btr.Min(), func(i btree.Item) bool {
// 		tlen3 += len(i.(String))
// 		return true
// 	})
// 	b.Log("time of btree get:", time.Since(t0))

// 	t0 = time.Now()
// 	sort.Strings(ss)
// 	b.Log("time of go sort:", time.Since(t0))

// 	if tlen0 != tlen1 || tlen0 != tlen2 {
// 		b.Fail()
// 	}
// }

// func TestHashPerformance(t *testing.T) {
// 	h1 := fnv1a.HashString64("Hello World!")
// 	fmt.Println("FNV-1a hash of 'Hello World!':", h1)

// 	records := make([]string, 10000)
// 	for i := 0; i < len(records); i++ {
// 		records[i] = (&ccurl.Platform{}).Eth10() + datacompression.RandomAccount()
// 	}

// 	t0 := time.Now()
// 	for i := 0; i < len(records); i++ {
// 		h0, h1 := murmur.Sum128(codec.String(records[i]).Encode())
// 		records[i] = (codec.Bytes(codec.Uint64(h0).Encode()).ToString() + codec.Bytes(codec.Uint64(h1).Encode()).ToString())
// 	}
// 	fmt.Println("murmur "+fmt.Sprint(10000), time.Since(t0))

// 	t0 = time.Now()
// 	for i := 0; i < len(records); i++ {
// 		h0 := fnv1a.HashString64(records[i])
// 		records[i] = codec.Bytes(codec.Uint64(h0).Encode()).ToString() + codec.Bytes(codec.Uint64(h0).Encode()).ToString()

// 	}
// 	fmt.Println("fnv1a "+fmt.Sprint(10000), time.Since(t0))

// 	hash, _ := murmur.Sum128([]byte("FNV-1a hash of 'Hello World!':"))
// 	fmt.Println(hash)
// }

func BenchmarkTransitionImport(b *testing.B) {
	store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	meta := commutative.NewPath()
	store.Inject((ccurlcommon.ETH10_ACCOUNT_PREFIX), meta)

	t0 := time.Now()
	url := ccurl.NewConcurrentUrl(store)
	for i := 0; i < 150000; i++ {
		acct := datacompression.RandomAccount()
		if _, err := url.NewAccount(0, acct); err != nil { // Preload account structure {
			b.Error(err)
		}
	}
	fmt.Println("Write "+fmt.Sprint(100000*9), time.Since(t0))

	t0 = time.Now()
	acctTrans := indexer.Univalues(common.Clone(url.Export())).To(indexer.ITCAccess{})

	fmt.Println("Export "+fmt.Sprint(150000*9), time.Since(t0))

	accountMerkle := indexer.NewAccountMerkle(ccurlcommon.NewPlatform(), rlpEncoder, merkle.Keccak256{}.Hash)

	fmt.Println("-------------")
	t0 = time.Now()
	url.Import(acctTrans)
	accountMerkle.Import(acctTrans)
	fmt.Println("url + accountMerkle Import "+fmt.Sprint(150000*9), time.Since(t0))
}

// func BenchmarkConcurrentTransitionImport(b *testing.B) {
// 	store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
// 	meta := commutative.NewPath()
// 	store.Inject((ccurlcommon.ETH10_ACCOUNT_PREFIX), meta)

// 	t0 := time.Now()
// 	url := ccurl.NewConcurrentUrl(store)
// 	for i := 0; i < 90000; i++ {
// 		acct := datacompression.RandomAccount()
// 		if _, err := url.NewAccount(0, acct); err != nil { // Preload account structure {
// 			b.Error(err)
// 		}
// 	}
// 	fmt.Println("Write "+fmt.Sprint(100000*9), time.Since(t0))

// 	t0 = time.Now()
// 	acctTrans := indexer.Univalues(common.Clone(url.Export())).To(indexer.ITCAccess{})

// 	fmt.Println("Export "+fmt.Sprint(150000*9), time.Since(t0))

// 	accountMerkle := indexer.NewAccountMerkle(ccurlcommon.NewPlatform(), rlpEncoder, merkle.Keccak256{}.Hash)

// 	t0 = time.Now()
// 	common.ParallelExecute(
// 		func() { url.Import(acctTrans) },
// 		func() { accountMerkle.Import(acctTrans) },
// 	)
// 	fmt.Println("ParallelExecute Import "+fmt.Sprint(150000*9), time.Since(t0))
// }

func BenchmarkRandomAccountSort(t *testing.B) {
	store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	meta := commutative.NewPath()
	store.Inject((ccurlcommon.ETH10_ACCOUNT_PREFIX), meta)

	t0 := time.Now()
	url := ccurl.NewConcurrentUrl(store)
	for i := 0; i < 100000; i++ {
		acct := datacompression.RandomAccount()
		if _, err := url.NewAccount(0, acct); err != nil { // Preload account structure {
			// b.Error(err)
		}
	}
	fmt.Println("Write "+fmt.Sprint(100000*9), time.Since(t0))

	t0 = time.Now()
	in := indexer.Univalues(common.Clone(url.Export())).To(indexer.ITCAccess{})

	t0 = time.Now()
	indexer.Univalues(in).Sort(nil)
	fmt.Println("Univalues(in).Sort()", len(in), "entires in :", time.Since(t0))

	// t0 = time.Now()
	// indexer.Univalues(in).SortByDefault()
	// fmt.Println("Univalues(in).SortByDefault()", len(in), "entires in :", time.Since(t0))

	// t0 = time.Now()
	// indexer.Univalues(in).SortWithQuickMethod()
	// fmt.Println("Univalues(in).SortWithQuickMethod()", len(in), "entires in :", time.Since(t0))

}
