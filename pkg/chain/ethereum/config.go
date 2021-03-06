package ethereum

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
)

// Config is a struct that contains the configuration needed to connect to an
// Ethereum node. This information will give access to an Ethereum network.
type Config struct {
	ethlike.Config

	// MaxGasPrice specifies the maximum gas price the client is
	// willing to pay for the transaction to be mined. The offered transaction
	// gas price can not be higher than the max gas price value. If the maximum
	// allowed gas price is reached, no further resubmission attempts are
	// performed.
	MaxGasPrice *Wei

	// BalanceAlertThreshold defines a minimum value of the operator's
	// account balance below which an alert will be triggered.
	BalanceAlertThreshold *Wei
}

// ContractAddress finds a given contract's address configuration and returns it
// as Ethereum address.
func (c *Config) ContractAddress(contractName string) (common.Address, error) {
	addressString, exists := c.ContractAddresses[contractName]
	if !exists {
		return common.Address{}, fmt.Errorf(
			"no address information for [%v] in configuration",
			contractName,
		)
	}

	if !common.IsHexAddress(addressString) {
		return common.Address{}, fmt.Errorf(
			"configured address [%v] for contract [%v] "+
				"is not valid hex address",
			addressString,
			contractName,
		)
	}

	address := common.HexToAddress(addressString)
	return address, nil
}
