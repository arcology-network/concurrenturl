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

// Package storagecommitter provides functionality for committing storage changes to url2a datastore.
package storagecommitter

import (
	"fmt"
	"time"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/array"
	structure "github.com/arcology-network/common-lib/exp/product"
	platform "github.com/arcology-network/concurrenturl/platform"
	"github.com/arcology-network/concurrenturl/storage"
	"github.com/arcology-network/concurrenturl/univalue"
	"github.com/ethereum/go-ethereum/common/hexutil"

	importer "github.com/arcology-network/concurrenturl/importer"
	interfaces "github.com/arcology-network/concurrenturl/interfaces"
)

// StateCommitter represents a storage committer.
type StateCommitter struct {
	store       interfaces.Datastore
	importer    *importer.Importer
	imuImporter *importer.Importer // transitions that will take effect anyway regardless of execution failures or conflicts
	Platform    *platform.Platform

	// Have to be defined here to avoid having duplicate addresses in the account list and dictionary.
	acctList []*structure.Triplet[string, *storage.Account, []*univalue.Univalue]
	acctDict map[string]*structure.Triplet[string, *storage.Account, []*univalue.Univalue]
}

// NewStorageCommitter creates a new StateCommitter instance.
func NewStorageCommitter(store interfaces.Datastore) *StateCommitter {
	platform := platform.NewPlatform()
	return &StateCommitter{
		store:       store,
		importer:    importer.NewImporter(store, platform),
		imuImporter: importer.NewImporter(store, platform),
		Platform:    platform, //[]committercommon.FilteredTransitionsInterface{&importer.NonceFilter{}, &importer.BalanceFilter{}},

		acctList: []*structure.Triplet[string, *storage.Account, []*univalue.Univalue]{},
		acctDict: map[string]*structure.Triplet[string, *storage.Account, []*univalue.Univalue]{},
	}
}

// New creates a new StateCommitter instance.
func (this *StateCommitter) New(args ...interface{}) *StateCommitter {
	return &StateCommitter{
		Platform: platform.NewPlatform(),
	}
}

// Importer returns the importer of the StateCommitter.
func (this *StateCommitter) Importer() *importer.Importer { return this.importer }

// Init initializes the StateCommitter with the given datastore.
func (this *StateCommitter) Init(store interfaces.Datastore) {
	this.importer.Init(store)
	this.imuImporter.Init(store)
}

// Clear clears the StateCommitter.
func (this *StateCommitter) Clear() {
	this.importer.Store().Clear()

	t0 := time.Now()
	this.importer.Clear()
	fmt.Println("importer.Clear(): ", time.Since(t0))

	t0 = time.Now()
	this.imuImporter.Clear()
	fmt.Println("imuImporter.Clear(): ", time.Since(t0))
}

// Import imports the given transitions into the StateCommitter.
func (this *StateCommitter) Import(transitions []*univalue.Univalue, args ...interface{}) *StateCommitter {
	this.AddToAccountDict(transitions)

	// Move the Peristent transitions(nonce and gas fee) to another list.
	invTransitions := array.MoveIf(&transitions, func(i int, v *univalue.Univalue) bool {
		return v.Persistent()
	})

	array.Remove(&transitions, nil) // Remove the Peristent transitions from the transition lists
	common.ParallelExecute(
		func() { this.imuImporter.Import(invTransitions, args...) },
		func() { this.importer.Import(transitions, args...) })
	return this
}

// Add the transaction to the account dictionary.
func (this *StateCommitter) AddToAccountDict(transitions []*univalue.Univalue) {
	if !common.IsType[*storage.EthDataStore](this.store) {
		return
	}

	array.Foreach(transitions, func(i int, v **univalue.Univalue) {
		addrString := platform.GetAccountAddr(*(*v).GetPath())
		acctTriplet, ok := this.acctDict[addrString]
		if !ok { // Does not exist
			addr, _ := hexutil.Decode(platform.GetAccountAddr(addrString))
			acctTriplet = &structure.Triplet[string, *storage.Account, []*univalue.Univalue]{ // Create a new triplet
				addrString,
				this.store.(*storage.EthDataStore).NewAccount(addr), // Create an Eth account using the default configuration.
				[]*univalue.Univalue{*v}}
			this.acctDict[addrString] = acctTriplet // Add the triplet to the dictionary
			return
		}
		acctTriplet.Third = append(acctTriplet.Third, *v) // Add the transition to the triplet
	})
}

// Sort sorts the transitions in the StateCommitter.
func (this *StateCommitter) Sort() *StateCommitter {
	common.ParallelExecute(
		func() { this.imuImporter.SortDeltaSequences() },
		func() { this.importer.SortDeltaSequences() })
	return this
}

// Finalize finalizes the transitions in the StateCommitter.
func (this *StateCommitter) Finalize(txs []uint32) *StateCommitter {
	if txs != nil && len(txs) == 0 { // Commit all the transactions when txs == nil
		return this
	}

	common.ParallelExecute(
		func() {
			this.imuImporter.MergeStateDelta()
		},
		func() {
			this.importer.WhiteList(txs)    // Remove all the transitions generated by the conflicting transactions
			this.importer.MergeStateDelta() // Finalize states
		},
	)

	common.ParallelExecute(
		func() {
			this.imuImporter.OrganizeByAccount()
			this.importer.OrganizeByAccount()
		},
		func() {
			this.importer.MergeStateDelta() // Finalize states
		},
	)
	return this
}

// CopyToDbBuffer copies the transitions to the DB buffer.
func (this *StateCommitter) CopyToDbBuffer() ([32]byte, []string, []interface{}) {
	keys, values := this.importer.KVs()
	invKeys, invVals := this.imuImporter.KVs()

	keys, values = append(keys, invKeys...), append(values, invVals...)
	return this.importer.Store().Precommit(keys, values), keys, values // save the transitions to the DB buffer
}

// Commit commits the transitions in the StateCommitter.
func (this *StateCommitter) Precommit(txs []uint32) [32]byte {
	if txs != nil && len(txs) == 0 {
		this.Clear()
		// panic("No transactions to commit")
		return [32]byte{}
	}
	this.Finalize(txs)
	hash, _, _ := this.CopyToDbBuffer()
	return hash // Export transitions and save them to the DB buffer.
}

// Commit commits the transitions in the StateCommitter.
func (this *StateCommitter) Commit() *StateCommitter {
	store := this.importer.Store()
	store.Commit(0) // Commit to the state store
	this.Clear()
	return this
}
