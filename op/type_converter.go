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

package opadapter

import (
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/arcology-network/common-lib/exp/array"
	storage "github.com/arcology-network/concurrenturl/storage"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type Convertible storage.AccountResult

// Convert from Ethereum storage proof to Optimism proof format.
// The main difference is that the Ethereum storage proof is a list of hex STRINGS with the 0x prefix, but
// the Optimism storage proof is a list of hex BYTES without the 0x prefix.
func (Convertible) ToStorageProof(res []storage.StorageResult) []OptimismStorageProof {
	opProof := array.Append(res, func(i int, storageResult storage.StorageResult) OptimismStorageProof {
		// decoded := []byte{}
		// if err := rlp.DecodeBytes(storageResult.Value.ToInt().Bytes(), &decoded); err != nil {
		// 	decoded = storageResult.Value.ToInt().Bytes()
		// }

		return OptimismStorageProof{
			Key:   ethcommon.BytesToHash(hexutil.MustDecode(storageResult.Key)),
			Value: *storageResult.Value, // hexutil.Big(*new(big.Int).SetBytes()),
			Proof: array.Append(storageResult.Proof, func(i int, hexStr string) hexutil.Bytes {
				return hexutil.MustDecode(hexStr) // strip 0x prefix and decode hex string to bytes
			}),
		}
	})
	return opProof
}

// Convert from Ethereum account proof to Optimism proof format.
func (this Convertible) New() OptimismAccountResult {
	return OptimismAccountResult{
		AccountProof: array.Append(this.AccountProof, func(i int, hexStr string) hexutil.Bytes {
			return hexutil.MustDecode(hexStr) // strip 0x prefix and decode hex string to bytes
		}),
		Address:      this.Address,
		Balance:      this.Balance,
		CodeHash:     this.CodeHash,
		Nonce:        this.Nonce,
		StorageHash:  this.StorageHash,
		StorageProof: this.ToStorageProof(this.StorageProof),
	}
}
