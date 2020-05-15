package ethutil

import (
	"context"
	"time"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

// MiningWaiter allows to block the execution until the given transaction is
// mined.
type MiningWaiter struct {
	backend bind.DeployBackend
	checkInterval time.Duration
	maxGasPrice *big.Int
}

// NewMiningWaiter creates a new MiningWaiter instance for the provided
// client backend.
func NewMiningWaiter(
	backend bind.DeployBackend,
	checkInterval time.Duration,
	maxGasPrice *big.Int,
) *MiningWaiter {
	return &MiningWaiter{
		backend,
		checkInterval,
		maxGasPrice,
	}
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

type ResubmitTransactionFn func(gasPrice *big.Int) (*types.Transaction, error)

func (mw MiningWaiter) ForceMining(
	originalTransaction *types.Transaction,
	resubmitFn ResubmitTransactionFn,
) {
	transaction := originalTransaction
	for {
		receipt, err := mw.WaitMined(mw.checkInterval, transaction)
		if err != nil {
			logger.Infof(
				"transaction [%v] not yet mined: [%v]",
				transaction.Hash().TerminalString(),
				err,
			)
		}
		
		// transaction mined, we are good
		if receipt != nil {
			logger.Infof(
				"transaction [%v] mined with status [%v] at block [%v]",
				transaction.Hash().TerminalString(),
				receipt.Status,
				receipt.BlockNumber,
			)
			return
		}

		// add 20% to the previous gas price
		gasPrice := transaction.GasPrice()
		twentyPercent := new(big.Int).Div(gasPrice, big.NewInt(5))
		gasPrice = new(big.Int).Add(gasPrice, twentyPercent)
		
		// transaction not yet mined but we reached the maximum allowed gas 
		// price; giving up, we need to wait for the last submitted TX to be
		// mined
		if gasPrice.Cmp(mw.maxGasPrice) > 0 {
			logger.Infof("reached the maximum allowed gas price; stopping resubmissions")
			return
		}

		// transaction not yet mined and we can still increase gas price
		// resubmitting transaction with 20% higher gas price
		logger.Infof(
			"resubmitting previous transaction [%v] with higher gas price [%v]",
			transaction.Hash().TerminalString(),
			gasPrice,
		)
	
		transaction, err = resubmitFn(gasPrice)
		if err != nil {
			logger.Errorf("failed resubmitting TX with higher gas price: [%v]", err)
			return
		}
	}
}