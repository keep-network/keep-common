package ethereum

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
)

func TestUnmarshalTextValue(t *testing.T) {
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
		"missing unit": {
			value:          "702",
			expectedResult: big.NewInt(702),
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
			value:          "2.9 wei",
			expectedResult: big.NewInt(2),
		},
		"decimal ether": {
			value:          "0.8 ether",
			expectedResult: big.NewInt(800000000000000000),
		},
		"multiple decimal digits": {
			value:          "5.6789 Gwei",
			expectedResult: big.NewInt(5678900000),
		},
		"missing decimal digit": {
			value:          "6. Gwei",
			expectedResult: big.NewInt(6000000000),
		},
		"no space": {
			value:          "9ether",
			expectedResult: big.NewInt(9000000000000000000),
		},
		"int overflow amount": {
			value:          "5000000000000000000000",
			expectedResult: int5000ether,
		},
		"int overflow amount after conversion": {
			value:          "5000 ether",
			expectedResult: int5000ether,
		},
		"double space": {
			value:         "100  Gwei",
			expectedError: fmt.Errorf("failed to parse value: [100  Gwei]"),
		},
		"leading space": {
			value:         " 3 wei",
			expectedError: fmt.Errorf("failed to parse value: [ 3 wei]"),
		},
		"trailing space": {
			value:         "3 wei ",
			expectedError: fmt.Errorf("failed to parse value: [3 wei ]"),
		},

		"invalid comma delimeter": {
			value:         "3,5 ether",
			expectedError: fmt.Errorf("failed to parse value: [3,5 ether]"),
		},
		"only decimal number": {
			value:         ".7 Gwei",
			expectedError: fmt.Errorf("failed to parse value: [.7 Gwei]"),
		},
		"duplicated delimeters": {
			value:         "3..4 wei",
			expectedError: fmt.Errorf("failed to parse value: [3..4 wei]"),
		},
		"multiple decimals": {
			value:         "3.4.5 wei",
			expectedError: fmt.Errorf("failed to parse value: [3.4.5 wei]"),
		},
		"invalid thousand separator": {
			value:         "4 500 gwei",
			expectedError: fmt.Errorf("failed to parse value: [4 500 gwei]"),
		},
		"two values": {
			value:         "3 wei2wei",
			expectedError: fmt.Errorf("invalid unit: wei2wei; please use one of: ether, gwei, wei"),
		},
		"two values separated with space": {
			value:         "3 wei 2wei",
			expectedError: fmt.Errorf("failed to parse value: [3 wei 2wei]"),
		},
		"two values separated with break line": {
			value:         "3 wei\n2wei",
			expectedError: fmt.Errorf("failed to parse value: [3 wei\n2wei]"),
		},
		"invalid unit: ETH": {
			value:         "6 ETH",
			expectedError: fmt.Errorf("invalid unit: ETH; please use one of: ether, gwei, wei"),
		},
		"invalid unit: weinot": {
			value:         "100 weinot",
			expectedError: fmt.Errorf("invalid unit: weinot; please use one of: ether, gwei, wei"),
		},
		"invalid unit: notawei": {
			value:         "100 notawei",
			expectedError: fmt.Errorf("invalid unit: notawei; please use one of: ether, gwei, wei"),
		},
		"only unit": {
			value:         "wei",
			expectedError: fmt.Errorf("failed to parse value: [wei]"),
		},
		"invalid number": {
			value:         "one wei",
			expectedError: fmt.Errorf("failed to parse value: [one wei]"),
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			w := Wei{}
			err := w.UnmarshalText([]byte(test.value))
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

			if test.expectedResult != nil && test.expectedResult.Cmp(w.Int) != 0 {
				t.Errorf(
					"invalid value\nexpected: %v\nactual:   %v",
					test.expectedResult.String(),
					w.Int.String(),
				)
			}
		})
	}
}
