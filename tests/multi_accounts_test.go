package ccurltest

import (
	"testing"

	// codec "github.com/arcology-network/common-lib/codec"

	datacompression "github.com/arcology-network/common-lib/addrcompressor"
	"github.com/arcology-network/common-lib/common"
	ccurl "github.com/arcology-network/concurrenturl"
	committercommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	importer "github.com/arcology-network/concurrenturl/importer"
	cache "github.com/arcology-network/eu/cache"
)

func TestMultiAccountCreation(t *testing.T) {
	store := chooseDataStore()
	// store := datastore.NewDataStore(nil, datastore.NewCachePolicy(0, 1), datastore.NewMemDB(), encoder, decoder)

	store.Inject((committercommon.ETH10_ACCOUNT_PREFIX), commutative.NewPath())

	writeCache := cache.NewWriteCache(store, committercommon.NewPlatform())

	accounts := make([]string, 10)
	for i := 0; i < len(accounts); i++ {
		accounts[i] = datacompression.RandomAccount()

		if _, err := writeCache.CreateNewAccount(0, accounts[i]); err != nil { // NewAccount account structure {
			t.Error(err)
		}

		// if _, err := committer.NewAccount(0, accounts[i]); err != nil { // Preload account structure {
		// 	t.Error(err)
		// }
	}
	raw := writeCache.Export(importer.Sorter)
	acctTrans := importer.Univalues(common.Clone(raw)).To(importer.ITCTransition{})

	paths := committercommon.NewPlatform().GetSysPaths()
	if len(acctTrans) != len(paths)*len(accounts) {
		t.Error("Error: Transition counts don't match up")
	}

	committer := ccurl.NewStorageCommitter(store)
	committer.Import(acctTrans)
	committer.Sort()
	committer.Commit([]uint32{0})
	writeCache.Clear()
	// acctTrans = importer.Univalues(common.Clone(raw)).To(importer.ITCTransition{})
	// encoded := importer.Univalues(acctTrans).Encode()

	// out := importer.Univalues{}.Decode(encoded).(importer.Univalues)
	// if len(acctTrans) != len(out) {
	// 	t.Error("Error: Transition counts don't match up")
	// }

	// accountMerkle := importer.NewAccountMerkle(committer.Platform, rlpEncoder, merkle.Keccak256{}.Hash)
	// accountMerkle.Import(out)
	// accountMerkle.Build()
}
