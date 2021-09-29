package celoutil

import (
	"context"
	"github.com/celo-org/celo-blockchain/accounts/abi/bind"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/keep-network/keep-common/pkg/chain/celo"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
	"testing"
)

var config = celo.Config{
	Config: ethlike.Config{
		MiningCheckInterval: 1,
	},
	MaxGasPrice: celo.WrapWei(big.NewInt(45000000000)), // 45 Gwei
}

var originalTransactorOptions = &bind.TransactOpts{
	Nonce: big.NewInt(100),
}

func TestForceMining_NoResubmission(t *testing.T) {
	originalTransaction := createTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := &mockAdaptedCeloClientWithReceipt{}

	var resubmissions []*bind.TransactOpts

	resubmitFn := func(
		newTransactorOptions *bind.TransactOpts,
	) (*types.Transaction, error) {
		resubmissions = append(resubmissions, newTransactorOptions)
		return createTransaction(newTransactorOptions.GasPrice), nil
	}

	// receipt is already there
	chain.receipt = &types.Receipt{}

	waiter := NewMiningWaiter(chain, config)
	waiter.ForceMining(
		originalTransaction,
		originalTransactorOptions,
		resubmitFn,
	)

	resubmissionCount := len(resubmissions)
	if resubmissionCount != 0 {
		t.Fatalf("expected no resubmissions; has: [%v]", resubmissionCount)
	}
}

func TestForceMining_OneResubmission(t *testing.T) {
	originalTransaction := createTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := &mockAdaptedCeloClientWithReceipt{}

	var resubmissions []*bind.TransactOpts

	resubmitFn := func(
		newTransactorOptions *bind.TransactOpts,
	) (*types.Transaction, error) {
		resubmissions = append(resubmissions, newTransactorOptions)
		// first resubmission succeeded
		chain.receipt = &types.Receipt{}
		return createTransaction(newTransactorOptions.GasPrice), nil
	}

	waiter := NewMiningWaiter(chain, config)
	waiter.ForceMining(
		originalTransaction,
		originalTransactorOptions,
		resubmitFn,
	)

	resubmissionCount := len(resubmissions)
	if resubmissionCount != 1 {
		t.Fatalf("expected one resubmission; has: [%v]", resubmissionCount)
	}

	resubmission := resubmissions[0]

	assertNonceUnchanged(t, resubmission)

	if resubmission.GasLimit != originalTransaction.Gas() {
		t.Fatalf("gas limit should be the same as in original transaction")
	}

	expectedGasPrice := big.NewInt(24000000000)
	if resubmission.GasPrice.Cmp(expectedGasPrice) != 0 {
		t.Fatalf(
			"unexpected gas price value\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedGasPrice,
			resubmission.GasPrice,
		)
	}
}

func TestForceMining_MultipleAttempts(t *testing.T) {
	originalTransaction := createTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := &mockAdaptedCeloClientWithReceipt{}

	var resubmissions []*bind.TransactOpts

	expectedAttempts := 3
	expectedResubmissionGasPrices := []*big.Int{
		big.NewInt(24000000000), // + 20%
		big.NewInt(28800000000), // + 20%
		big.NewInt(34560000000), // + 20%
	}

	attemptsSoFar := 1
	resubmitFn := func(
		newTransactorOptions *bind.TransactOpts,
	) (*types.Transaction, error) {
		resubmissions = append(resubmissions, newTransactorOptions)
		if attemptsSoFar == expectedAttempts {
			chain.receipt = &types.Receipt{}
		} else {
			attemptsSoFar++
		}
		return createTransaction(newTransactorOptions.GasPrice), nil
	}

	waiter := NewMiningWaiter(chain, config)
	waiter.ForceMining(
		originalTransaction,
		originalTransactorOptions,
		resubmitFn,
	)

	resubmissionCount := len(resubmissions)
	if resubmissionCount != expectedAttempts {
		t.Fatalf(
			"expected [%v] resubmission; has: [%v]",
			expectedAttempts,
			resubmissionCount,
		)
	}

	for index, resubmission := range resubmissions {
		assertNonceUnchanged(t, resubmission)

		if resubmission.GasLimit != originalTransaction.Gas() {
			t.Fatalf(
				"resubmission [%v] gas limit should be the same as in "+
					"original transaction",
				index,
			)
		}

		price := resubmission.GasPrice
		if price.Cmp(expectedResubmissionGasPrices[index]) != 0 {
			t.Fatalf(
				"unexpected resubmission [%v] gas price\n"+
					"expected: [%v]\n"+
					"actual:   [%v]",
				index,
				expectedResubmissionGasPrices[index],
				price,
			)
		}
	}
}

func TestForceMining_MaxAllowedPriceReached(t *testing.T) {
	originalTransaction := createTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := &mockAdaptedCeloClientWithReceipt{}

	var resubmissions []*bind.TransactOpts

	expectedAttempts := 5
	expectedResubmissionGasPrices := []*big.Int{
		big.NewInt(24000000000), // + 20%
		big.NewInt(28800000000), // + 20%
		big.NewInt(34560000000), // + 20%
		big.NewInt(41472000000), // + 20%
		big.NewInt(45000000000), // max allowed
	}

	resubmitFn := func(
		newTransactorOptions *bind.TransactOpts,
	) (*types.Transaction, error) {
		resubmissions = append(resubmissions, newTransactorOptions)
		// not setting mockBackend.receipt, mining takes a very long time
		return createTransaction(newTransactorOptions.GasPrice), nil
	}

	waiter := NewMiningWaiter(chain, config)
	waiter.ForceMining(
		originalTransaction,
		originalTransactorOptions,
		resubmitFn,
	)

	resubmissionCount := len(resubmissions)
	if resubmissionCount != expectedAttempts {
		t.Fatalf(
			"expected [%v] resubmission; has: [%v]",
			expectedAttempts,
			resubmissionCount,
		)
	}

	for index, resubmission := range resubmissions {
		assertNonceUnchanged(t, resubmission)

		if resubmission.GasLimit != originalTransaction.Gas() {
			t.Fatalf(
				"resubmission [%v] gas limit should be the same as in "+
					"original transaction",
				index,
			)
		}

		price := resubmission.GasPrice
		if price.Cmp(expectedResubmissionGasPrices[index]) != 0 {
			t.Fatalf(
				"unexpected resubmission [%v] gas price\n"+
					"expected: [%v]\n"+
					"actual:   [%v]",
				index,
				expectedResubmissionGasPrices[index],
				price,
			)
		}
	}
}

func TestForceMining_OriginalPriceHigherThanMaxAllowed(t *testing.T) {
	// original transaction was priced at 46 Gwei, the maximum allowed gas price
	// is 45 Gwei
	originalTransaction := createTransaction(big.NewInt(46000000000))

	chain := &mockAdaptedCeloClientWithReceipt{}

	var resubmissions []*bind.TransactOpts

	resubmitFn := func(
		newTransactorOptions *bind.TransactOpts,
	) (*types.Transaction, error) {
		resubmissions = append(resubmissions, newTransactorOptions)
		// not setting mockBackend.receipt, mining takes a very long time
		return createTransaction(newTransactorOptions.GasPrice), nil
	}

	waiter := NewMiningWaiter(chain, config)
	waiter.ForceMining(
		originalTransaction,
		originalTransactorOptions,
		resubmitFn,
	)

	resubmissionCount := len(resubmissions)
	if resubmissionCount != 0 {
		t.Fatalf("expected no resubmissions; has: [%v]", resubmissionCount)
	}
}

func assertNonceUnchanged(
	t *testing.T,
	newTransactionOptions *bind.TransactOpts,
) {
	if newTransactionOptions.Nonce.Cmp(originalTransactorOptions.Nonce) != 0 {
		t.Fatalf("nonce should remain unchanged")
	}
}

func createTransaction(gasPrice *big.Int) *types.Transaction {
	return types.NewTransaction(
		0,
		common.Address{},
		nil,
		25000,
		gasPrice,
		nil,
		nil,
		nil,
		[]byte{},
	)
}

type mockAdaptedCeloClientWithReceipt struct {
	*mockAdaptedCeloClient

	receipt *types.Receipt
}

func (maccwr *mockAdaptedCeloClientWithReceipt) TransactionReceipt(
	ctx context.Context,
	txHash common.Hash,
) (*types.Receipt, error) {
	return maccwr.receipt, nil
}
