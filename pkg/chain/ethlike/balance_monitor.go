package ethlike

import (
	"context"
	"time"
)

// BalanceSource provides a balance info for the given address.
type BalanceSource func(address Address) (*Token, error)

// BalanceMonitor provides the possibility to monitor balances for given
// accounts.
type BalanceMonitor struct {
	balanceSource BalanceSource
}

// NewBalanceMonitor creates a new instance of the balance monitor.
func NewBalanceMonitor(balanceSource BalanceSource) *BalanceMonitor {
	return &BalanceMonitor{balanceSource}
}

// Observe starts a process which checks the address balance with the given
// tick and triggers an alert in case the balance falls below the
// alert threshold value.
func (bm *BalanceMonitor) Observe(
	ctx context.Context,
	address Address,
	alertThreshold *Token,
	tick time.Duration,
) {
	check := func() {
		balance, err := bm.balanceSource(address)
		if err != nil {
			logger.Errorf("balance monitor error: [%v]", err)
			return
		}

		if balance.Cmp(alertThreshold.Int) == -1 {
			logger.Errorf(
				"balance for account [%v] is below [%v]; "+
					"account should be funded",
				address.TerminalString(),
				alertThreshold.Text(10),
			)
		}
	}

	go func() {
		ticker := time.NewTicker(tick)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				check()
			case <-ctx.Done():
				return
			}
		}
	}()
}
