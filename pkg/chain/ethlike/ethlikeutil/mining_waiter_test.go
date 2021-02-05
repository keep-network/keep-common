package ethlikeutil

import (
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"github.com/keep-network/keep-common/pkg/chain/ethlike/ethliketest"
	"math/big"
	"testing"
	"time"
)

const checkInterval = 100 * time.Millisecond

var maxGasPrice = big.NewInt(45000000000) // 45 Gwei

func TestForceMining_FirstMined(t *testing.T) {
	originalTransaction := createTransaction(big.NewInt(20000000000)) // 20 Gwei

	mockBackend := &ethliketest.MockDeployBackend{}

	var resubmissionGasPrices []*big.Int

	resubmitFn := func(gasPrice *big.Int) (ethlike.Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		return createTransaction(gasPrice), nil
	}

	// receipt is already there
	mockBackend.Receipt = &ethliketest.MockReceipt{}

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

	mockBackend := &ethliketest.MockDeployBackend{}

	var resubmissionGasPrices []*big.Int

	resubmitFn := func(gasPrice *big.Int) (ethlike.Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		// first resubmission succeeded
		mockBackend.Receipt = &ethliketest.MockReceipt{}
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

	mockBackend := &ethliketest.MockDeployBackend{}

	var resubmissionGasPrices []*big.Int

	expectedAttempts := 3
	expectedResubmissionGasPrices := []*big.Int{
		big.NewInt(24000000000), // + 20%
		big.NewInt(28800000000), // + 20%
		big.NewInt(34560000000), // + 20%
	}

	attemptsSoFar := 1
	resubmitFn := func(gasPrice *big.Int) (ethlike.Transaction, error) {
		resubmissionGasPrices = append(resubmissionGasPrices, gasPrice)
		if attemptsSoFar == expectedAttempts {
			mockBackend.Receipt = &ethliketest.MockReceipt{}
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

	mockBackend := &ethliketest.MockDeployBackend{}

	var resubmissionGasPrices []*big.Int

	expectedAttempts := 5
	expectedResubmissionGasPrices := []*big.Int{
		big.NewInt(24000000000), // + 20%
		big.NewInt(28800000000), // + 20%
		big.NewInt(34560000000), // + 20%
		big.NewInt(41472000000), // + 20%
		big.NewInt(45000000000), // max allowed
	}

	resubmitFn := func(gasPrice *big.Int) (ethlike.Transaction, error) {
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

func TestForceMining_OriginalPriceHigherThanMaxAllowed(t *testing.T) {
	// original transaction was priced at 46 Gwei, the maximum allowed gas price
	// is 45 Gwei
	originalTransaction := createTransaction(big.NewInt(46000000000))

	mockBackend := &ethliketest.MockDeployBackend{}

	var resubmissionGasPrices []*big.Int

	resubmitFn := func(gasPrice *big.Int) (ethlike.Transaction, error) {
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
	if resubmissionCount != 0 {
		t.Fatalf("expected no resubmissions; has: [%v]", resubmissionCount)
	}
}

func createTransaction(gasPrice *big.Int) ethlike.Transaction {
	hash := "0x121D387731bBbC988B312206c74F77D004D6B84b"
	to := "0x131D387731bBbC988B312206c74F77D004D6B84b"

	return &ethliketest.MockTransaction{
		TxHash:     &ethliketest.MockHash{hash},
		TxNonce:    10,
		TxGasLimit: 200000,
		TxGasPrice: gasPrice,
		TxTo:       &ethliketest.MockAddress{to},
		TxAmount:   big.NewInt(0),
		TxData:     []byte{},
	}
}
