package ethlike

import (
	"context"
	"encoding/hex"
	"math/big"
)

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [32]byte

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// Address represents the 20 byte address of an ETH-like account.
type Address [20]byte

func (a Address) String() string {
	return hex.EncodeToString(a[:])
}

// Header represents a block header in the ETH-like blockchain.
type Header struct {
	Number *big.Int
}

// Block represents an entire block in the ETH-like blockchain.
type Block struct {
	*Header
}

// Transaction represents an ETH-like chain transaction.
type Transaction struct {
	Hash     Hash
	GasPrice *big.Int
}

// Receipt represents the results of a transaction.
type Receipt struct {
	Status      uint64
	BlockNumber *big.Int
}

// Subscription represents an event subscription where events are delivered
// on a data channel.
type Subscription interface {
	// Unsubscribe cancels the sending of events to the data channel and closes
	// the error channel.
	Unsubscribe()

	// Err returns the subscription error channel. The error channel receives
	// a value if there is an issue with the subscription. Only one value will
	// ever be sent. The error channel is closed by Unsubscribe.
	Err() <-chan error
}

// ChainReader provides access to the blockchain.
type ChainReader interface {
	// BlockByNumber gets the block by its number. The block number argument
	// can be nil to select the latest block.
	BlockByNumber(ctx context.Context, number *big.Int) (*Block, error)

	// SubscribeNewHead subscribes to notifications about changes of the
	// head block of the canonical chain.
	SubscribeNewHead(
		ctx context.Context,
		ch chan<- *Header,
	) (Subscription, error)
}

// TransactionReader provides access to past transactions and their receipts.
type TransactionReader interface {
	// TransactionReceipt returns the receipt of a mined transaction.
	// Note that the transaction may not be included in the current canonical
	// chain even if a receipt exists.
	TransactionReceipt(ctx context.Context, txHash Hash) (*Receipt, error)
}

// ContractTransactor defines the methods needed to allow operating with
// contract on a write only basis.
type ContractTransactor interface {
	// PendingNonceAt retrieves the current pending nonce associated
	// with an account.
	PendingNonceAt(ctx context.Context, account Address) (uint64, error)
}

// Chain represents an ETH-like chain handle.
type Chain interface {
	ChainReader
	TransactionReader
	ContractTransactor
}
