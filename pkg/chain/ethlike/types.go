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
}

type ChainReader interface {
	BlockByNumber(ctx context.Context, number *big.Int) (Block, error)

	SubscribeNewHead(
		ctx context.Context,
		ch chan<- Header,
	) (Subscription, error)
}

type ContractTransactor interface {
	PendingNonceAt(ctx context.Context, account Address) (uint64, error)
}
