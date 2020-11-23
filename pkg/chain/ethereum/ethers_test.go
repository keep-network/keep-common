package ethereum

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
)

func TestUnmarshalText(t *testing.T) {

	int5000ether, _ := new(big.Int).SetString("5000000000000000000000", 10)

	var tests = map[string]struct {
		value          string
		expectedResult *big.Int
		expectedError  error
	}{
		"decimal value": {
			value:          "0.9",
			expectedResult: big.NewInt(0),
		},
		"lowest value": {
			value:          "1",
			expectedResult: big.NewInt(1),
		},
		"unit: wei": {
			value:          "4 wei",
			expectedResult: big.NewInt(4),
		},
		"unit: gwei": {
			value:          "30 gwei",
			expectedResult: big.NewInt(30000000000),
		},
		"unit: ether": {
			value:          "2 ether",
			expectedResult: big.NewInt(2000000000000000000),
		},
		"unit: mixed case": {
			value:          "5 GWei",
			expectedResult: big.NewInt(5000000000),
		},
		"decimal wei": {
			value:          "2.99 wei",
			expectedResult: big.NewInt(2),
		},
		"decimal ether": {
			value:          "0.8 ether",
			expectedResult: big.NewInt(800000000000000000),
		},
		"no space": {
			value:          "9ether",
			expectedResult: big.NewInt(9000000000000000000),
		},
		"double space": {
			value:          "100  Gwei",
			expectedResult: big.NewInt(100000000000),
		},
		"int overflow amount": {
			value:          "5000000000000000000000",
			expectedResult: int5000ether,
		},
		"int overflow amount after conversion": {
			value:          "5000 ether",
			expectedResult: int5000ether,
		},
		"invalid comma delimeter": {
			value:         "3,5 ether",
			expectedError: fmt.Errorf("failed to parse value: [3,5 ether]"),
		},
		"invalid thousand separator": {
			value:         "4 500 gwei",
			expectedError: fmt.Errorf("failed to parse value: [4 500 gwei]"),
		},
		"invalid unit: ETH": {
			value:         "6 ETH",
			expectedError: fmt.Errorf("invalid unit: ETH; please use one of: wei, Gwei, ether"),
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {

			e := &Ethers{}
			err := e.UnmarshalText([]byte(test.value))
			if test.expectedError != nil {
				if !reflect.DeepEqual(test.expectedError, err) {
					t.Errorf(
						"invalid error\nexpected: %v\nactual:   %v",
						test.expectedError,
						err,
					)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if test.expectedResult != nil && test.expectedResult.Cmp(e.Int) != 0 {
				t.Errorf(
					"invalid value\nexpected: %v\nactual:   %v",
					test.expectedResult.String(),
					e.Int.String(),
				)
			}
		})
	}
}
