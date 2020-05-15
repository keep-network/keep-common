package ethutil

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const checkInterval = 100 * time.Millisecond

var maxGasPrice = big.NewInt(45000000000) // 45 Gwei

func TestForceMining_FirstMined(t *testing.T) {
	originalTransaction := createTransaction(big.NewInt(20000000000)) // 20 Gwei

	mockBackend := &mockDeployBackend{}

	var resubmissionGasPrices []*big.Int

	resubmitFn := func(gasPrice *big.Int) (*types.Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		return createTransaction(gasPrice), nil
	}

	// receipt is already there
	mockBackend.receipt = &types.Receipt{}

	waiter := NewMiningWaiter(mockBackend, checkInterval, maxGasPrice)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionGasPrices)
	if resubmissionCount != 0 {
		t.Fatalf("expected no resubmissions; has: [%v]", resubmissionCount)
	}
}

func TestForceMining_SecondMined(t *testing.T) {
	originalTransaction := createTransaction(big.NewInt(20000000000)) // 20 Gwei

	mockBackend := &mockDeployBackend{}

	var resubmissionGasPrices []*big.Int

	resubmitFn := func(gasPrice *big.Int) (*types.Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		// first resubmission succeeded
		mockBackend.receipt = &types.Receipt{}
		return createTransaction(gasPrice), nil
	}

	waiter := NewMiningWaiter(mockBackend, checkInterval, maxGasPrice)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionGasPrices)
	if resubmissionCount != 1 {
		t.Fatalf("expected one resubmission; has: [%v]", resubmissionCount)
	}
}

func TestForceMining_MultipleAttempts(t *testing.T) {
	originalTransaction := createTransaction(big.NewInt(20000000000)) // 20 Gwei

	mockBackend := &mockDeployBackend{}

	var resubmissionGasPrices []*big.Int

	expectedAttempts := 3
	expectedResubmissionGasPrices := []*big.Int{
		big.NewInt(24000000000), // + 20%
		big.NewInt(28800000000), // + 20%
		big.NewInt(34560000000), // + 20%
	}

	attemptsSoFar := 1
	resubmitFn := func(gasPrice *big.Int) (*types.Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		if attemptsSoFar == expectedAttempts {
			mockBackend.receipt = &types.Receipt{}
		} else {
			attemptsSoFar++
		}
		return createTransaction(gasPrice), nil
	}

	waiter := NewMiningWaiter(mockBackend, checkInterval, maxGasPrice)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionGasPrices)
	if resubmissionCount != expectedAttempts {
		t.Fatalf(
			"expected [%v] resubmission; has: [%v]",
			expectedAttempts,
			resubmissionCount,
		)
	}

	for resubmission, price := range resubmissionGasPrices {
		if price.Cmp(expectedResubmissionGasPrices[resubmission]) != 0 {
			t.Fatalf(
				"unexpected [%v] resubmission gas price\nexpected: [%v]\nactual:   [%v]",
				resubmission,
				expectedResubmissionGasPrices[resubmission],
				price,
			)
		}
	}
}

func TestForceMining_MaxAllowedPriceReached(t *testing.T) {
	originalTransaction := createTransaction(big.NewInt(20000000000)) // 20 Gwei

	mockBackend := &mockDeployBackend{}

	var resubmissionGasPrices []*big.Int

	expectedAttempts := 4
	expectedResubmissionGasPrices := []*big.Int{
		big.NewInt(24000000000), // + 20%
		big.NewInt(28800000000), // + 20%
		big.NewInt(34560000000), // + 20%
		big.NewInt(41472000000), // + 20%
		// the next one would be 49766400000 but since maxGasPrice = 45 Gwei
		// resubmissions should stop here
	}

	resubmitFn := func(gasPrice *big.Int) (*types.Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		// not setting mockBackend.receipt, mining takes a very long time
		return createTransaction(gasPrice), nil
	}

	waiter := NewMiningWaiter(mockBackend, checkInterval, maxGasPrice)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionGasPrices)
	if resubmissionCount != expectedAttempts {
		t.Fatalf(
			"expected [%v] resubmission; has: [%v]",
			expectedAttempts,
			resubmissionCount,
		)
	}

	for resubmission, price := range resubmissionGasPrices {
		if price.Cmp(expectedResubmissionGasPrices[resubmission]) != 0 {
			t.Fatalf(
				"unexpected [%v] resubmission gas price\nexpected: [%v]\nactual:   [%v]",
				resubmission,
				expectedResubmissionGasPrices[resubmission],
				price,
			)
		}
	}
}

func createTransaction(gasPrice *big.Int) *types.Transaction {
	return types.NewTransaction(
		10, // nonce
		common.HexToAddress("0x131D387731bBbC988B312206c74F77D004D6B84b"), // to
		big.NewInt(0), // amount
		200000,        // gas limit
		gasPrice,      // gas price
		[]byte{},      // data
	)
}

type mockDeployBackend struct {
	receipt *types.Receipt
}

func (mdb *mockDeployBackend) TransactionReceipt(
	ctx context.Context,
	txHash common.Hash,
) (*types.Receipt, error) {
	return mdb.receipt, nil
}

func (mdb *mockDeployBackend) CodeAt(
	ctx context.Context,
	account common.Address,
	blockNumber *big.Int,
) ([]byte, error) {
	panic("not implemented")
}
