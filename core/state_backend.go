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
	"context"
	"errors"
	"math"
	"math/big"
	"sync"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

var (
	// This nil assignment ensures compile time that StateBackend implements bind.ContractBackend.
	_ bind.ContractBackend = (*StateBackend)(nil)

	// ErrNotImplemented returns if the function is not implemented
	ErrNotImplemented = errors.New("not implemented")
)

// StateBackend implements bind.ContractBackend. Its main purpose is to allow using contract bindings at specific block and state.
type StateBackend struct {
	block      *types.Block
	stateDB    *state.StateDB
	blockchain *BlockChain // Ethereum blockchain to handle the consensus

	mu sync.Mutex
}

// NewStateBackend creates a new binding backend
func NewStateBackend(block *types.Block, stateDB *state.StateDB, blockchain *BlockChain) *StateBackend {
	return &StateBackend{
		block:      block,
		stateDB:    stateDB,
		blockchain: blockchain,
	}
}

// CodeAt returns the code associated with a certain account in the blockchain.
func (b *StateBackend) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.stateDB.GetCode(contract), nil
}

// BalanceAt returns the wei balance of a certain account in the blockchain.
func (b *StateBackend) BalanceAt(ctx context.Context, contract common.Address, blockNumber *big.Int) (*big.Int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.stateDB.GetBalance(contract), nil
}

// NonceAt returns the nonce of a certain account in the blockchain.
func (b *StateBackend) NonceAt(ctx context.Context, contract common.Address, blockNumber *big.Int) (uint64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.stateDB.GetNonce(contract), nil
}

// StorageAt returns the value of key in the storage of an account in the blockchain.
func (b *StateBackend) StorageAt(ctx context.Context, contract common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	val := b.stateDB.GetState(contract, key)
	return val[:], nil
}

// TransactionReceipt returns the receipt of a transaction.
func (b *StateBackend) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return nil, ErrNotImplemented
}

// PendingCodeAt returns the code associated with an account in the pending state.
func (b *StateBackend) PendingCodeAt(ctx context.Context, contract common.Address) ([]byte, error) {
	return nil, ErrNotImplemented
}

// CallContract executes a contract call.
func (b *StateBackend) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Ensure message is initialized properly.
	if call.GasPrice == nil {
		call.GasPrice = common.Big0
	}
	if call.Gas == 0 {
		call.Gas = 50000000
	}
	if call.Value == nil {
		call.Value = common.Big0
	}

	// Create new call message
	msg := types.NewMessage(call.From, call.To, 0, call.Value, call.Gas, call.GasPrice, call.Data, false)
	evmContext := NewEVMContext(msg, b.block.Header(), b.blockchain, nil)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(evmContext, b.stateDB, b.blockchain.Config(), vm.Config{})
	gaspool := new(GasPool).AddGas(math.MaxUint64)
	r, _, _, err := NewStateTransition(vmenv, msg, gaspool).TransitionDb()
	return r, err
}

// PendingCallContract executes a contract call on the pending state.
func (b *StateBackend) PendingCallContract(ctx context.Context, call ethereum.CallMsg) ([]byte, error) {
	return nil, ErrNotImplemented
}

// PendingNonceAt implements PendingStateReader.PendingNonceAt, retrieving
// the nonce currently pending for the account.
func (b *StateBackend) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return 0, ErrNotImplemented
}

// SuggestGasPrice implements ContractTransactor.SuggestGasPrice. Since the simulated
// chain doens't have miners, we just return a gas price of 1 for any call.
func (b *StateBackend) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return nil, ErrNotImplemented
}

// EstimateGas executes the requested code against the currently pending block/state and
// returns the used amount of gas.
func (b *StateBackend) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	return 0, ErrNotImplemented
}

// SendTransaction updates the pending block to include the given transaction.
// It panics if the transaction is invalid.
func (b *StateBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return ErrNotImplemented
}
