package ethutil

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
	"time"
)

type ethlikeAdapter struct {
	delegate EthereumClient
}

func (ea *ethlikeAdapter) BlockByNumber(
	ctx context.Context,
	number *big.Int,
) (*ethlike.Block, error) {
	block, err := ea.delegate.BlockByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	return &ethlike.Block{
		Header: &ethlike.Header{
			Number: block.Number(),
		},
	}, nil
}

func (ea *ethlikeAdapter) SubscribeNewHead(
	ctx context.Context,
	headersChan chan<- *ethlike.Header,
) (ethlike.Subscription, error) {
	internalHeadersChan := make(chan *types.Header)

	subscription, err := ea.delegate.SubscribeNewHead(ctx, internalHeadersChan)
	if err != nil {
		return nil, err
	}

	stop := make(chan struct{})

	go func() {
		for {
			select {
			case header := <-internalHeadersChan:
				headersChan <- &ethlike.Header{
					Number: header.Number,
				}
			case <-stop:
				return
			}
		}
	}()

	return &subscriptionWrapper{
		unsubscribeFn: func() {
			close(stop)
			subscription.Unsubscribe()
		},
		errChan: subscription.Err(),
	}, nil
}

type subscriptionWrapper struct {
	unsubscribeFn func()
	errChan       <-chan error
}

func (sw *subscriptionWrapper) Unsubscribe() {
	sw.unsubscribeFn()
}

func (sw *subscriptionWrapper) Err() <-chan error {
	return sw.errChan
}

func (ea *ethlikeAdapter) TransactionReceipt(
	ctx context.Context,
	txHash ethlike.Hash,
) (*ethlike.Receipt, error) {
	receipt, err := ea.delegate.TransactionReceipt(
		ctx,
		common.Hash(txHash),
	)
	if err != nil {
		return nil, err
	}

	return &ethlike.Receipt{
		Status:      receipt.Status,
		BlockNumber: receipt.BlockNumber,
	}, nil
}

func (ea *ethlikeAdapter) PendingNonceAt(
	ctx context.Context,
	account ethlike.Address,
) (uint64, error) {
	return ea.delegate.PendingNonceAt(ctx, common.Address(account))
}

// NewBlockCounter creates a new BlockCounter instance for the provided
// Ethereum client.
func NewBlockCounter(client EthereumClient) (*ethlike.BlockCounter, error) {
	return ethlike.CreateBlockCounter(&ethlikeAdapter{client})
}

// NewMiningWaiter creates a new MiningWaiter instance for the provided
// Ethereum client. It accepts two parameters setting up monitoring rules
// of the transaction mining status.
func NewMiningWaiter(
	client EthereumClient,
	checkInterval time.Duration,
	maxGasPrice *big.Int,
) *ethlike.MiningWaiter {
	return ethlike.NewMiningWaiter(
		&ethlikeAdapter{client},
		checkInterval,
		maxGasPrice,
	)
}

// NewNonceManager creates NonceManager instance for the provided account
// using the provided Ethereum client.
func NewNonceManager(
	client EthereumClient,
	account common.Address,
) *ethlike.NonceManager {
	return ethlike.NewNonceManager(
		ethlike.Address(account),
		&ethlikeAdapter{client},
	)
}
