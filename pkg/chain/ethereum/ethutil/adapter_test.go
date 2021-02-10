package ethutil

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
	"reflect"
	"testing"
	"time"
)

func TestBlockSourceAdapter_LatestBlock(t *testing.T) {
	client := &mockAdaptedEthereumClient{
		blocks: []*big.Int{
			big.NewInt(0),
			big.NewInt(1),
			big.NewInt(2),
		},
	}

	adapter := NewBlockSourceAdapter(client)

	block, err := adapter.LatestBlock(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	expectedBlockNumber := big.NewInt(2)
	if expectedBlockNumber.Cmp(block) != 0 {
		t.Errorf(
			"unexpected last block number\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedBlockNumber,
			block,
		)
	}
}

func TestBlockSourceAdapter_SubscribeNewBlocks(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(
		context.Background(),
		10*time.Millisecond,
	)
	defer cancelCtx()

	client := &mockAdaptedEthereumClient{
		blocks: []*big.Int{
			big.NewInt(0),
			big.NewInt(1),
			big.NewInt(2),
		},
	}
	adapter := NewBlockSourceAdapter(client)

	blocksChan := make(chan *big.Int)
	_, err := adapter.SubscribeNewBlocks(ctx, blocksChan)
	if err != nil {
		t.Fatal(err)
	}

	blocks := make([]*big.Int, 0)

loop:
	for {
		select {
		case block := <-blocksChan:
			blocks = append(blocks, block)

			if len(blocks) == 3 {
				break loop
			}
		case <-ctx.Done():
			t.Fatal("timeout has been exceeded")
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

func TestTransactionSourceAdapter_TransactionReceipt(t *testing.T) {
	client := &mockAdaptedEthereumClient{
		transactions: map[common.Hash]*types.Receipt{
			common.HexToHash("0xFF"): {
				Status:      1,
				BlockNumber: big.NewInt(100),
			},
		},
	}

	adapter := NewTransactionSourceAdapter(client)

	receipt, err := adapter.TransactionReceipt(context.Background(), "0xFF")
	if err != nil {
		t.Fatal(err)
	}

	expectedReceipt := &ethlike.TransactionReceipt{
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

func TestNonceSourceAdapter_PendingNonceAt(t *testing.T) {
	client := &mockAdaptedEthereumClient{
		nonces: map[common.Address]uint64{
			common.HexToAddress("0xFF"): 100,
		},
	}

	adapter := NewNonceSourceAdapter(client)

	nonce, err := adapter.PendingNonceAt(context.Background(), "0xFF")
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

type mockAdaptedEthereumClient struct {
	*mockEthereumClient

	blocks       []*big.Int
	transactions map[common.Hash]*types.Receipt
	nonces       map[common.Address]uint64
}

func (maec *mockAdaptedEthereumClient) BlockByNumber(
	ctx context.Context,
	number *big.Int,
) (*types.Block, error) {
	index := len(maec.blocks) - 1

	if number != nil {
		index = int(number.Int64())
	}

	return types.NewBlockWithHeader(
		&types.Header{Number: maec.blocks[index]},
	), nil
}

func (maec *mockAdaptedEthereumClient) SubscribeNewHead(
	ctx context.Context,
	ch chan<- *types.Header,
) (ethereum.Subscription, error) {
	go func() {
		for _, block := range maec.blocks {
			ch <- &types.Header{Number: block}
		}
	}()

	return &subscriptionWrapper{
		unsubscribeFn: func() {},
		errChan:       make(chan error),
	}, nil
}

func (maec *mockAdaptedEthereumClient) TransactionReceipt(
	ctx context.Context,
	txHash common.Hash,
) (*types.Receipt, error) {
	if tx, ok := maec.transactions[txHash]; ok {
		return tx, nil
	}

	return nil, fmt.Errorf("no tx with given hash")
}

func (maec *mockAdaptedEthereumClient) PendingNonceAt(
	ctx context.Context,
	account common.Address,
) (uint64, error) {
	if nonce, ok := maec.nonces[account]; ok {
		return nonce, nil
	}

	return 0, fmt.Errorf("no nonce for given account")
}
