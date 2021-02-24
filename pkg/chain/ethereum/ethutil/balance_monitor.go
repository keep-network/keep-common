package ethutil

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"time"
)

// BalanceSource provides a balance info for the given address.
type BalanceSource func(address common.Address) (*ethereum.Wei, error)

// BalanceMonitor provides the possibility to monitor balances for given
// accounts.
type BalanceMonitor struct {
	delegate *ethlike.BalanceMonitor
}

// NewBalanceMonitor creates a new instance of the balance monitor.
func NewBalanceMonitor(balanceSource BalanceSource) *BalanceMonitor {
	balanceSourceAdapter := func(
		address ethlike.Address,
	) (*ethlike.Token, error) {
		balance, err := balanceSource(common.Address(address))
		if err != nil {
			return nil, err
		}

		return &balance.Token, err
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
	address common.Address,
	alertThreshold *ethereum.Wei,
	tick time.Duration,
) {
	bm.delegate.Observe(
		ctx,
		ethlike.Address(address),
		&alertThreshold.Token,
		tick,
	)
}
