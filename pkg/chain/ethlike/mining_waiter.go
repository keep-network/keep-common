package ethlike

import (
	"context"
	"fmt"
	"math/big"
	"time"
)

// MiningWaiter allows to block the execution until the given transaction is
// mined as well as monitor the transaction and perform an appropriate action
// in case it is not mined in the given timeout. This action is meant to
// increase the transaction's chance for being picked up by miners.
//
// Specific action depends on transaction type:
// - legacy pre EIP-1559 transaction: bumps up the gas price by 20%
// - dynamic fee post EIP-1559 transaction: bumps up the gas tip cap by 20%
//   and adjusts the gas fee cap accordingly
type MiningWaiter struct {
	chain         Chain
	checkInterval time.Duration
	maxGasFeeCap  *big.Int
}

// NewMiningWaiter creates a new MiningWaiter instance for the provided
// client backend. It accepts two parameters setting up monitoring rules of the
// transaction mining status.
//
// Check interval is the time given for the transaction to be mined. If the
// transaction is not mined within that time, the mining waiter performs
// appropriate actions to increase their chance for being picked up by miners.
//
// Max gas fee cap specifies the maximum price the client is willing to pay
// per gas, for the transaction to be mined. The offered price can not
// be higher than this value. If the maximum allowed price is reached, no
// further resubmission attempts are performed.
func NewMiningWaiter(
	chain Chain,
	checkInterval time.Duration,
	maxGasFeeCap *big.Int,
) *MiningWaiter {
	return &MiningWaiter{
		chain,
		checkInterval,
		maxGasFeeCap,
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
		receipt, _ := mw.chain.TransactionReceipt(
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

// ResubmitParams determines price parameters for transaction resubmission.
type ResubmitParams struct {
	GasPrice  *big.Int
	GasFeeCap *big.Int
	GasTipCap *big.Int
}

// ResubmitTransactionFn implements the code for resubmitting the transaction
// after mining waiter performs the action. It should guarantee the same nonce
// is used for transaction resubmission.
type ResubmitTransactionFn func(params *ResubmitParams) (*Transaction, error)

// ForceMining blocks until the transaction is mined and performs an appropriate
// action to increase mining probability in the intervals defined by MiningWaiter
// in case the transaction has not been mined yet. It accepts the original
// transaction reference and the function responsible for executing transaction
// resubmission.
func (mw *MiningWaiter) ForceMining(
	originalTransaction *Transaction,
	resubmitFn ResubmitTransactionFn,
) {
	switch originalTransaction.Type {
	case LegacyTxType, AccessListTxType:
		mw.forceMiningLegacyTx(originalTransaction, resubmitFn)
	case DynamicFeeTxType:
		mw.forceMiningDynamicFeeTx(originalTransaction, resubmitFn)
	default:
		logger.Errorf(
			"could not start mining waiter; unsupported transaction type [%v]",
			originalTransaction.Type,
		)
	}
}

func (mw *MiningWaiter) forceMiningLegacyTx(
	originalTransaction *Transaction,
	resubmitFn ResubmitTransactionFn,
) {
	logger.Infof(
		"starting mining waiter for legacy transaction: [%v]",
		originalTransaction.Hash.TerminalString(),
	)

	// For legacy transactions, the `maxGasFeeCap` is considered to be the same
	// as `maxGasPrice`. This is because both parameters means the same:
	// the maximum possible price per gas.
	maxGasPrice := mw.maxGasFeeCap

	// If the original transaction's gas price was higher or equal the max
	// allowed we do nothing; we need to wait for it to be mined.
	if originalTransaction.GasPrice.Cmp(maxGasPrice) >= 0 {
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

		// Transaction mined, we are good.
		if receipt != nil {
			logger.Infof(
				"transaction [%v] mined with status [%v] at block [%v]",
				transaction.Hash.TerminalString(),
				receipt.Status,
				receipt.BlockNumber,
			)
			return
		}

		// Transaction not yet mined, if the previous gas price was the maximum
		// one, we no longer resubmit.
		gasPrice := transaction.GasPrice
		if gasPrice.Cmp(maxGasPrice) == 0 {
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
		if gasPrice.Cmp(maxGasPrice) > 0 {
			gasPrice = maxGasPrice
		}

		// Transaction not yet mined and we are still under the maximum allowed
		// gas price; resubmitting transaction with 20% higher gas price
		// evaluated earlier.
		logger.Infof(
			"resubmitting previous transaction [%v] "+
				"with a higher gas price [%v]",
			transaction.Hash.TerminalString(),
			gasPrice,
		)
		transaction, err = resubmitFn(
			&ResubmitParams{
				GasPrice: gasPrice,
			},
		)
		if err != nil {
			logger.Warningf(
				"could not resubmit TX with a higher gas price: [%v]",
				err,
			)
			return
		}
	}
}

func (mw *MiningWaiter) forceMiningDynamicFeeTx(
	originalTransaction *Transaction,
	resubmitFn ResubmitTransactionFn,
) {
	logger.Infof(
		"starting mining waiter for dynamic fee transaction: [%v]",
		originalTransaction.Hash.TerminalString(),
	)

	// If the original transaction's gas fee cap was higher or equal the max
	// allowed we do nothing; we need to wait for it to be mined.
	if originalTransaction.GasFeeCap.Cmp(mw.maxGasFeeCap) >= 0 {
		logger.Infof(
			"original transaction gas fee cap is higher than the max allowed; " +
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

		// Transaction mined, we are good.
		if receipt != nil {
			logger.Infof(
				"transaction [%v] mined with status [%v] at block [%v]",
				transaction.Hash.TerminalString(),
				receipt.Status,
				receipt.BlockNumber,
			)
			return
		}

		// Transaction not yet mined, if the previous gas fee cap was the
		// maximum one, we no longer resubmit.
		oldGasFeeCap := transaction.GasFeeCap
		if oldGasFeeCap.Cmp(mw.maxGasFeeCap) == 0 {
			logger.Infof(
				"reached the maximum allowed gas fee cap; " +
					"stopping resubmissions",
			)
			return
		}

		// Increase the gas tip cap by 20%. A minimum increase by 10% comparing
		// to the previous value is required for transaction replacement to be
		// accepted by miners as mentioned in:
		// https://github.com/ethereum/go-ethereum/pull/22898/files#r636583352.
		// We increase it even more than the required level to greatly increase
		// the transaction's chance for being picked up by miners.
		oldGasTipCap := transaction.GasTipCap
		newGasTipCap := new(big.Int).Add(
			oldGasTipCap,
			new(big.Int).Div(oldGasTipCap, big.NewInt(5)),
		)

		// Fetch latest base fee from the chain. It's needed to compute the
		// new value of gas fee cap.
		latestBaseFee, err := mw.latestBaseFee()
		if err != nil {
			logger.Errorf("could not get latest base fee: [%v]", err)
			continue
		}

		// Compute new value of gas fee cap using the latest base fee
		// and new gas tip cap. The `gasFeeCap = 2 * baseFee + gasTipCap`
		// equation originates from `go-ethereum` which estimates this
		// parameter in that way.
		// See: https://github.com/ethereum/go-ethereum/pull/23038.
		// Having the `baseFee` taken twice means the `gasFeeCap` should
		// be resilient for six consecutive increases of the `baseFee`.
		// This is because `baseFee` can be increased by 12.5% at maximum
		// within a single increase.
		newGasFeeCap := new(big.Int).Add(
			new(big.Int).Mul(latestBaseFee, big.NewInt(2)),
			newGasTipCap,
		)

		// The new gas fee cap value needs to be at least 10% bigger
		// than the old value. Otherwise, the transaction replacement
		// won't be accepted by miners as mentioned in:
		// https://github.com/ethereum/go-ethereum/pull/22898/files#r636583352.
		// If that's the case (e.g. the `baseFee` dramatically decreased since
		// the previous transaction) we need to set the new gas fee cap value
		// to the minimum value acceptable by miners.
		requiredGasFeeCapThreshold := new(big.Int).Add(
			oldGasFeeCap,
			new(big.Int).Div(oldGasFeeCap, big.NewInt(10)),
		)
		if newGasFeeCap.Cmp(requiredGasFeeCapThreshold) < 0 {
			newGasFeeCap = requiredGasFeeCapThreshold
		}

		// If we reached the maximum allowed gas fee cap, submit one more time
		// with the maximum.
		if newGasFeeCap.Cmp(mw.maxGasFeeCap) > 0 {
			newGasFeeCap = mw.maxGasFeeCap

			// Check if the threshold condition is fulfilled once again.
			// If the maximum allowed gas fee cap is below the threshold,
			// there is no sense to submit the transaction as it won't
			// be accepted by the miners.
			if newGasFeeCap.Cmp(requiredGasFeeCapThreshold) < 0 {
				logger.Infof(
					"could not fulfill required gas fee cap threshold as " +
						"the maximum gas fee cap value defined in config " +
						"has been reached; " +
						"stopping resubmissions",
				)
				return
			}
		}

		// Transaction not yet mined and we are still under the maximum allowed
		// gas fee cap; resubmitting transaction with gas fee and tip parameters
		// evaluated earlier.
		logger.Infof(
			"resubmitting previous transaction [%v] "+
				"with a higher gas fee cap [%v] and tip cap [%v]",
			transaction.Hash.TerminalString(),
			newGasFeeCap,
			newGasTipCap,
		)
		transaction, err = resubmitFn(
			&ResubmitParams{
				GasFeeCap: newGasFeeCap,
				GasTipCap: newGasTipCap,
			},
		)
		if err != nil {
			logger.Warningf(
				"could not resubmit TX with a higher "+
					"gas fee cap and tip cap: [%v]",
				err,
			)
			return
		}
	}
}

func (mw *MiningWaiter) latestBaseFee() (*big.Int, error) {
	latestBlock, err := mw.chain.BlockByNumber(
		context.Background(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("could not get latest block: [%v]", err)
	}

	baseFee := latestBlock.BaseFee
	if baseFee == nil {
		return nil, fmt.Errorf("not an EIP-1559 block")
	}

	return baseFee, nil
}
