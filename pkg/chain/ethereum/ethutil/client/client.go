package client

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// ChainClient wraps the core `bind.ContractBackend` interface with
// some other interfaces allowing to expose additional methods provided
// by client implementations.
type ChainClient interface {
	bind.ContractBackend
	ethereum.ChainReader
	ethereum.TransactionReader

	BalanceAt(
		ctx context.Context,
		account common.Address,
		blockNumber *big.Int,
	) (*big.Int, error)
}
