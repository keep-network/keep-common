package ethutil

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// NonceManager tracks the nonce for the account and allows to update it after
// each successfully submitted transaction. Tracking the nonce locall is
// required when transactions are submitted from multiple goroutines or when
// multiple Ethereum clients are deployed behind a load balancer, there are no
// sticky sessions and mempool synchronization between them takes some time.
//
// NonceManager provides no synchronization and is NOT safe for concurrent use.
// It is up to the client code to implement the required synchronization.
//
// An example execution might work as follows:
// 1. Obtain transaction lock,
// 2. Calculate CurrentNonce(),
// 3. Submit transaction with the calculated nonce,
// 4. Call IncrementNonce(),
// 5. Release transaction lock.
type NonceManager struct {
	account    common.Address
	transactor bind.ContractTransactor
	localNonce uint64
}

// NewNonceManager creates NonceManager instance for the provided account using
// the provided contract transactor. Contract transactor is used for every
// CurrentNonce execution to check the pending nonce value as seen by the
// Ethereum client.
func NewNonceManager(
	account common.Address,
	transactor bind.ContractTransactor,
) *NonceManager {
	return &NonceManager{
		account:    account,
		transactor: transactor,
		localNonce: 0,
	}
}

// CurrentNonce returns the nonce value that should be used for the next
// transaction. The nonce is evaluated as the higher value from the local
// nonce and pending nonce fetched from the Ethereum client.
//
// CurrentNonce is NOT safe for concurrent use. It is up to the code using this
// function to provide the required synchronization, optionally including
// IncrementNonce call as well.
func (nm *NonceManager) CurrentNonce() (uint64, error) {
	pendingNonce, err := nm.transactor.PendingNonceAt(
		context.TODO(),
		nm.account,
	)
	if err != nil {
		return 0, err
	}

	if pendingNonce < nm.localNonce {
		logger.Infof(
			"local nonce [%v] is higher than pending [%v]; using the local one",
			nm.localNonce,
			pendingNonce,
		)
	}

	if pendingNonce > nm.localNonce {
		logger.Infof(
			"local nonce [%v] is lower than pending [%v]; updating",
			nm.localNonce,
			pendingNonce,
		)

		nm.localNonce = pendingNonce
	}

	return nm.localNonce, nil
}

// IncrementNonce increments the value of the nonce kept locally by one.
// This function is NOT safe for concurrent use. It is up to the client code
// using this function to provide the required synchronization.
func (nm *NonceManager) IncrementNonce() uint64 {
	nm.localNonce++
	return nm.localNonce
}
