package celoutil

import (
	"context"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
)

type BlockSourceAdapter struct {
	delegate CeloClient
}

func NewBlockSourceAdapter(delegate CeloClient) *BlockSourceAdapter {
	return &BlockSourceAdapter{delegate}
}

func (bsa *BlockSourceAdapter) LatestBlock(
	ctx context.Context,
) (*big.Int, error) {
	block, err := bsa.delegate.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	return block.Number(), err
}

func (bsa *BlockSourceAdapter) SubscribeNewBlocks(
	ctx context.Context,
	blocksChan chan<- *big.Int,
) (ethlike.Subscription, error) {
	headersChan := make(chan *types.Header)

	subscription, err := bsa.delegate.SubscribeNewHead(ctx, headersChan)
	if err != nil {
		return nil, err
	}

	stop := make(chan struct{})

	go func() {
		for {
			select {
			case header := <-headersChan:
				blocksChan <- header.Number
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

type TransactionSourceAdapter struct {
	delegate CeloClient
}

func NewTransactionSourceAdapter(
	delegate CeloClient,
) *TransactionSourceAdapter {
	return &TransactionSourceAdapter{delegate}
}

func (tsa *TransactionSourceAdapter) TransactionReceipt(
	ctx context.Context,
	txHash string,
) (*ethlike.TransactionReceipt, error) {
	receipt, err := tsa.delegate.TransactionReceipt(
		ctx,
		common.HexToHash(txHash),
	)
	if err != nil {
		return nil, err
	}

	return &ethlike.TransactionReceipt{
		Status:      receipt.Status,
		BlockNumber: receipt.BlockNumber,
	}, nil
}

type NonceSourceAdapter struct {
	delegate CeloClient
}

func NewNonceSourceAdapter(delegate CeloClient) *NonceSourceAdapter {
	return &NonceSourceAdapter{delegate}
}

func (nsa *NonceSourceAdapter) PendingNonceAt(
	ctx context.Context,
	account string,
) (uint64, error) {
	return nsa.delegate.PendingNonceAt(ctx, common.HexToAddress(account))
}
