package ethereum

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
)

func TxHashExtractor(tx ethlike.Transaction) ethlike.Hash {
	return tx.(*types.Transaction).Hash()
}

type ClientAdapter struct {
	delegate *ethclient.Client
}

func NewClientAdapter(delegate *ethclient.Client) *ClientAdapter {
	return &ClientAdapter{delegate: delegate}
}

func (ca *ClientAdapter) TransactionReceipt(
	ctx context.Context,
	txHash ethlike.Hash,
) (ethlike.Receipt, error) {
	receipt, err := ca.delegate.TransactionReceipt(ctx, txHash.(common.Hash))
	if err != nil {
		return nil, err
	}

	return &receiptAdapter{receipt}, nil
}

func (ca *ClientAdapter) BlockByNumber(
	ctx context.Context,
	number *big.Int,
) (ethlike.Block, error) {
	return ca.delegate.BlockByNumber(ctx, number)
}

func (ca *ClientAdapter) SubscribeNewHead(
	ctx context.Context,
	headerChan chan<- ethlike.Header,
) (ethlike.Subscription, error) {
	internalChan := make(chan *types.Header)

	go func() {
		for {
			select {
			case header := <-internalChan:
				headerChan <- &headerAdapter{header}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ca.delegate.SubscribeNewHead(ctx, internalChan)
}

func (ca *ClientAdapter) PendingNonceAt(
	ctx context.Context,
	account ethlike.Address,
) (uint64, error) {
	return ca.delegate.PendingNonceAt(
		ctx,
		account.(common.Address),
	)
}

type headerAdapter struct {
	delegate *types.Header
}

func (ha *headerAdapter) Number() *big.Int {
	return ha.delegate.Number
}

type receiptAdapter struct {
	delegate *types.Receipt
}

func (ra *receiptAdapter) Status() uint64 {
	return ra.delegate.Status
}

func (ra *receiptAdapter) BlockNumber() *big.Int {
	return ra.delegate.BlockNumber
}
