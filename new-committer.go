/*
 *   Copyright (c) 2024 Arcology Network

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

// Package storagecommitter provides functionality for committing storage changes to url2a datastore.
package storagecommitter

import (
	"math"

	platform "github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/storage"
	"github.com/arcology-network/storage-committer/univalue"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/arcology-network/common-lib/exp/associative"
	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/slice"
	importer "github.com/arcology-network/storage-committer/importer"
	interfaces "github.com/arcology-network/storage-committer/interfaces"
	intf "github.com/arcology-network/storage-committer/interfaces"
)

// StateCommitterV2 represents a storage committer.
type StateCommitterV2 struct {
	store interfaces.Datastore
	// store    storage.StoreRouter
	Platform *platform.Platform

	byPath *Indexer[string, *univalue.Univalue, []*univalue.Univalue]
	byTxID *Indexer[uint32, *univalue.Univalue, []*univalue.Univalue]
	byEth  *Indexer[[20]byte, *univalue.Univalue, *associative.Pair[*storage.Account, []*univalue.Univalue]]
	byCtrn *Indexer[string, *univalue.Univalue, interface{}]
}

// NewStorageCommitter creates a new StateCommitterV2 instance.
func NewStorageCommitterV2(store interfaces.Datastore) *StateCommitterV2 {
	plat := platform.NewPlatform()

	return &StateCommitterV2{
		store:    store,
		Platform: plat, //[]stgcommcommon.FilteredTransitionsInterface{&importer.NonceFilter{}, &importer.BalanceFilter{}},

		// An index by path, transitions have the same path will be put together in a list
		// This index will be used for apply transitions on the original state. So all the transitions
		// should be put into this index.
		byPath: NewIndexer(
			store,
			func(v *univalue.Univalue) (string, bool) {
				return *(*v).GetPath(), true
			},
			nil,
			func(_ string, v *univalue.Univalue) []*univalue.Univalue { return []*univalue.Univalue{v} },
			func(_ string, v *univalue.Univalue, vals *[]*univalue.Univalue) { *vals = append(*vals, v) },
		),
		// An index by tx number, transitions have the same tx number will be put together in a list.
		// This index will be used to remove the transitions generated by the conflicting transactions.
		// So, the immutable transitions should not be put into this index.
		byTxID: NewIndexer(
			store,
			func(v *univalue.Univalue) (uint32, bool) {
				if !v.Persistent() {
					return v.GetTx(), true
				}
				return math.MaxUint32, false
			},
			nil,
			func(_ uint32, v *univalue.Univalue) []*univalue.Univalue { return []*univalue.Univalue{v} },
			func(_ uint32, v *univalue.Univalue, vals *[]*univalue.Univalue) { *vals = append(*vals, v) },
		),

		// An index by account address, transitions have the same account address will be put together in a list
		// This is for ETH storage, concurrent container related sub-paths won't be put into this index.
		byEth: NewIndexer(
			store,
			func(v *univalue.Univalue) ([20]byte, bool) {
				if !platform.IsEthPath(*v.GetPath()) {
					return [20]byte{}, false
				}
				addr, _ := hexutil.Decode(platform.GetAccountAddr(*v.GetPath()))
				return ethcommon.BytesToAddress(addr), platform.IsEthPath(*v.GetPath())
			},
			nil,
			func(addr [20]byte, v *univalue.Univalue) *associative.Pair[*storage.Account, []*univalue.Univalue] {
				return &associative.Pair[*storage.Account, []*univalue.Univalue]{
					First:  store.Preload(addr[:]).(*storage.Account),
					Second: []*univalue.Univalue{v},
				}
			},
			func(_ [20]byte, v *univalue.Univalue, pair **associative.Pair[*storage.Account, []*univalue.Univalue]) {
				(**pair).Second = append((**pair).Second, v)
			},
		),

		// This index records the transitions that are related to the concurrent containers
		// The transitions will be put into the index if the account address is a concurrent container and
		// later stored in the concurrent container storage, which is different from the ETH storage.
		// All the transitions will be under the same key.
		byCtrn: NewIndexer(
			store,
			func(v *univalue.Univalue) (string, bool) {
				return *v.GetPath(), !platform.IsEthPath(*v.GetPath()) // All under the same key
			},
			nil,
			func(_ string, v *univalue.Univalue) interface{} { return v },
			func(_ string, v *univalue.Univalue, oldv *interface{}) { (*oldv) = v },
		),
	}
}

// New creates a new StateCommitterV2 instance.
func (this *StateCommitterV2) New(args ...interface{}) *StateCommitterV2 {
	return &StateCommitterV2{
		Platform: platform.NewPlatform(),
	}
}

// Importer returns the importer of the StateCommitterV2.
func (this *StateCommitterV2) Store() interfaces.Datastore { return this.store }

// Init initializes the StateCommitterV2 with the given datastore.
// func (this *StateCommitterV2) Init(store interfaces.Datastore) {
// 	this.importer.Init(store)
// 	this.imuImporter.Init(store)
// }

// Import imports the given transitions into the StateCommitterV2.
func (this *StateCommitterV2) Import(transitions []*univalue.Univalue, args ...interface{}) *StateCommitterV2 {
	this.byPath.Add(transitions)
	this.byTxID.Add(transitions)
	this.byEth.Add(transitions)
	this.byCtrn.Add(transitions)
	return this
}

// Finalize finalizes the transitions in the StateCommitterV2.
func (this *StateCommitterV2) Whitelist(txs []uint32) *StateCommitterV2 {
	if len(txs) == 0 {
		return this
	}

	whitelistDict := mapi.FromSlice(txs, func(_ uint32) bool { return true })
	this.byTxID.ParallelForeachDo(func(txid uint32, vec []*univalue.Univalue) {
		if _, ok := whitelistDict[uint32(txid)]; !ok {
			for _, v := range vec {
				v.SetPath(nil) // Mark the transition status, so that it can be removed later.
			}
		}
	})
	return this
}

// Finalize finalizes the transitions in the StateCommitterV2.
func (this *StateCommitterV2) Finalize(txs []uint32) *StateCommitterV2 {
	this.byPath.ParallelForeachDo(func(_ string, v []*univalue.Univalue) {
		importer.DeltaSequenceV2(v).Finalize()
	})
	return this
}

// Commit commits the transitions in the StateCommitterV2.
func (this *StateCommitterV2) Precommit(txs []uint32) [32]byte {
	this.byPath.ParallelForeachDo(func(_ string, v []*univalue.Univalue) {
		importer.DeltaSequenceV2(v).Finalize() // Finalize all the transitions
	})

	// Write the concurrent transitions to the concurrent container storage. All in the same slice.
	this.Store().(*storage.StoreRouter).CCStore().PrecommitV2(this.byCtrn.Keys(), this.byCtrn.Values())

	// Write Eth transitions to the Eth storage
	return this.Store().(*storage.StoreRouter).EthStore().PrecommitV2(this.byEth.Values()) // Write to the DB buffer
}

// Commit commits the transitions in the StateCommitterV2.
func (this *StateCommitterV2) Commit(blockNum uint64) *StateCommitterV2 {
	keys := this.byPath.Keys()
	typedVals := slice.Transform(this.byPath.Values(), func(_ int, v []*univalue.Univalue) intf.Type {
		return v[0].Value().(intf.Type)
	})

	this.Store().(*storage.StoreRouter).RefreshCache(blockNum, keys, typedVals) // Update the cache
	this.Store().(*storage.StoreRouter).CCStore().Commit(0)                     // Write the container storage
	this.Store().(*storage.StoreRouter).EthStore().Commit(0)                    // Write the Eth storage
	return this
}

// Clear clears the StateCommitterV2.
func (this *StateCommitterV2) Clear() {
	this.byPath.Clear()
	this.byTxID.Clear()
	this.byEth.Clear()
	this.byCtrn.Clear()
}