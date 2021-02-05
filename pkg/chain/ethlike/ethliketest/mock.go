package ethliketest

import (
	"context"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
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

func (mdb *MockDeployBackend) CodeAt(
	ctx context.Context,
	account ethlike.Address,
	blockNumber *big.Int,
) ([]byte, error) {
	panic("not implemented")
}
