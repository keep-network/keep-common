package celo

import (
	"fmt"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
)

// Config is a struct that contains the configuration needed to connect to an
// Celo node. This information will give access to a Celo network.
type Config struct {
	ethlike.CommonConfig
}

// ContractAddress finds a given contract's address configuration and returns it
// as Celo address.
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
			"configured address [%v] for contract [%v] "+
				"is not valid hex address",
			addressString,
			contractName,
		)
	}

	address := common.HexToAddress(addressString)
	return &address, nil
}
