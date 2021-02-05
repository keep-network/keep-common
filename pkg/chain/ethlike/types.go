package ethlike

import (
	"context"
	"math/big"
)

type Hash interface {
	TerminalString() string
}

type Address interface {
	Hex() string
}

type Receipt interface {
	Status() uint64

	BlockNumber() *big.Int
}

type Transaction interface {
	Hash() Hash

	GasPrice() *big.Int
}

type DeployBackend interface {
	TransactionReceipt(ctx context.Context, txHash Hash) (Receipt, error)

	CodeAt(
		ctx context.Context,
		account Address,
		blockNumber *big.Int,
	) ([]byte, error)
}
