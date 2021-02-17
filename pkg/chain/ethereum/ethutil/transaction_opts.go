package ethutil

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// TransactionOptions represents custom transaction options which can be used
// while invoking contracts methods.
type TransactionOptions struct {
	// GasLimit specifies a gas limit to set on a transaction call; should be
	// ignored if set to 0.
	GasLimit uint64
	// GasPrice specifies a gas price to set on a transaction call; should be
	// ignored if set to nil.
	GasPrice *big.Int
}

// Apply takes a bind.TransactOpts pointer and applies the options described in
// TransactionOptions to it. Note that a GasLimit of 0 or a GasPrice of nil are
// not applied to the passed options; these values indicate that the options
// should remain unchanged.
func (to TransactionOptions) Apply(transactorOptions *bind.TransactOpts) {
	if customGasLimit := to.GasLimit; customGasLimit != 0 {
		transactorOptions.GasLimit = customGasLimit
	}
	if customGasPrice := to.GasPrice; customGasPrice != nil {
		transactorOptions.GasPrice = customGasPrice
	}
}
