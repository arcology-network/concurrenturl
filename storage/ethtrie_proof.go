package storage

import (
	"sync"

	"github.com/arcology-network/common-lib/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethmpt "github.com/ethereum/go-ethereum/trie"
	// ethapi "github.com/ethereum/go-ethereum/internal/ethapi"
)

// MkerkleProofManager is a manager for merkle proofs. It keeps track of the number of times a merkle
// tree has been accessed and keeps the most recently used merkle trees in memory. It a mkerkle tree isn't
// in memory, it will be loaded from the database.
type MerkleProofManager struct {
	maxCached  int                       // max number of merkle proofs to keep in memory
	merkleDict map[[32]byte]*MerkleProof // The merkle tree for each root.
	db         *ethmpt.Database

	lock sync.Mutex
}

// NewMerkleProofManager creates a new MerkleProofManager, which keeps a cache of merkle trees in memory.
// When the cache is full, the merkle tree with the lowest ratio of visits/totalVisits will be removed.
func NewMerkleProofManager(maxCached int, db *ethmpt.Database) *MerkleProofManager {
	return &MerkleProofManager{
		maxCached:  maxCached,
		merkleDict: map[[32]byte]*MerkleProof{},
		db:         db,
	}
}

// GetProof returns a merkle proof for the given account and storage keys.
func (this *MerkleProofManager) GetProof(rootHash [32]byte, acctStr string, storageKeys []string) (*AccountResult, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	merkle, _ := this.merkleDict[rootHash]
	if merkle == nil {
		datastore, err := LoadEthDataStore(this.db, rootHash)
		if err != nil {
			return nil, err
		}

		// Create a new merkle tree and add it to the cache.
		merkle = &MerkleProof{totalVisits: 1, visits: 1, DataStore: datastore, Ethdb: this.db}
		this.merkleDict[rootHash] = merkle

		// Check if the cache is full. Clean up the cache if it is full.
		if len(this.merkleDict) > this.maxCached {
			keys, merkles := common.MapKVs(this.merkleDict)

			// The visit ratio is the number of times a merkle tree has been accessed divided by the total number of times all the merkle trees have been accessed.
			ratios := common.Append(merkles, func(_ int, v *MerkleProof) float64 { return float64(v.visits) / float64(v.totalVisits) })

			// The entry has the lowest ratio of visits/totalVisits will be removed.
			idx, _ := common.MinElement(ratios, func(v0, v1 float64) bool { return v0 < v1 })
			delete(this.merkleDict, keys[idx])
		}
	}

	// Increment the number of visits for all the merkle trees by 1.
	common.MapForeach(this.merkleDict, func(_ [32]byte, v **MerkleProof) { (**v).totalVisits++ })

	return merkle.GetProof(acctStr, storageKeys)
}

type MerkleProof struct {
	totalVisits uint64 // Total number of times all the merkle trees have been accessed since this Merkle tree is created.
	visits      int    // Number of times this merkle Merkle has been accessed.
	DataStore   *EthDataStore
	Ethdb       *ethmpt.Database
}

func NewMerkleProof(ethdb *ethmpt.Database, root [32]byte) (*MerkleProof, error) {
	store, err := LoadEthDataStore(ethdb, root)
	if err != nil {
		return nil, err
	}

	return &MerkleProof{
		1,
		1,
		store,
		ethdb,
	}, nil
}

// GetProof returns a merkle proof for the given account and storage keys.
func (this *MerkleProof) GetProof(acctAddr string, storageKeys []string) (*AccountResult, error) {
	this.visits++

	// Get the account either from the cache or from the database.
	account, err := this.DataStore.GetAccount(acctAddr, new(ethmpt.AccessListCache))
	if err != nil {
		return nil, err
	}

	storageHash := types.EmptyRootHash
	codeHash := account.GetCodeHash()

	if account.storageTrie != nil {
		storageHash = account.storageTrie.Hash()
	} else {
		// no storageTrie means the account does not exist, so the codeHash is the hash of an empty bytearray.
		codeHash = crypto.Keccak256Hash(nil)
	}

	// Create the storage proof for each storage key.
	storageProof := make([]StorageResult, len(storageKeys))
	for i, hexKey := range storageKeys {
		key, err := decodeHash(hexKey)
		if err != nil {
			return nil, err
		}

		if account.storageTrie != nil {
			proof, storageError := account.Prove(key) // Get the storage proof
			if storageError != nil {
				return nil, storageError
			}

			v, _ := account.storageTrie.Get(key[:]) // Get the storage value
			storageProof[i] = StorageResult{hexKey, (*hexutil.Big)(ethcommon.BytesToHash(v).Big()), toHexSlice(proof)}
		} else {
			storageProof[i] = StorageResult{hexKey, &hexutil.Big{}, []string{}}
		}
	}

	// create the account Proof
	acctBytes, _ := hexutil.Decode(acctAddr)
	accountProof, proofErr := this.DataStore.GetAccountProof(acctBytes) // Get the account proof
	if proofErr != nil {
		return nil, proofErr
	}

	return &AccountResult{
		Address:      ethcommon.BytesToAddress([]byte(acctAddr)),
		AccountProof: toHexSlice(accountProof),
		Balance:      (*hexutil.Big)(account.StateAccount.Balance),
		CodeHash:     codeHash,
		Nonce:        hexutil.Uint64(account.StateAccount.Nonce),
		StorageHash:  storageHash,
		StorageProof: storageProof,
	}, nil // state.Error()
}
