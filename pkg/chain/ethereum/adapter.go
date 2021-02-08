package ethereum

import (
	"context"
	"github.com/ethereum/go-ethereum"
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
	delegate ethclient.Client
}

func (ca *ClientAdapter) CodeAt(
	ctx context.Context,
	contract ethlike.Address,
	blockNumber *big.Int,
) ([]byte, error) {
	return ca.delegate.CodeAt(
		ctx,
		contract.(common.Address),
		blockNumber,
	)
}

func (ca *ClientAdapter) CallContract(
	ctx context.Context,
	call ethlike.CallMsg,
	blockNumber *big.Int,
) ([]byte, error) {
	return ca.delegate.CallContract(
		ctx,
		call.(ethereum.CallMsg),
		blockNumber,
	)
}

func (ca *ClientAdapter) PendingCodeAt(
	ctx context.Context,
	account ethlike.Address,
) ([]byte, error) {
	return ca.delegate.PendingCodeAt(
		ctx,
		account.(common.Address),
	)
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

func (ca *ClientAdapter) SuggestGasPrice(
	ctx context.Context,
) (*big.Int, error) {
	return ca.delegate.SuggestGasPrice(ctx)
}

func (ca *ClientAdapter) EstimateGas(
	ctx context.Context,
	call ethlike.CallMsg,
) (uint64, error) {
	return ca.delegate.EstimateGas(ctx, call.(ethereum.CallMsg))
}

func (ca *ClientAdapter) SendTransaction(
	ctx context.Context,
	tx ethlike.Transaction,
) error {
	return ca.delegate.SendTransaction(ctx, tx.(*types.Transaction))
}

func (ca *ClientAdapter) FilterLogs(
	ctx context.Context,
	query ethlike.FilterQuery,
) ([]ethlike.Log, error) {
	logs, err := ca.delegate.FilterLogs(ctx, query.(ethereum.FilterQuery))
	if err != nil {
		return nil, err
	}

	result := make([]ethlike.Log, 0)
	for _, log := range logs {
		result = append(result, log)
	}

	return result, nil
}

func (ca *ClientAdapter) SubscribeFilterLogs(
	ctx context.Context,
	query ethlike.FilterQuery,
	logChan chan<- ethlike.Log,
) (ethlike.Subscription, error) {
	internalChan := make(chan types.Log)

	go func() {
		for {
			select {
			case log := <-internalChan:
				logChan <- log
			case <-ctx.Done():
				return
			}
		}
	}()

	return ca.delegate.SubscribeFilterLogs(
		ctx,
		query.(ethereum.FilterQuery),
		internalChan,
	)
}

func (ca *ClientAdapter) BlockByHash(
	ctx context.Context,
	hash ethlike.Hash,
) (ethlike.Block, error) {
	return ca.delegate.BlockByHash(ctx, hash.(common.Hash))
}

func (ca *ClientAdapter) BlockByNumber(
	ctx context.Context,
	number *big.Int,
) (ethlike.Block, error) {
	return ca.delegate.BlockByNumber(ctx, number)
}

func (ca *ClientAdapter) HeaderByHash(
	ctx context.Context,
	hash ethlike.Hash,
) (ethlike.Header, error) {
	header, err := ca.delegate.HeaderByHash(ctx, hash.(common.Hash))
	if err != nil {
		return nil, err
	}

	return &headerAdapter{header}, nil
}

func (ca *ClientAdapter) HeaderByNumber(
	ctx context.Context,
	number *big.Int,
) (ethlike.Header, error) {
	header, err := ca.delegate.HeaderByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	return &headerAdapter{header}, nil
}

func (ca *ClientAdapter) TransactionCount(
	ctx context.Context,
	blockHash ethlike.Hash,
) (uint, error) {
	return ca.delegate.TransactionCount(ctx, blockHash.(common.Hash))
}

func (ca *ClientAdapter) TransactionInBlock(
	ctx context.Context,
	blockHash ethlike.Hash,
	index uint,
) (ethlike.Transaction, error) {
	return ca.delegate.TransactionInBlock(ctx, blockHash.(common.Hash), index)
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

func (ca *ClientAdapter) TransactionByHash(
	ctx context.Context,
	txHash ethlike.Hash,
) (ethlike.Transaction, bool, error) {
	return ca.delegate.TransactionByHash(ctx, txHash.(common.Hash))
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

func (ca *ClientAdapter) BalanceAt(
	ctx context.Context,
	account ethlike.Address,
	blockNumber *big.Int,
) (*big.Int, error) {
	return ca.delegate.BalanceAt(ctx, account.(common.Address), blockNumber)
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
