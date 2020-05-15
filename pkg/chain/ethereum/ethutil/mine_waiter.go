package ethutil

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

// MiningWaiter allows to block the execution until the given transaction is
// mined as well as monitor the transaction and bump up the gas price in case
// it is not mined in the given timeout.
type MiningWaiter struct {
	backend       bind.DeployBackend
	checkInterval time.Duration
	maxGasPrice   *big.Int
}

// NewMiningWaiter creates a new MiningWaiter instance for the provided
// client backend. It accepts two parameters setting up monitoring rules of the
// transaction mining status.
//
// Check interval is the time given for the transaction to be mined. If the
// transaction is not mined within that time, the gas price is increased by
// 20% and transaction is replaced with the one with a higher gas price.
//
// Max gas price specifies the maximum gas price the client is willing to pay
// for the transaction to be mined. The offered transaction gas price can not
// be higher than this value. If the maximum allowed gas price is reached, no
// further resubmission attempts are performed.
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

	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()

	for {
		receipt, _ := mw.backend.TransactionReceipt(context.TODO(), tx.Hash())
		if receipt != nil {
			return receipt, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
		}
	}
}

// ResubmitTransactionFn implements the code for resubmitting the transaction
// with the higher gas price. It should guarantee the same nonce is used for
// transaction resubmission.
type ResubmitTransactionFn func(gasPrice *big.Int) (*types.Transaction, error)

// ForceMining blocks until the transaction is mined and bumps up the gas price
// by 20% in the intervals defined by MiningWaiter in case the transaction has
// not been mined yet. It accepts the original transaction reference and the
// function responsible for executing transaction resubmission.
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

		// transaction not yet mined, add 20% to the previous gas price
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

		// transaction not yet mined and we are still under the maximum allowed
		// gas price; resubmitting transaction with 20% higher gas price
		// evaluated earlier
		logger.Infof(
			"resubmitting previous transaction [%v] with a higher gas price [%v]",
			transaction.Hash().TerminalString(),
			gasPrice,
		)
		transaction, err = resubmitFn(gasPrice)
		if err != nil {
			logger.Errorf("failed resubmitting TX with a higher gas price: [%v]", err)
			return
		}
	}
}
