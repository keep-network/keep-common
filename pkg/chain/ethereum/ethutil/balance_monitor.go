package ethutil

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
	"time"
)

// BalanceSource provides a balance info for the given address.
type BalanceSource func(address common.Address) (*big.Int, error)

// BalanceMonitor provides the possibility to monitor balances for given
// accounts.
type BalanceMonitor struct {
	delegate *ethlike.BalanceMonitor
}

// NewBalanceMonitor creates a new instance of the balance monitor.
func NewBalanceMonitor(balanceSource BalanceSource) *BalanceMonitor {
	balanceSourceAdapter := func(address ethlike.Address) (*big.Int, error) {
		return balanceSource(common.Address(address))
	}

	return &BalanceMonitor{
		ethlike.NewBalanceMonitor(balanceSourceAdapter),
	}
}

// Observe starts a process which checks the address balance with the given
// tick and triggers an alert in case the balance falls below the
// alert threshold value.
func (bm *BalanceMonitor) Observe(
	ctx context.Context,
	address string,
	alertThreshold *big.Int,
	tick time.Duration,
) {
	bm.delegate.Observe(
		ctx,
		ethlike.Address(common.HexToAddress(address)),
		alertThreshold,
		tick,
	)
}
