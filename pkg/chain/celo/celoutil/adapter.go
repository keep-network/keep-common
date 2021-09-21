package celoutil

import (
	"context"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
)

type ethlikeAdapter struct {
	delegate CeloClient
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
			// TODO: Set the base fee.
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
