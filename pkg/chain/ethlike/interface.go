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

type Log interface {
	// TODO: methods
}

type CallMsg interface {
	// TODO: methods
}

type FilterQuery interface {
	// TODO: methods
}

type DeployBackend interface {
	TransactionReceipt(ctx context.Context, txHash Hash) (Receipt, error)
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

type TransactionReader interface {
	TransactionByHash(
		ctx context.Context,
		txHash Hash,
	) (Transaction, bool, error)

	TransactionReceipt(ctx context.Context, txHash Hash) (Receipt, error)
}

type ContractCaller interface {
	CodeAt(
		ctx context.Context,
		contract Address,
		blockNumber *big.Int,
	) ([]byte, error)

	CallContract(
		ctx context.Context,
		call CallMsg,
		blockNumber *big.Int,
	) ([]byte, error)
}

type ContractTransactor interface {
	PendingCodeAt(ctx context.Context, account Address) ([]byte, error)

	PendingNonceAt(ctx context.Context, account Address) (uint64, error)

	SuggestGasPrice(ctx context.Context) (*big.Int, error)

	EstimateGas(ctx context.Context, call CallMsg) (gas uint64, err error)

	SendTransaction(ctx context.Context, tx Transaction) error
}

type ContractFilterer interface {
	FilterLogs(ctx context.Context, query FilterQuery) ([]Log, error)

	SubscribeFilterLogs(
		ctx context.Context,
		query FilterQuery,
		ch chan<- Log,
	) (Subscription, error)
}

type ContractBackend interface {
	ContractCaller
	ContractTransactor
	ContractFilterer
}

type Client interface {
	ContractBackend
	ChainReader
	TransactionReader

	BalanceAt(
		ctx context.Context,
		account Address,
		blockNumber *big.Int,
	) (*big.Int, error)
}
