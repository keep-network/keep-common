package celo

import (
	"fmt"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"reflect"
	"testing"

	"github.com/celo-org/celo-blockchain/common"
)

func TestContractAddress(t *testing.T) {
	contractName1 := "KeepECDSAContract"
	validContractAddressString := "0xbb2Ea17985f13D43e3AEC3963506A1B25ADDd57F"
	validContractAddress := common.HexToAddress(validContractAddressString)

	contractName2 := "InvalidContract"
	invalidHex := "0xZZZ"

	config := &Config{
		Config: ethlike.Config{
			ContractAddresses: map[string]string{
				contractName1: validContractAddressString,
				contractName2: invalidHex,
			},
		},
	}

	var tests = map[string]struct {
		contractName    string
		expectedAddress *common.Address
		expectedError   error
	}{
		"contract name matching valid configuration": {
			contractName:    contractName1,
			expectedAddress: &validContractAddress,
		},
		"invalid contract hex address": {
			contractName:    contractName2,
			expectedAddress: nil,
			expectedError: fmt.Errorf(
				"configured address [%v] for contract [%v] "+
					"is not valid hex address",
				invalidHex,
				contractName2,
			),
		},
		"missing contract configuration": {
			contractName:    "Peekaboo",
			expectedAddress: nil,
			expectedError: fmt.Errorf("no address information " +
				"for [Peekaboo] in configuration"),
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {

			actualAddress, err := config.ContractAddress(test.contractName)
			if !reflect.DeepEqual(test.expectedError, err) {
				t.Errorf(
					"unexpected error\nexpected: %v\nactual:   %v\n",
					test.expectedError,
					err,
				)
			}

			if !reflect.DeepEqual(test.expectedAddress, actualAddress) {
				t.Errorf(
					"unexpected address\nexpected: %v\nactual:   %v\n",
					test.expectedAddress,
					actualAddress,
				)
			}
		})
	}
}
