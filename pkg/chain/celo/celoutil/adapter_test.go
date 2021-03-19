package celoutil

import (
	"context"
	"fmt"
	"github.com/celo-org/celo-blockchain"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
	"reflect"
	"testing"
	"time"
)

func TestEthlikeAdapter_BlockByNumber(t *testing.T) {
	client := &mockCeloClient{
		blocks: []*big.Int{
			big.NewInt(0),
			big.NewInt(1),
			big.NewInt(2),
		},
	}

	adapter := &ethlikeAdapter{client}

	blockOne, err := adapter.BlockByNumber(context.Background(), big.NewInt(1))
	if err != nil {
		t.Fatal(err)
	}

	lastBlock, err := adapter.BlockByNumber(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	expectedBlockOneNumber := big.NewInt(1)
	if expectedBlockOneNumber.Cmp(blockOne.Number) != 0 {
		t.Errorf(
			"unexpected block number\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedBlockOneNumber,
			blockOne.Number,
		)
	}

	expectedLastBlockNumber := big.NewInt(2)
	if expectedLastBlockNumber.Cmp(lastBlock.Number) != 0 {
		t.Errorf(
			"unexpected last block number\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedLastBlockNumber,
			lastBlock.Number,
		)
	}
}

func TestEthlikeAdapter_SubscribeNewHead(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(
		context.Background(),
		10*time.Millisecond,
	)
	defer cancelCtx()

	client := &mockCeloClient{
		blocks: []*big.Int{
			big.NewInt(0),
			big.NewInt(1),
			big.NewInt(2),
		},
	}

	adapter := &ethlikeAdapter{client}

	// no more than 3 elements should be put into this
	// channel by SubscribeNewHead
	headerChan := make(chan *ethlike.Header, 100)
	_, err := adapter.SubscribeNewHead(ctx, headerChan)
	if err != nil {
		t.Fatal(err)
	}

	<-ctx.Done()

	expectedHeaderChanLen := 3
	headerChanLen := len(headerChan)
	if expectedHeaderChanLen != headerChanLen {
		t.Errorf(
			"unexpected number of blocks\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedHeaderChanLen,
			headerChanLen,
		)
	}

	blocks := make([]*big.Int, 0)
	for header := range headerChan {
		blocks = append(blocks, header.Number)

		// headerChan is not closed so we have to break manually
		if len(blocks) == 3 {
			break
		}
	}

	expectedBlocks := []*big.Int{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(2),
	}
	if !reflect.DeepEqual(expectedBlocks, blocks) {
		t.Errorf(
			"unexpected blocks\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedBlocks,
			blocks,
		)
	}
}

func TestEthlikeAdapter_TransactionReceipt(t *testing.T) {
	var hash [32]byte
	copy(hash[:], []byte{255})

	client := &mockCeloClient{
		transactions: map[common.Hash]*types.Receipt{
			common.BytesToHash(hash[:]): {
				Status:      1,
				BlockNumber: big.NewInt(100),
			},
		},
	}

	adapter := &ethlikeAdapter{client}

	receipt, err := adapter.TransactionReceipt(
		context.Background(),
		hash,
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedReceipt := &ethlike.Receipt{
		Status:      1,
		BlockNumber: big.NewInt(100),
	}
	if !reflect.DeepEqual(expectedReceipt, receipt) {
		t.Errorf(
			"unexpected tx receipt\n"+
				"expected: [%+v]\n"+
				"actual:   [%+v]",
			expectedReceipt,
			receipt,
		)
	}
}

func TestEthlikeAdapter_PendingNonceAt(t *testing.T) {
	var address [20]byte
	copy(address[:], []byte{255})

	client := &mockCeloClient{
		nonces: map[common.Address]uint64{
			common.BytesToAddress(address[:]): 100,
		},
	}

	adapter := &ethlikeAdapter{client}

	nonce, err := adapter.PendingNonceAt(context.Background(), address)
	if err != nil {
		t.Fatal(err)
	}

	expectedNonce := uint64(100)
	if expectedNonce != nonce {
		t.Errorf(
			"unexpected nonce\n"+
				"expected: [%+v]\n"+
				"actual:   [%+v]",
			expectedNonce,
			nonce,
		)
	}
}

type mockCeloClient struct {
	blocks       []*big.Int
	transactions map[common.Hash]*types.Receipt
	nonces       map[common.Address]uint64
}

func (mcc *mockCeloClient) CodeAt(
	ctx context.Context,
	contract common.Address,
	blockNumber *big.Int,
) ([]byte, error) {
	panic("implement")
}

func (mcc *mockCeloClient) CallContract(
	ctx context.Context,
	call celo.CallMsg,
	blockNumber *big.Int,
) ([]byte, error) {
	panic("implement")
}

func (mcc *mockCeloClient) PendingCodeAt(
	ctx context.Context,
	account common.Address,
) ([]byte, error) {
	panic("implement")
}

func (mcc *mockCeloClient) PendingNonceAt(
	ctx context.Context,
	account common.Address,
) (uint64, error) {
	if nonce, ok := mcc.nonces[account]; ok {
		return nonce, nil
	}

	return 0, fmt.Errorf("no nonce for given account")
}

func (mcc *mockCeloClient) SuggestGasPrice(
	ctx context.Context,
) (*big.Int, error) {
	panic("implement")
}

func (mcc *mockCeloClient) EstimateGas(
	ctx context.Context,
	call celo.CallMsg,
) (uint64, error) {
	panic("implement")
}

func (mcc *mockCeloClient) SendTransaction(
	ctx context.Context,
	tx *types.Transaction,
) error {
	panic("implement")
}

func (mcc *mockCeloClient) FilterLogs(
	ctx context.Context,
	query celo.FilterQuery,
) ([]types.Log, error) {
	panic("implement")
}

func (mcc *mockCeloClient) SubscribeFilterLogs(
	ctx context.Context,
	query celo.FilterQuery,
	ch chan<- types.Log,
) (celo.Subscription, error) {
	panic("implement")
}

func (mcc *mockCeloClient) BlockByHash(
	ctx context.Context,
	hash common.Hash,
) (*types.Block, error) {
	panic("implement")
}

func (mcc *mockCeloClient) BlockByNumber(
	ctx context.Context,
	number *big.Int,
) (*types.Block, error) {
	index := len(mcc.blocks) - 1

	if number != nil {
		index = int(number.Int64())
	}

	return types.NewBlockWithHeader(
		&types.Header{Number: mcc.blocks[index]},
	), nil
}

func (mcc *mockCeloClient) HeaderByHash(
	ctx context.Context,
	hash common.Hash,
) (*types.Header, error) {
	panic("implement")
}

func (mcc *mockCeloClient) HeaderByNumber(
	ctx context.Context,
	number *big.Int,
) (*types.Header, error) {
	panic("implement")
}

func (mcc *mockCeloClient) TransactionCount(
	ctx context.Context,
	blockHash common.Hash,
) (uint, error) {
	panic("implement")
}

func (mcc *mockCeloClient) TransactionInBlock(
	ctx context.Context,
	blockHash common.Hash,
	index uint,
) (*types.Transaction, error) {
	panic("implement")
}

func (mcc *mockCeloClient) SubscribeNewHead(
	ctx context.Context,
	ch chan<- *types.Header,
) (celo.Subscription, error) {
	go func() {
		for _, block := range mcc.blocks {
			ch <- &types.Header{Number: block}
		}
	}()

	return &subscriptionWrapper{
		unsubscribeFn: func() {},
		errChan:       make(chan error),
	}, nil
}

func (mcc *mockCeloClient) TransactionByHash(
	ctx context.Context,
	txHash common.Hash,
) (*types.Transaction, bool, error) {
	panic("implement")
}

func (mcc *mockCeloClient) TransactionReceipt(
	ctx context.Context,
	txHash common.Hash,
) (*types.Receipt, error) {
	if tx, ok := mcc.transactions[txHash]; ok {
		return tx, nil
	}

	return nil, fmt.Errorf("no tx with given hash")
}

func (mcc *mockCeloClient) BalanceAt(
	ctx context.Context,
	account common.Address,
	blockNumber *big.Int,
) (*big.Int, error) {
	panic("implement")
}
