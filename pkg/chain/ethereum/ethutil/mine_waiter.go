package ethutil

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

// MiningWaiter allows to block the execution until the given transaction is
// mined.
type MiningWaiter struct {
	backend bind.DeployBackend
}

// NewMiningWaiter creates a new MiningWaiter instance for the provided
// client backend.
func NewMiningWaiter(backend bind.DeployBackend) *MiningWaiter {
	return &MiningWaiter{backend}
}

// WaitMined blocks the current execution until the transaction with the given
// hash is mined. Execution is blocked until the transaction is mined or until
// the given timeout passes.
func (mw *MiningWaiter) WaitMined(
	timeout time.Duration, 
	tx *types.Transaction,
) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return bind.WaitMined(ctx, mw.backend, tx)
}