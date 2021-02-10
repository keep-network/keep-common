package celoutil

import (
	"context"
	celo "github.com/celo-org/celo-blockchain"
	"github.com/celo-org/celo-blockchain/accounts/abi/bind"
	"github.com/celo-org/celo-blockchain/common"
	"math/big"
)

// CeloClient wraps the core `bind.ContractBackend` interface with
// some other interfaces allowing to expose additional methods provided
// by client implementations.
type CeloClient interface {
	bind.ContractBackend
	celo.ChainReader
	celo.TransactionReader

	BalanceAt(
		ctx context.Context,
		account common.Address,
		blockNumber *big.Int,
	) (*big.Int, error)
}
