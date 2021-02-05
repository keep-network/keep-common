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

type Header interface {
	Number() *big.Int
}

type Block interface {
	Number() *big.Int

	NumberU64() uint64
}

type Receipt interface {
	Status() uint64

	BlockNumber() *big.Int
}

type Transaction interface {
	Hash() Hash

	GasPrice() *big.Int
}

type Subscription interface {
	Unsubscribe()
	Err() <-chan error
}

type DeployBackend interface {
	TransactionReceipt(ctx context.Context, txHash Hash) (Receipt, error)

	CodeAt(
		ctx context.Context,
		account Address,
		blockNumber *big.Int,
	) ([]byte, error)
}

type ChainReader interface {
	BlockByHash(ctx context.Context, hash Hash) (Block, error)

	BlockByNumber(ctx context.Context, number *big.Int) (Block, error)

	HeaderByHash(ctx context.Context, hash Hash) (Header, error)

	HeaderByNumber(ctx context.Context, number *big.Int) (Header, error)

	TransactionCount(ctx context.Context, blockHash Hash) (uint, error)

	TransactionInBlock(
		ctx context.Context,
		blockHash Hash,
		index uint,
	) (Transaction, error)

	SubscribeNewHead(
		ctx context.Context,
		ch chan<- Header,
	) (Subscription, error)
}
