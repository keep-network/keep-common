package celoutil

import (
	"context"
	"github.com/celo-org/celo-blockchain/accounts/abi/bind"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/keep-network/keep-common/pkg/chain/celo"
	"math/big"
	"time"
)

var (
	// DefaultMiningCheckInterval is the default interval in which transaction
	// mining status is checked. If the transaction is not mined within this
	// time, the gas price is increased and transaction is resubmitted.
	// This value can be overwritten in the configuration file.
	DefaultMiningCheckInterval = 60 * time.Second

	// DefaultMaxGasPrice specifies the default maximum gas price the client is
	// willing to pay for the transaction to be mined. The offered transaction
	// gas price can not be higher than the max gas price value. If the maximum
	// allowed gas price is reached, no further resubmission attempts are
	// performed. This value can be overwritten in the configuration file.
	DefaultMaxGasPrice = big.NewInt(500000000000) // 500 Gwei
)

// MiningWaiter allows to block the execution until the given transaction is
// mined as well as monitor the transaction and bump up the gas price in case
// it is not mined in the given timeout.
type MiningWaiter struct {
	client        CeloClient
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
	client CeloClient,
	config celo.Config,
) *MiningWaiter {
	checkInterval := DefaultMiningCheckInterval
	maxGasPrice := DefaultMaxGasPrice
	if config.MiningCheckInterval != 0 {
		checkInterval = time.Duration(config.MiningCheckInterval) * time.Second
	}
	if config.MaxGasPrice != nil {
		maxGasPrice = config.MaxGasPrice.Int
	}

	logger.Infof("using [%v] mining check interval", checkInterval)
	logger.Infof("using [%v] wei max gas price", maxGasPrice)

	return &MiningWaiter{
		client,
		checkInterval,
		maxGasPrice,
	}
}

// waitMined blocks the current execution until the transaction with the given
// hash is mined. Execution is blocked until the transaction is mined or until
// the given timeout passes.
func (mw *MiningWaiter) waitMined(
	timeout time.Duration,
	transaction *types.Transaction,
) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()

	for {
		receipt, _ := mw.client.TransactionReceipt(
			context.TODO(),
			transaction.Hash(),
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
type ResubmitTransactionFn func(
	newTransactorOptions *bind.TransactOpts,
) (*types.Transaction, error)

// ForceMining blocks until the transaction is mined and bumps up the gas price
// by 20% in the intervals defined by MiningWaiter in case the transaction has
// not been mined yet. It accepts the original transaction reference and the
// function responsible for executing transaction resubmission.
func (mw *MiningWaiter) ForceMining(
	originalTransaction *types.Transaction,
	originalTransactorOptions *bind.TransactOpts,
	resubmitFn ResubmitTransactionFn,
) {
	// If the original transaction's gas price was higher or equal the max
	// allowed we do nothing; we need to wait for it to be mined.
	if originalTransaction.GasPrice().Cmp(mw.maxGasPrice) >= 0 {
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
				transaction.Hash().TerminalString(),
				err,
			)
		}

		// Transaction mined, we are good.
		if receipt != nil {
			logger.Infof(
				"transaction [%v] mined with status [%v] at block [%v]",
				transaction.Hash().TerminalString(),
				receipt.Status,
				receipt.BlockNumber,
			)
			return
		}

		// Transaction not yet mined, if the previous gas price was the maximum
		// one, we no longer resubmit.
		gasPrice := transaction.GasPrice()
		if gasPrice.Cmp(mw.maxGasPrice) == 0 {
			logger.Infof(
				"reached the maximum allowed gas price; " +
					"stopping resubmissions",
			)
			return
		}

		// If we still have some margin, add 20% to the previous gas price.
		twentyPercent := new(big.Int).Div(gasPrice, big.NewInt(5))
		gasPrice = new(big.Int).Add(gasPrice, twentyPercent)

		// If we reached the maximum allowed gas price, submit one more time
		// with the maximum.
		if gasPrice.Cmp(mw.maxGasPrice) > 0 {
			gasPrice = mw.maxGasPrice
		}

		// Transaction not yet mined and we are still under the maximum allowed
		// gas price; resubmitting transaction with 20% higher gas price
		// evaluated earlier.
		logger.Infof(
			"resubmitting previous transaction [%v] "+
				"with a higher gas price [%v]",
			transaction.Hash().TerminalString(),
			gasPrice,
		)

		// Copy transactor options.
		newTransactorOptions := new(bind.TransactOpts)
		*newTransactorOptions = *originalTransactorOptions
		newTransactorOptions.GasLimit = originalTransaction.Gas()
		newTransactorOptions.GasPrice = gasPrice

		transaction, err = resubmitFn(newTransactorOptions)
		if err != nil {
			logger.Warningf(
				"could not resubmit TX with a higher gas price: [%v]",
				err,
			)
			return
		}
	}
}
