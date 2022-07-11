package ethereum

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
)

// Config is a struct that contains the configuration needed to connect to an
// Ethereum node. This information will give access to an Ethereum network.
type Config struct {
	ethlike.Config

	// MaxGasFeeCap specifies the maximum gas fee cap the client is
	// willing to pay for the transaction to be mined. The offered transaction
	// gas cost can not be higher than the max gas fee cap value. If the maximum
	// allowed gas fee cap is reached, no further resubmission attempts are
	// performed. This value is used for all types of Ethereum transactions.
	// For legacy transactions, this value works as a maximum gas price
	// and for EIP-1559 transactions, this value works as max gas fee cap.
	MaxGasFeeCap *Wei

	// BalanceAlertThreshold defines a minimum value of the operator's
	// account balance below which an alert will be triggered.
	BalanceAlertThreshold *Wei
}

// ErrAddressNotConfigured is an error that is returned when an address for the given
// contract name was not found in the Config.
var ErrAddressNotConfigured = errors.New("address not configured")

// ContractAddress finds a given contract's address configuration and returns it
// as Ethereum address.
func (c *Config) ContractAddress(contractName string) (common.Address, error) {
	addressString, exists := c.ContractAddresses[contractName]
	if !exists {
		return common.Address{}, ErrAddressNotConfigured
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
