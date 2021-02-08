package ethutil

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/keep-network/keep-common/pkg/chain/ethlike/ethlikeutil"
	"math/big"
)

type BlockSourceAdapter struct {
	delegate EthereumClient
}

func NewBlockSourceAdapter(delegate EthereumClient) *BlockSourceAdapter {
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
) (ethlikeutil.Subscription, error) {
	headersChan := make(chan *types.Header)

	go func() {
		for {
			select {
			case header := <-headersChan:
				blocksChan <- header.Number
			case <-ctx.Done():
				return
			}
		}
	}()

	return bsa.delegate.SubscribeNewHead(ctx, headersChan)
}

type TransactionSourceAdapter struct {
	delegate EthereumClient
}

func NewTransactionSourceAdapter(
	delegate EthereumClient,
) *TransactionSourceAdapter {
	return &TransactionSourceAdapter{delegate}
}

func (tsa *TransactionSourceAdapter) TransactionReceipt(
	ctx context.Context,
	txHash string,
) (*ethlikeutil.TransactionReceipt, error) {
	receipt, err := tsa.delegate.TransactionReceipt(
		ctx,
		common.HexToHash(txHash),
	)
	if err != nil {
		return nil, err
	}

	return &ethlikeutil.TransactionReceipt{
		Status:      receipt.Status,
		BlockNumber: receipt.BlockNumber,
	}, nil
}

type NonceSourceAdapter struct {
	delegate EthereumClient
}

func NewNonceSourceAdapter(delegate EthereumClient) *NonceSourceAdapter {
	return &NonceSourceAdapter{delegate}
}

func (nsa *NonceSourceAdapter) PendingNonceAt(
	ctx context.Context,
	account string,
) (uint64, error) {
	return nsa.delegate.PendingNonceAt(ctx, common.HexToAddress(account))
}
