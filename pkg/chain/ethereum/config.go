package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// Account is a struct that contains the configuration for accessing an
// Ethereum network and a contract on the network.
type Account struct {
	// Keyfile is a full path to a key file.  Normally this file is one of the
	// imported keys in your local Ethereum server.  It can normally be found in
	// a directory <some-path>/data/keystore/ and starts with its creation date
	// "UTC--.*".
	KeyFile string

	// KeyFilePassword is the password used to unlock the account specified in
	// KeyFile.
	KeyFilePassword string
}

// Config is a struct that contains the configuration needed to connect to an
// Ethereum node.   This information will give access to an Ethereum network.
type Config struct {
	// Example: "ws://192.168.0.157:8546".
	URL string

	// Example: "http://192.168.0.157:8545".
	URLRPC string

	// A  map from contract names to contract addresses.
	ContractAddresses map[string]string

	Account Account

	// MiningCheckInterval is the interval in which transaction
	// mining status is checked. If the transaction is not mined within this
	// time, the gas price is increased and transaction is resubmitted.
	MiningCheckInterval int

	// MaxGasPrice specifies the maximum gas price the client is
	// willing to pay for the transaction to be mined. The offered transaction
	// gas price can not be higher than the max gas price value. If the maximum
	// allowed gas price is reached, no further resubmission attempts are
	// performed.
	MaxGasPrice uint64

	// RequestsPerSecondLimit sets the maximum average number of requests
	// per second which can be executed against the Ethereum node.
	// All types of chain requests are rate-limited,
	// including view function calls.
	RequestsPerSecondLimit int

	// ConcurrencyLimit sets the maximum number of concurrent requests which
	// can be executed against the Ethereum node at the same time.
	// This limit affects all types of chain requests,
	// including view function calls.
	ConcurrencyLimit int

	// BalanceAlertThreshold defines a minimum value of the operator's account
	// balance below which an alert will be triggered.
	BalanceAlertThreshold uint64
}

// ContractAddress finds a given contract's address configuration and returns it
// as ethereum Address.
func (c Config) ContractAddress(contractName string) (*common.Address, error) {
	addressString, exists := c.ContractAddresses[contractName]
	if !exists {
		return nil, fmt.Errorf(
			"no address information for [%v] in configuration",
			contractName,
		)
	}

	if !common.IsHexAddress(addressString) {
		return nil, fmt.Errorf(
			"configured address [%v] for contract [%v] is not valid hex address",
			addressString,
			contractName,
		)
	}

	address := common.HexToAddress(addressString)
	return &address, nil
}
