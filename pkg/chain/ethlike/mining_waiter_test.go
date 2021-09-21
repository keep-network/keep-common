package ethlike

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"
)

const checkInterval = 100 * time.Millisecond

var maxGasPrice = big.NewInt(45000000000) // 45 Gwei

func TestForceMining_Legacy_FirstMined(t *testing.T) {
	originalTransaction := createLegacyTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := newMockChain()

	var resubmissionGasPrices []*big.Int

	resubmitFn := func(gasPrice *big.Int) (*Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		return createLegacyTransaction(gasPrice), nil
	}

	// receipt is already there
	chain.receipt = &Receipt{}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasPrice,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionGasPrices)
	if resubmissionCount != 0 {
		t.Fatalf("expected no resubmissions; has: [%v]", resubmissionCount)
	}
}

func TestForceMining_Legacy_SecondMined(t *testing.T) {
	originalTransaction := createLegacyTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := newMockChain()

	var resubmissionGasPrices []*big.Int

	resubmitFn := func(gasPrice *big.Int) (*Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		// first resubmission succeeded
		chain.receipt = &Receipt{}
		return createLegacyTransaction(gasPrice), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasPrice,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionGasPrices)
	if resubmissionCount != 1 {
		t.Fatalf("expected one resubmission; has: [%v]", resubmissionCount)
	}
}

func TestForceMining_Legacy_MultipleAttempts(t *testing.T) {
	originalTransaction := createLegacyTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := newMockChain()

	var resubmissionGasPrices []*big.Int

	expectedAttempts := 3
	expectedResubmissionGasPrices := []*big.Int{
		big.NewInt(24000000000), // + 20%
		big.NewInt(28800000000), // + 20%
		big.NewInt(34560000000), // + 20%
	}

	attemptsSoFar := 1
	resubmitFn := func(gasPrice *big.Int) (*Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		if attemptsSoFar == expectedAttempts {
			chain.receipt = &Receipt{}
		} else {
			attemptsSoFar++
		}
		return createLegacyTransaction(gasPrice), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasPrice,
	)
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

func TestForceMining_Legacy_MaxAllowedPriceReached(t *testing.T) {
	originalTransaction := createLegacyTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := newMockChain()

	var resubmissionGasPrices []*big.Int

	expectedAttempts := 5
	expectedResubmissionGasPrices := []*big.Int{
		big.NewInt(24000000000), // + 20%
		big.NewInt(28800000000), // + 20%
		big.NewInt(34560000000), // + 20%
		big.NewInt(41472000000), // + 20%
		big.NewInt(45000000000), // max allowed
	}

	resubmitFn := func(gasPrice *big.Int) (*Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		// not setting mockBackend.receipt, mining takes a very long time
		return createLegacyTransaction(gasPrice), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasPrice,
	)
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

func TestForceMining_Legacy_OriginalPriceHigherThanMaxAllowed(t *testing.T) {
	// original transaction was priced at 46 Gwei, the maximum allowed gas price
	// is 45 Gwei
	originalTransaction := createLegacyTransaction(big.NewInt(46000000000))

	chain := newMockChain()

	var resubmissionGasPrices []*big.Int

	resubmitFn := func(gasPrice *big.Int) (*Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		// not setting mockBackend.receipt, mining takes a very long time
		return createLegacyTransaction(gasPrice), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasPrice,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionGasPrices)
	if resubmissionCount != 0 {
		t.Fatalf("expected no resubmissions; has: [%v]", resubmissionCount)
	}
}

func createLegacyTransaction(gasPrice *big.Int) *Transaction {
	hashSlice, err := hex.DecodeString(
		"121D387731bBbC988B312206c74F77D004D6B84b",
	)
	if err != nil {
		return nil
	}

	var hash [32]byte
	copy(hash[:], hashSlice)

	return &Transaction{
		Hash:     hash,
		GasPrice: gasPrice,
	}
}

type mockChain struct {
	*mockChainReader
	*mockTransactionReader
	*mockContractTransactor
}

func newMockChain() *mockChain {
	return &mockChain{
		mockChainReader:        &mockChainReader{},
		mockTransactionReader:  &mockTransactionReader{},
		mockContractTransactor: &mockContractTransactor{},
	}
}

type mockChainReader struct{}

func (mcr *mockChainReader) BlockByNumber(
	ctx context.Context,
	number *big.Int,
) (*Block, error) {
	panic("implement me")
}

func (mcr *mockChainReader) SubscribeNewHead(
	ctx context.Context,
	ch chan<- *Header,
) (Subscription, error) {
	panic("implement me")
}

type mockTransactionReader struct {
	receipt *Receipt
}

func (mtr *mockTransactionReader) TransactionReceipt(
	ctx context.Context,
	txHash Hash,
) (*Receipt, error) {
	return mtr.receipt, nil
}
