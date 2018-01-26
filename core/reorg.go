// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"math/big"
	mrand "math/rand"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/contracts/casper"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// needReorg returns whether the blockchain needs to reorg
func (bc *BlockChain) needReorg(newBlock *types.Block, newTd *big.Int, newState *state.StateDB) bool {
	if bc.config.IsCasper(newBlock.Number()) {
		reorg, handled := bc.casperReorg(newBlock, newState)
		if handled {
			return reorg
		}
	}
	// default reorg policy
	// If the total difficulty is higher than our known, add it to the canonical chain
	// Second clause in the if statement reduces the vulnerability to selfish mining.
	// Please refer to http://www.cs.cornell.edu/~ie53/publications/btcProcFC.pdf
	localTd := bc.GetTd(bc.currentBlock.Hash(), bc.currentBlock.NumberU64())
	reorg := newTd.Cmp(localTd) > 0
	if !reorg && newTd.Cmp(localTd) == 0 {
		// Split same-difficulty blocks by number, then at random
		reorg = newBlock.NumberU64() < bc.currentBlock.NumberU64() || (newBlock.NumberU64() == bc.currentBlock.NumberU64() && mrand.Float64() < 0.5)
	}

	return reorg
}

func (bc *BlockChain) casperReorg(newBlock *types.Block, newState *state.StateDB) (bool, bool) {
	// Casper for current chain
	state, err := bc.StateAt(bc.currentBlock.Root())
	if err != nil {
		log.Warn("Failed to get current chain state", "err", err)
		return false, true
	}
	stateBackend := NewStateBackend(bc.currentBlock, state, bc)
	contract, err := casper.New(stateBackend)
	if err != nil {
		log.Warn("Failed to get current Casper contract", "err", err)
		return false, true
	}

	// Casper for new chain
	newStateBackend := NewStateBackend(newBlock, newState, bc)
	newContract, err := casper.New(newStateBackend)
	if err != nil {
		log.Warn("Failed to get new chain state", "err", err)
		return false, true
	}

	justified, err := contract.GetLastJustifiedEpoch(&bind.CallOpts{})
	if err != nil {
		log.Warn("Failed to get current chain status from Casper", "err", err)
		return false, true
	}
	newJustified, err := newContract.GetLastJustifiedEpoch(&bind.CallOpts{})
	if err != nil {
		log.Warn("Failed to get new chain status from Casper", "err", err)
		return false, true
	}
	if justified.Cmp(newJustified) < 0 {
		return true, true
	} else if justified.Cmp(newJustified) > 0 {
		return false, true
	}
	return false, false
}
