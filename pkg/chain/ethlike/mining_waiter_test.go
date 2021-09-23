package ethlike

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"
)

const checkInterval = 100 * time.Millisecond

var maxGasFeeCap = big.NewInt(45000000000) // 45 Gwei

func TestForceMining_Legacy_NoResubmission(t *testing.T) {
	originalTransaction := createLegacyTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := newMockChain()

	var resubmissionParams []*ResubmitParams

	resubmitFn := func(params *ResubmitParams) (*Transaction, error) {
		resubmissionParams = append(resubmissionParams, params)
		return createLegacyTransaction(params.GasPrice), nil
	}

	// receipt is already there
	chain.receipt = &Receipt{}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasFeeCap,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionParams)
	if resubmissionCount != 0 {
		t.Fatalf("expected no resubmissions; has: [%v]", resubmissionCount)
	}
}

func TestForceMining_Legacy_OneResubmission(t *testing.T) {
	originalTransaction := createLegacyTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := newMockChain()

	var resubmissionParams []*ResubmitParams

	resubmitFn := func(params *ResubmitParams) (*Transaction, error) {
		resubmissionParams = append(resubmissionParams, params)
		// first resubmission succeeded
		chain.receipt = &Receipt{}
		return createLegacyTransaction(params.GasPrice), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasFeeCap,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionParams)
	if resubmissionCount != 1 {
		t.Fatalf("expected one resubmission; has: [%v]", resubmissionCount)
	}

	resubmission := resubmissionParams[0]

	if resubmission.GasFeeCap != nil || resubmission.GasTipCap != nil {
		t.Fatalf("gas fee and tip cap should be nil")
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

func TestForceMining_Legacy_MultipleAttempts(t *testing.T) {
	originalTransaction := createLegacyTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := newMockChain()

	var resubmissionParams []*ResubmitParams

	expectedAttempts := 3
	expectedResubmissionGasPrices := []*big.Int{
		big.NewInt(24000000000), // + 20%
		big.NewInt(28800000000), // + 20%
		big.NewInt(34560000000), // + 20%
	}

	attemptsSoFar := 1
	resubmitFn := func(params *ResubmitParams) (*Transaction, error) {
		resubmissionParams = append(resubmissionParams, params)
		if attemptsSoFar == expectedAttempts {
			chain.receipt = &Receipt{}
		} else {
			attemptsSoFar++
		}
		return createLegacyTransaction(params.GasPrice), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasFeeCap,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionParams)
	if resubmissionCount != expectedAttempts {
		t.Fatalf(
			"expected [%v] resubmission; has: [%v]",
			expectedAttempts,
			resubmissionCount,
		)
	}

	for index, resubmission := range resubmissionParams {
		if resubmission.GasFeeCap != nil || resubmission.GasTipCap != nil {
			t.Fatalf("resubmission [%v] gas fee and tip cap should be nil", index)
		}

		price := resubmission.GasPrice
		if price.Cmp(expectedResubmissionGasPrices[index]) != 0 {
			t.Fatalf(
				"unexpected resubmission [%v] gas price\n"+
					"expected: [%v]\n"+
					"actual:   [%v]",
				resubmission,
				expectedResubmissionGasPrices[index],
				price,
			)
		}
	}
}

func TestForceMining_Legacy_MaxAllowedPriceReached(t *testing.T) {
	originalTransaction := createLegacyTransaction(big.NewInt(20000000000)) // 20 Gwei

	chain := newMockChain()

	var resubmissionParams []*ResubmitParams

	expectedAttempts := 5
	expectedResubmissionGasPrices := []*big.Int{
		big.NewInt(24000000000), // + 20%
		big.NewInt(28800000000), // + 20%
		big.NewInt(34560000000), // + 20%
		big.NewInt(41472000000), // + 20%
		big.NewInt(45000000000), // max allowed
	}

	resubmitFn := func(params *ResubmitParams) (*Transaction, error) {
		resubmissionParams = append(resubmissionParams, params)
		// not setting mockBackend.receipt, mining takes a very long time
		return createLegacyTransaction(params.GasPrice), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasFeeCap,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionParams)
	if resubmissionCount != expectedAttempts {
		t.Fatalf(
			"expected [%v] resubmission; has: [%v]",
			expectedAttempts,
			resubmissionCount,
		)
	}

	for index, resubmission := range resubmissionParams {
		if resubmission.GasFeeCap != nil || resubmission.GasTipCap != nil {
			t.Fatalf("resubmission [%v] gas fee and tip cap should be nil", index)
		}

		price := resubmission.GasPrice
		if price.Cmp(expectedResubmissionGasPrices[index]) != 0 {
			t.Fatalf(
				"unexpected resubmission [%v] gas price\n"+
					"expected: [%v]\n"+
					"actual:   [%v]",
				resubmission,
				expectedResubmissionGasPrices[index],
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

	var resubmissionParams []*ResubmitParams

	resubmitFn := func(params *ResubmitParams) (*Transaction, error) {
		resubmissionParams = append(resubmissionParams, params)
		// not setting mockBackend.receipt, mining takes a very long time
		return createLegacyTransaction(params.GasPrice), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasFeeCap,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionParams)
	if resubmissionCount != 0 {
		t.Fatalf("expected no resubmissions; has: [%v]", resubmissionCount)
	}
}

func TestForceMining_DynamicFee_NoResubmission(t *testing.T) {
	originalBaseFee := big.NewInt(10000000000)   // 10 Gwei
	originalGasTipCap := big.NewInt(4000000000)  // 4 Gwei
	originalGasFeeCap := big.NewInt(24000000000) // 24 Gwei (2 * baseFee + gasTipCap)

	originalTransaction := createDynamicFeeTransaction(
		originalGasFeeCap,
		originalGasTipCap,
	)

	chain := newMockChain()

	// Base fee remains unchanged.
	chain.blocks = []*Block{
		{&Header{big.NewInt(1), originalBaseFee}},
	}

	var resubmissionParams []*ResubmitParams

	resubmitFn := func(params *ResubmitParams) (*Transaction, error) {
		resubmissionParams = append(resubmissionParams, params)
		return createDynamicFeeTransaction(
			params.GasFeeCap,
			params.GasTipCap,
		), nil
	}

	// Receipt is already there.
	chain.receipt = &Receipt{}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasFeeCap,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionParams)
	if resubmissionCount != 0 {
		t.Fatalf("expected no resubmissions; has: [%v]", resubmissionCount)
	}
}

func TestForceMining_DynamicFee_OneResubmission(t *testing.T) {
	originalBaseFee := big.NewInt(10000000000)   // 10 Gwei
	originalGasTipCap := big.NewInt(4000000000)  // 4 Gwei
	originalGasFeeCap := big.NewInt(24000000000) // 24 Gwei (2 * baseFee + gasTipCap)

	var tests = map[string]struct {
		nextBaseFee       *big.Int
		expectedGasFeeCap *big.Int
		expectedGasTipCap *big.Int
	}{
		"base fee decreased": {
			// Base fee decreased to 5 Gwei.
			nextBaseFee: big.NewInt(5000000000),
			// Gas fee cap should be computed as: 2 * 5 Gwei + 4.8 Gwei = 14.8 Gwei.
			// However, this value doesn't fulfill the required fee bump threshold.
			// In result, the new gas fee value should be bumped up by 10% to
			// satisfy the condition: 24 Gwei * 1.1 = 26.4 Gwei
			expectedGasFeeCap: big.NewInt(26400000000),
			// Gas tip cap should be bumped up by 20%: 4 Gwei * 1.2 = 4.8 Gwei
			expectedGasTipCap: big.NewInt(4800000000),
		},
		"base fee unchanged": {
			// Base fee remains 10 Gwei.
			nextBaseFee: originalBaseFee,
			// Gas fee cap should be computed as: 2 * 10 Gwei + 4.8 Gwei = 24.8 Gwei.
			// However, this value doesn't fulfill the required fee bump threshold.
			// In result, the new gas fee value should be bumped up by 10% to
			// satisfy the condition: 24 Gwei * 1.1 = 26.4 Gwei
			expectedGasFeeCap: big.NewInt(26400000000),
			// Gas tip cap should be bumped up by 20%: 4 Gwei * 1.2 = 4.8 Gwei
			expectedGasTipCap: big.NewInt(4800000000),
		},
		"base fee increased": {
			// Base fee increased to 20 Gwei.
			nextBaseFee: big.NewInt(20000000000),
			// Gas fee cap should be computed as: 2 * 20 Gwei + 4.8 Gwei = 44.8 Gwei.
			expectedGasFeeCap: big.NewInt(44800000000),
			// Gas tip cap should be bumped up by 20%: 4 Gwei * 1.2 = 4.8 Gwei
			expectedGasTipCap: big.NewInt(4800000000),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			originalTransaction := createDynamicFeeTransaction(
				originalGasFeeCap,
				originalGasTipCap,
			)

			chain := newMockChain()

			chain.blocks = []*Block{
				{&Header{big.NewInt(1), test.nextBaseFee}},
			}

			var resubmissionParams []*ResubmitParams

			resubmitFn := func(params *ResubmitParams) (*Transaction, error) {
				resubmissionParams = append(resubmissionParams, params)
				// First resubmission succeeded.
				chain.receipt = &Receipt{}
				return createDynamicFeeTransaction(
					params.GasFeeCap,
					params.GasTipCap,
				), nil
			}

			waiter := NewMiningWaiter(
				chain,
				checkInterval,
				maxGasFeeCap,
			)
			waiter.ForceMining(
				originalTransaction,
				resubmitFn,
			)

			resubmissionCount := len(resubmissionParams)
			if resubmissionCount != 1 {
				t.Fatalf(
					"expected one resubmission; has: [%v]",
					resubmissionCount,
				)
			}

			resubmission := resubmissionParams[0]

			if resubmission.GasPrice != nil {
				t.Fatalf("gas price should be nil")
			}

			if resubmission.GasFeeCap.Cmp(test.expectedGasFeeCap) != 0 {
				t.Fatalf(
					"unexpected gas fee cap value\n"+
						"expected: [%v]\n"+
						"actual:   [%v]",
					test.expectedGasFeeCap,
					resubmission.GasFeeCap,
				)
			}

			if resubmission.GasTipCap.Cmp(test.expectedGasTipCap) != 0 {
				t.Fatalf(
					"unexpected gas tip cap value\n"+
						"expected: [%v]\n"+
						"actual:   [%v]",
					test.expectedGasTipCap,
					resubmission.GasTipCap,
				)
			}
		})
	}
}

func TestForceMining_DynamicFee_MultipleAttemps(t *testing.T) {
	originalBaseFee := big.NewInt(10000000000)   // 10 Gwei
	originalGasTipCap := big.NewInt(4000000000)  // 4 Gwei
	originalGasFeeCap := big.NewInt(24000000000) // 24 Gwei (2 * baseFee + gasTipCap)

	originalTransaction := createDynamicFeeTransaction(
		originalGasFeeCap,
		originalGasTipCap,
	)

	chain := newMockChain()

	// Base fee remains unchanged.
	chain.blocks = []*Block{
		{&Header{big.NewInt(1), originalBaseFee}},
	}

	var resubmissionParams []*ResubmitParams

	expectedAttempts := 3
	expectedResubmissionParams := []*ResubmitParams{
		// gasFeeCap +10%, gasTipCap +20%
		{nil, big.NewInt(26400000000), big.NewInt(4800000000)},
		// gasFeeCap +10%, gasTipCap +20%
		{nil, big.NewInt(29040000000), big.NewInt(5760000000)},
		// gasFeeCap +10%, gasTipCap +20%
		{nil, big.NewInt(31944000000), big.NewInt(6912000000)},
	}

	attemptsSoFar := 1
	resubmitFn := func(params *ResubmitParams) (*Transaction, error) {
		resubmissionParams = append(resubmissionParams, params)
		if attemptsSoFar == expectedAttempts {
			chain.receipt = &Receipt{}
		} else {
			attemptsSoFar++
		}
		return createDynamicFeeTransaction(
			params.GasFeeCap,
			params.GasTipCap,
		), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasFeeCap,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionParams)
	if resubmissionCount != expectedAttempts {
		t.Fatalf(
			"expected [%v] resubmission; has: [%v]",
			expectedAttempts,
			resubmissionCount,
		)
	}

	for index, resubmission := range resubmissionParams {
		if resubmission.GasPrice != nil {
			t.Fatalf("resubmission [%v] gas price should be nil", index)
		}

		expectedGasFeeCap := expectedResubmissionParams[index].GasFeeCap
		if resubmission.GasFeeCap.Cmp(expectedGasFeeCap) != 0 {
			t.Fatalf(
				"unexpected resubmission [%v] gas fee cap value\n"+
					"expected: [%v]\n"+
					"actual:   [%v]",
				index,
				expectedGasFeeCap,
				resubmission.GasFeeCap,
			)
		}

		expectedGasTipCap := expectedResubmissionParams[index].GasTipCap
		if resubmission.GasTipCap.Cmp(expectedGasTipCap) != 0 {
			t.Fatalf(
				"unexpected resubmission [%v]  gas tip cap value\n"+
					"expected: [%v]\n"+
					"actual:   [%v]",
				index,
				expectedGasTipCap,
				resubmission.GasTipCap,
			)
		}
	}
}

func TestForceMining_DynamicFee_MaxAllowedPriceReached(t *testing.T) {
	originalBaseFee := big.NewInt(10000000000)   // 10 Gwei
	originalGasTipCap := big.NewInt(4000000000)  // 4 Gwei
	originalGasFeeCap := big.NewInt(24000000000) // 24 Gwei (2 * baseFee + gasTipCap)

	originalTransaction := createDynamicFeeTransaction(
		originalGasFeeCap,
		originalGasTipCap,
	)

	chain := newMockChain()

	chain.blocks = []*Block{
		{
			&Header{
				big.NewInt(1),
				// Massive increase of base fee to 30 Gwei. This is needed
				// to exceed the maximum gas fee cap value.
				new(big.Int).Mul(originalBaseFee, big.NewInt(3)),
			},
		},
	}

	var resubmissionParams []*ResubmitParams

	resubmitFn := func(params *ResubmitParams) (*Transaction, error) {
		resubmissionParams = append(resubmissionParams, params)
		// Not setting mockBackend.receipt, mining takes a very long time.
		return createDynamicFeeTransaction(
			params.GasFeeCap,
			params.GasTipCap,
		), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasFeeCap,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionParams)
	if resubmissionCount != 1 {
		t.Fatalf(
			"expected one resubmission; has: [%v]",
			resubmissionCount,
		)
	}

	resubmission := resubmissionParams[0]

	if resubmission.GasPrice != nil {
		t.Fatalf("gas price should be nil")
	}

	// The new gas fee value should be computed as usual:
	// 2 * 30 Gwei + 4.8 gwei = 64.8 Gwei. However, max gas fee cap value
	// is 45 Gwei so this value should be used instead.
	expectedGasFeeCap := big.NewInt(45000000000)
	if resubmission.GasFeeCap.Cmp(expectedGasFeeCap) != 0 {
		t.Fatalf(
			"unexpected gas fee cap value\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedGasFeeCap,
			resubmission.GasFeeCap,
		)
	}

	// Gas tip cap should be bumped up by 20%: 4 Gwei * 1.2 = 4.8 Gwei
	expectedGasTipCap := big.NewInt(4800000000)
	if resubmission.GasTipCap.Cmp(expectedGasTipCap) != 0 {
		t.Fatalf(
			"unexpected gas tip cap value\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedGasTipCap,
			resubmission.GasTipCap,
		)
	}
}

func TestForceMining_DynamicFee_MaxAllowedPriceReachedButBelowThreshold(t *testing.T) {
	originalBaseFee := big.NewInt(10000000000)   // 10 Gwei
	originalGasTipCap := big.NewInt(4000000000)  // 4 Gwei
	originalGasFeeCap := big.NewInt(24000000000) // 24 Gwei (2 * baseFee + gasTipCap)

	originalTransaction := createDynamicFeeTransaction(
		originalGasFeeCap,
		originalGasTipCap,
	)

	chain := newMockChain()

	// Base fee remains unchanged.
	chain.blocks = []*Block{
		{&Header{big.NewInt(1), originalBaseFee}},
	}

	var resubmissionParams []*ResubmitParams

	expectedAttempts := 6
	expectedResubmissionParams := []*ResubmitParams{
		// gasFeeCap +10%, gasTipCap +20%
		{nil, big.NewInt(26400000000), big.NewInt(4800000000)},
		// gasFeeCap +10%, gasTipCap +20%
		{nil, big.NewInt(29040000000), big.NewInt(5760000000)},
		// gasFeeCap +10%, gasTipCap +20%
		{nil, big.NewInt(31944000000), big.NewInt(6912000000)},
		// gasFeeCap +10%, gasTipCap +20%
		{nil, big.NewInt(35138400000), big.NewInt(8294400000)},
		// gasFeeCap +10%, gasTipCap +20%
		{nil, big.NewInt(38652240000), big.NewInt(9953280000)},
		// gasFeeCap +10%, gasTipCap +20%
		{nil, big.NewInt(42517464000), big.NewInt(11943936000)},
		// The last attempt would bump the gas fee cap up to 46769210400.
		// However, this value exceeds the max gas fee cap which is 45000000000.
		// On the other hand, the max gas fee cap is under the required fee bump
		// threshold for this iteration so resubmission should not be performed
		// at all.
	}

	resubmitFn := func(params *ResubmitParams) (*Transaction, error) {
		resubmissionParams = append(resubmissionParams, params)
		// Not setting mockBackend.receipt, mining takes a very long time.
		return createDynamicFeeTransaction(
			params.GasFeeCap,
			params.GasTipCap,
		), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasFeeCap,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionParams)
	if resubmissionCount != expectedAttempts {
		t.Fatalf(
			"expected [%v] resubmission; has: [%v]",
			expectedAttempts,
			resubmissionCount,
		)
	}

	for index, resubmission := range resubmissionParams {
		if resubmission.GasPrice != nil {
			t.Fatalf("resubmission [%v] gas price should be nil", index)
		}

		expectedGasFeeCap := expectedResubmissionParams[index].GasFeeCap
		if resubmission.GasFeeCap.Cmp(expectedGasFeeCap) != 0 {
			t.Fatalf(
				"unexpected resubmission [%v] gas fee cap value\n"+
					"expected: [%v]\n"+
					"actual:   [%v]",
				index,
				expectedGasFeeCap,
				resubmission.GasFeeCap,
			)
		}

		expectedGasTipCap := expectedResubmissionParams[index].GasTipCap
		if resubmission.GasTipCap.Cmp(expectedGasTipCap) != 0 {
			t.Fatalf(
				"unexpected resubmission [%v]  gas tip cap value\n"+
					"expected: [%v]\n"+
					"actual:   [%v]",
				index,
				expectedGasTipCap,
				resubmission.GasTipCap,
			)
		}
	}
}

func TestForceMining_DynamicFee_OriginalPriceHigherThanMaxAllowed(t *testing.T) {
	// Original transaction has gas fee cap set at 46 Gwei, the maximum allowed
	// gas fee cap is 45 Gwei.
	originalBaseFee := big.NewInt(10000000000)   // 10 Gwei
	originalGasTipCap := big.NewInt(4000000000)  // 4 Gwei
	originalGasFeeCap := big.NewInt(46000000000) // 46 Gwei

	originalTransaction := createDynamicFeeTransaction(
		originalGasFeeCap,
		originalGasTipCap,
	)

	chain := newMockChain()

	// Base fee remains unchanged.
	chain.blocks = []*Block{
		{&Header{big.NewInt(1), originalBaseFee}},
	}

	var resubmissionParams []*ResubmitParams

	resubmitFn := func(params *ResubmitParams) (*Transaction, error) {
		resubmissionParams = append(resubmissionParams, params)
		// Not setting mockBackend.receipt, mining takes a very long time.
		return createDynamicFeeTransaction(
			params.GasFeeCap,
			params.GasTipCap,
		), nil
	}

	waiter := NewMiningWaiter(
		chain,
		checkInterval,
		maxGasFeeCap,
	)
	waiter.ForceMining(
		originalTransaction,
		resubmitFn,
	)

	resubmissionCount := len(resubmissionParams)
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

func createDynamicFeeTransaction(gasFeeCap, gasTipCap *big.Int) *Transaction {
	hashSlice, err := hex.DecodeString(
		"121D387731bBbC988B312206c74F77D004D6B84b",
	)
	if err != nil {
		return nil
	}

	var hash [32]byte
	copy(hash[:], hashSlice)

	return &Transaction{
		Hash:      hash,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Type:      DynamicFeeTxType,
	}
}

type mockChain struct {
	*mockChainReader
	*mockTransactionReader
	*mockContractTransactor
}

func newMockChain() *mockChain {
	return &mockChain{
		mockChainReader: &mockChainReader{
			blocks: make([]*Block, 0),
		},
		mockTransactionReader:  &mockTransactionReader{},
		mockContractTransactor: &mockContractTransactor{},
	}
}

type mockChainReader struct {
	blocks []*Block
}

func (mcr *mockChainReader) BlockByNumber(
	ctx context.Context,
	number *big.Int,
) (*Block, error) {
	if number == nil {
		return mcr.blocks[len(mcr.blocks)-1], nil
	}

	return mcr.blocks[number.Int64()], nil
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
