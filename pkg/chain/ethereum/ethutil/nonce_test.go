package ethutil

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestResolveAndIncrement(t *testing.T) {
	tests := map[string]struct {
		pendingNonce      uint64
		localNonce        uint64
		expectedNonce     uint64
		expectedNextNonce uint64
	}{
		"pending and local the same": {
			pendingNonce:      10,
			localNonce:        10,
			expectedNonce:     10,
			expectedNextNonce: 11,
		},
		"pending nonce higher": {
			pendingNonce:      121,
			localNonce:        120,
			expectedNonce:     121,
			expectedNextNonce: 122,
		},
		"pending nonce lower": {
			pendingNonce:      110,
			localNonce:        111,
			expectedNonce:     111,
			expectedNextNonce: 112,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			transactor := &mockTransactor{test.pendingNonce}
			manager := &NonceManager{
				transactor: transactor,
				localNonce: test.localNonce,
			}

			nonce, err := manager.CurrentNonce()
			if err != nil {
				t.Fatal(err)
			}

			if nonce != test.expectedNonce {
				t.Errorf(
					"unexpected nonce\nexpected: [%v]\nactual:  [%v]",
					test.expectedNonce,
					nonce,
				)
			}

			nextNonce := manager.IncrementNonce()

			if nextNonce != test.expectedNextNonce {
				t.Errorf(
					"unexpected nonce\nexpected: [%v]\nactual:  [%v]",
					test.expectedNextNonce,
					nextNonce,
				)
			}
		})
	}
}

type mockTransactor struct {
	nextNonce uint64
}

func (mt *mockTransactor) PendingCodeAt(
	ctx context.Context,
	account common.Address,
) ([]byte, error) {
	panic("not implemented")
}

func (mt *mockTransactor) PendingNonceAt(
	ctx context.Context,
	account common.Address,
) (uint64, error) {
	return mt.nextNonce, nil
}

func (mt *mockTransactor) SuggestGasPrice(
	ctx context.Context,
) (*big.Int, error) {
	panic("not implemented")
}

func (mt *mockTransactor) EstimateGas(
	ctx context.Context,
	call ethereum.CallMsg,
) (gas uint64, err error) {
	panic("not implemented")
}

func (mt *mockTransactor) SendTransaction(
	ctx context.Context,
	tx *types.Transaction,
) error {
	panic("not implemented")
}
