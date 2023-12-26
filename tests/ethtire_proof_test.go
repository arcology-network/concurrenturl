package ccurltest

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	ccurlstorage "github.com/arcology-network/concurrenturl/storage"
	storage "github.com/arcology-network/concurrenturl/storage"
	ethcommon "github.com/ethereum/go-ethereum/common"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	ethmpt "github.com/ethereum/go-ethereum/trie"
)

func TestConcurrentDB(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	bob := BobAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, bob); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	for i := 0; i < 1000; i++ {
		hash := ethcommon.BytesToHash(codec.Uint64(uint64(i)).Encode())
		if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), noncommutative.NewString(string(codec.Uint64(i).Encode()))); err != nil {
			t.Error(err)
		}

		if _, err := url.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), noncommutative.NewString(string(codec.Uint64(i).Encode()))); err != nil {
			t.Error(err)
		}
	}

	common.ParallelExecute(
		func() {
			for i := 1000; i < 2000; i++ {
				hash := ethcommon.BytesToHash(codec.Uint64(uint64(i)).Encode())
				if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), noncommutative.NewString("124")); err != nil {
					t.Error(err)
				}

				if _, err := url.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), noncommutative.NewString("124")); err != nil {
					t.Error(err)
				}
				time.Sleep(5 * time.Millisecond)
			}
			// },
			// func() {
			for i := 0; i < 1000; i++ {
				hash := ethcommon.BytesToHash(codec.Uint64(uint64(i)).Encode())
				if v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), new(noncommutative.String)); v != string(codec.Uint64(i).Encode()) {
					t.Error("Mismatch")
				}

				if v, _ := url.Read(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/"+hexutil.Encode(hash[:]), new(noncommutative.String)); v != string(codec.Uint64(i).Encode()) {
					t.Error("Mismatch")
				}
			}

		})
}

func TestEthWorldTrieProof(t *testing.T) {
	store := chooseDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	bob := BobAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, bob); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	url.Init(store)
	path := commutative.NewPath()
	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	if _, err := url.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	acctTrans = indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	url.Init(store)

	// Delete an non-existing entry, should NOT appear in the transitions
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil); err == nil {
		t.Error("Deleting an non-existing entry should've flaged an error", err)
	}

	raw := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.ITCTransition{})
	if acctTrans := raw; len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}

	// Delete an non-existing entry, should NOT appear in the transitions
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", noncommutative.NewString("124")); err != nil {
		t.Error("Failed to write", err)
	}

	if v, err := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/4", nil); v != "124" {
		t.Error("Wrong return value", err)
	}

	if v, _ := url.Read(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", new(noncommutative.String)); v != "124" {
		t.Error("Error: Wrong return value")
	}

	/* Bob updates */
	// if _, err := url.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
	// 	t.Error(err)
	// }

	// if _, err := url.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/ele1", noncommutative.NewString("6789")); err != nil {
	// 	t.Error(err)
	// }

	acctTrans = indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.ITCTransition{})
	url.Import(indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues))
	url.Sort()
	url.Commit([]uint32{1})

	url.Init(store)

	/* Account Proofs */
	dstore := url.Importer().Store().(*storage.EthDataStore)
	if _, err := dstore.IsProvable((alice)); err != nil {
		t.Error(err)
	}

	if d, err := dstore.IsProvable((bob)); err != nil || len(d) == 0 {
		t.Error(err)
	}

	if _, err := dstore.IsProvable(CarolAccount()); err == nil {
		t.Error("Error: Should've flagged an error")
	}

	kstr, _ := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000000")
	hash := ethcommon.BytesToHash(kstr)

	bobCache, _ := dstore.GetAccount(bob, &ethmpt.AccessListCache{})
	if _, err := bobCache.IsProvable((hash)); err != nil {
		t.Error(err)
	}

	bobTrie, _ := dstore.GetAccountFromTrie(bob, &ethmpt.AccessListCache{})
	if _, err := bobTrie.IsProvable((hash)); err != nil {
		t.Error(err)
	}
}

func TestGetProofAPI(t *testing.T) {
	url := ccurl.NewConcurrentUrl(ccurlstorage.NewParallelEthMemDataStore())

	alice := AliceAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	bob := BobAccount()
	if _, err := url.NewAccount(ccurlcommon.SYSTEM, bob); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.IPCTransition{})
	ts := indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues)
	url.Import(ts)
	url.Sort()
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	/* Alice updates */
	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ele0", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000009", noncommutative.NewInt64(1111)); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	/* Bob updates */
	if _, err := url.Write(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000001", noncommutative.NewInt64(9999)); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := url.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-0/ele1", noncommutative.NewString("6789")); err != nil {
		t.Error(err)
	}

	acctTrans = indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.IPCTransition{})
	ts = indexer.Univalues{}.Decode(indexer.Univalues(acctTrans).Encode()).(indexer.Univalues)
	url.Import(ts)
	url.Sort()
	url.Commit([]uint32{1})

	store := url.Importer().Store().(*storage.EthDataStore)

	/* Get proof direcly */
	bobAcct, _ := store.GetAccount(bob, &ethmpt.AccessListCache{})
	kstr, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")

	v, err := bobAcct.Trie().Get(kstr[:])
	if len(v) == 0 {
		t.Error(err)
	}

	hash := ethcommon.BytesToHash(kstr)
	if _, err := bobAcct.IsProvable((hash)); err != nil {
		t.Error(err)
	}

	if v, _ := url.Read(1, "blcc://eth1.0/account/"+bob+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", new(noncommutative.String)); v != "124" {
		t.Error("Error: Wrong return value")
	}

	/* Through API interface */
	roothash := store.Root()

	proof, err := storage.NewMerkleProof(store.EthDB(), roothash)
	if err != nil {
		t.Error(err)
	}

	accountResult, err := proof.GetProof(bob, []string{string("0x0000000000000000000000000000000000000000000000000000000000000000")})
	if accountResult.StorageProof[0].Value.ToInt().Cmp(big.NewInt(0)) == 0 {
		t.Error(err)
	}
}