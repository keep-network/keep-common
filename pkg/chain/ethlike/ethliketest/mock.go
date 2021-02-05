package ethliketest

import (
	"context"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
	"sync"
	"time"
)

type MockHash struct {
	Hash string
}

func (mh *MockHash) TerminalString() string {
	return mh.Hash
}

type MockAddress struct {
	Address string
}

func (ma *MockAddress) Hex() string {
	return ma.Address
}

type MockReceipt struct {
	ReceiptStatus      uint64
	ReceiptBlockNumber *big.Int
}

func (mr *MockReceipt) Status() uint64 {
	return mr.ReceiptStatus
}

func (mr *MockReceipt) BlockNumber() *big.Int {
	return mr.ReceiptBlockNumber
}

type MockTransaction struct {
	TxHash     ethlike.Hash
	TxNonce    uint64
	TxGasLimit uint64
	TxGasPrice *big.Int
	TxTo       ethlike.Address
	TxAmount   *big.Int
	TxData     []byte
}

func (mt *MockTransaction) Hash() ethlike.Hash {
	return mt.TxHash
}

func (mt *MockTransaction) GasPrice() *big.Int {
	return mt.TxGasPrice
}

type MockDeployBackend struct {
	Receipt ethlike.Receipt
}

func (mdb *MockDeployBackend) TransactionReceipt(
	ctx context.Context,
	txHash ethlike.Hash,
) (ethlike.Receipt, error) {
	return mdb.Receipt, nil
}

type MockContractTransactor struct {
	NextNonce uint64
}

func (mct *MockContractTransactor) PendingCodeAt(
	ctx context.Context,
	account ethlike.Address,
) ([]byte, error) {
	panic("implement me")
}

func (mct *MockContractTransactor) PendingNonceAt(
	ctx context.Context,
	account ethlike.Address,
) (uint64, error) {
	return mct.NextNonce, nil
}

func (mct *MockContractTransactor) SuggestGasPrice(
	ctx context.Context,
) (*big.Int, error) {
	panic("implement me")
}

func (mct *MockContractTransactor) EstimateGas(
	ctx context.Context,
	call ethlike.CallMsg,
) (gas uint64, err error) {
	panic("implement me")
}

func (mct *MockContractTransactor) SendTransaction(
	ctx context.Context,
	tx ethlike.Transaction,
) error {
	panic("implement me")
}

type MockCallMsg struct{}

type MockFilterQuery struct{}

type MockClient struct {
	requestDuration time.Duration

	events []string
	mutex  sync.Mutex
}

func NewMockClient(requestDuration time.Duration) *MockClient {
	return &MockClient{
		requestDuration: requestDuration,
		events:          make([]string, 0),
	}
}

func (mc *MockClient) EventsSnapshot() []string {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	snapshot := make([]string, len(mc.events))
	copy(snapshot[:], mc.events[:])
	return snapshot
}

func (mc *MockClient) mockRequest() {
	mc.mutex.Lock()
	mc.events = append(mc.events, "start")
	mc.mutex.Unlock()

	time.Sleep(mc.requestDuration)

	mc.mutex.Lock()
	mc.events = append(mc.events, "end")
	mc.mutex.Unlock()
}

func (mc *MockClient) CodeAt(
	ctx context.Context,
	contract ethlike.Address,
	blockNumber *big.Int,
) ([]byte, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) CallContract(
	ctx context.Context,
	call ethlike.CallMsg,
	blockNumber *big.Int,
) ([]byte, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) PendingCodeAt(
	ctx context.Context,
	account ethlike.Address,
) ([]byte, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) PendingNonceAt(
	ctx context.Context,
	account ethlike.Address,
) (uint64, error) {
	mc.mockRequest()
	return 0, nil
}

func (mc *MockClient) SuggestGasPrice(
	ctx context.Context,
) (*big.Int, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) EstimateGas(
	ctx context.Context,
	call ethlike.CallMsg,
) (uint64, error) {
	mc.mockRequest()
	return 0, nil
}

func (mc *MockClient) SendTransaction(
	ctx context.Context,
	tx ethlike.Transaction,
) error {
	mc.mockRequest()
	return nil
}

func (mc *MockClient) FilterLogs(
	ctx context.Context,
	query ethlike.FilterQuery,
) ([]ethlike.Log, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) SubscribeFilterLogs(
	ctx context.Context,
	query ethlike.FilterQuery,
	ch chan<- ethlike.Log,
) (ethlike.Subscription, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) BlockByHash(
	ctx context.Context,
	hash ethlike.Hash,
) (ethlike.Block, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) BlockByNumber(
	ctx context.Context,
	number *big.Int,
) (ethlike.Block, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) HeaderByHash(
	ctx context.Context,
	hash ethlike.Hash,
) (ethlike.Header, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) HeaderByNumber(
	ctx context.Context,
	number *big.Int,
) (ethlike.Header, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) TransactionCount(
	ctx context.Context,
	blockHash ethlike.Hash,
) (uint, error) {
	mc.mockRequest()
	return 0, nil
}

func (mc *MockClient) TransactionInBlock(
	ctx context.Context,
	blockHash ethlike.Hash,
	index uint,
) (ethlike.Transaction, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) SubscribeNewHead(
	ctx context.Context,
	ch chan<- ethlike.Header,
) (ethlike.Subscription, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) TransactionByHash(
	ctx context.Context,
	txHash ethlike.Hash,
) (ethlike.Transaction, bool, error) {
	mc.mockRequest()
	return nil, false, nil
}

func (mc *MockClient) TransactionReceipt(
	ctx context.Context,
	txHash ethlike.Hash,
) (ethlike.Receipt, error) {
	mc.mockRequest()
	return nil, nil
}

func (mc *MockClient) BalanceAt(
	ctx context.Context,
	account ethlike.Address,
	blockNumber *big.Int,
) (*big.Int, error) {
	mc.mockRequest()
	return nil, nil
}
