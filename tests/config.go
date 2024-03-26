package committertest

import (
	ethstg "github.com/arcology-network/storage-committer/ethstorage"
	"github.com/arcology-network/storage-committer/interfaces"
	storage "github.com/arcology-network/storage-committer/storage"
	// trie "github.com/ethereum/go-ethereum/trie"
)

const (
	ROOT_PATH   = "./tmp/filedb/"
	BACKUP_PATH = "./tmp/filedb-back/"
)

var (
	// encoder = platform.Codec{}.Encode
	// decoder = platform.Codec{}.Decode

	encoder = ethstg.Rlp{}.Encode
	decoder = ethstg.Rlp{}.Decode
)

func chooseDataStore() interfaces.Datastore {
	// return storage.NewParallelEthMemDataStore() // Eth trie datastore
	// return storage.NewHybirdStore() // Eth trie datastore
	store := storage.NewHybirdStore()
	store.DisableCache()
	return store
	// return storage.NewLevelDBDataStore("/tmp")
	// return datastore.NewDataStore(nil, datastore.NewCachePolicy(1000000, 1), memdb.NewMemoryDB(), encoder, decoder)
	// return storage.NewDataStore(nil, storage.NewCachePolicy(0, 1), storage.NewMemoryDB(), encoder, decoder)
}
