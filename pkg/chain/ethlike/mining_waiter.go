package ethlike

import (
	"context"
	"math/big"
	"time"
)

// MiningWaiter allows to block the execution until the given transaction is
// mined as well as monitor the transaction and bump up the gas price in case
// it is not mined in the given timeout.
type MiningWaiter struct {
	txReader      TransactionReader
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
	txReader TransactionReader,
	checkInterval time.Duration,
	maxGasPrice *big.Int,
) *MiningWaiter {
	return &MiningWaiter{
		txReader,
		checkInterval,
		maxGasPrice,
	}
}

// waitMined blocks the current execution until the transaction with the given
// hash is mined. Execution is blocked until the transaction is mined or until
// the given timeout passes.
func (mw *MiningWaiter) waitMined(
	timeout time.Duration,
	transaction *Transaction,
) (*Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()

	for {
		receipt, _ := mw.txReader.TransactionReceipt(
			context.TODO(),
			transaction.Hash,
		)
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
type ResubmitTransactionFn func(gasPrice *big.Int) (*Transaction, error)

// ForceMining blocks until the transaction is mined and bumps up the gas price
// by 20% in the intervals defined by MiningWaiter in case the transaction has
// not been mined yet. It accepts the original transaction reference and the
// function responsible for executing transaction resubmission.
func (mw MiningWaiter) ForceMining(
	originalTransaction *Transaction,
	resubmitFn ResubmitTransactionFn,
) {
	// if the original transaction's gas price was higher or equal the max
	// allowed we do nothing; we need to wait for it to be mined
	if originalTransaction.GasPrice.Cmp(mw.maxGasPrice) >= 0 {
		logger.Infof(
			"original transaction gas price is higher than the max allowed; " +
				"skipping resubmissions",
		)
		return
	}

	transaction := originalTransaction
	for {
		receipt, err := mw.waitMined(mw.checkInterval, transaction)
		if err != nil {
			logger.Infof(
				"transaction [%v] not yet mined: [%v]",
				transaction.Hash.TerminalString(),
				err,
			)
		}

		// transaction mined, we are good
		if receipt != nil {
			logger.Infof(
				"transaction [%v] mined with status [%v] at block [%v]",
				transaction.Hash.TerminalString(),
				receipt.Status,
				receipt.BlockNumber,
			)
			return
		}

		// transaction not yet mined, if the previous gas price was the maximum
		// one, we no longer resubmit
		gasPrice := transaction.GasPrice
		if gasPrice.Cmp(mw.maxGasPrice) == 0 {
			logger.Infof("reached the maximum allowed gas price; stopping resubmissions")
			return
		}

		// if we still have some margin, add 20% to the previous gas price
		twentyPercent := new(big.Int).Div(gasPrice, big.NewInt(5))
		gasPrice = new(big.Int).Add(gasPrice, twentyPercent)

		// if we reached the maximum allowed gas price, submit one more time
		// with the maximum
		if gasPrice.Cmp(mw.maxGasPrice) > 0 {
			gasPrice = mw.maxGasPrice
		}

		// transaction not yet mined and we are still under the maximum allowed
		// gas price; resubmitting transaction with 20% higher gas price
		// evaluated earlier
		logger.Infof(
			"resubmitting previous transaction [%v] with a higher gas price [%v]",
			transaction.Hash.TerminalString(),
			gasPrice,
		)
		transaction, err = resubmitFn(gasPrice)
		if err != nil {
			logger.Warningf("could not resubmit TX with a higher gas price: [%v]", err)
			return
		}
	}
}
