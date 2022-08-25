package flag

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	pflag "github.com/spf13/pflag"
)

func TestWeiVarFlag_Set(t *testing.T) {
	defaultValue := *ethereum.WrapWei(big.NewInt(0))
	flagName := "amount"

	tests := map[string]struct {
		value         string
		expectedError error
		expectedValue *big.Int
	}{
		"value without unit": {
			value:         "101",
			expectedValue: big.NewInt(101),
		},
		"value with wei unit": {
			value:         "202 wei",
			expectedValue: big.NewInt(202),
		},
		"value with gwei unit": {
			value:         "303 gwei",
			expectedValue: big.NewInt(303000000000),
		},
		"value with ether unit": {
			value:         "0.9 ether",
			expectedValue: big.NewInt(900000000000000000),
		},
		"value with invalid comma delimiter": {
			value: "3,5 ether",
			expectedError: fmt.Errorf(
				"invalid argument \"3,5 ether\" for \"--%s\" flag: failed to parse value: [3,5 ether]",
				flagName,
			),
		},
		"value with invalid unit": {
			value: "10 bei",
			expectedError: fmt.Errorf(
				"invalid argument \"10 bei\" for \"--%s\" flag: invalid unit: bei; please use one of: ether, gwei, wei",
				flagName,
			),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			flags := pflag.NewFlagSet("flag-set-"+testName, pflag.PanicOnError)

			var valueDest ethereum.Wei

			WeiVarFlag(flags, &valueDest, flagName, defaultValue, "")

			err := flags.Set(flagName, test.value)

			if !reflect.DeepEqual(test.expectedError, err) {
				t.Errorf(
					"unexpected error\nexpected: %v\nactual:   %v\n",
					test.expectedError,
					err,
				)
			}

			if valueDest.Cmp(test.expectedValue) != 0 {
				t.Errorf(
					"\nexpected: %s\nactual:   %s",
					test.expectedValue,
					valueDest,
				)
			}
		})
	}
}

func TestWeiVarFlag_DefaultValue(t *testing.T) {
	defaultValue := *ethereum.WrapWei(big.NewInt(859756))
	flagName := "amount"

	flags := pflag.NewFlagSet("flag-set", pflag.PanicOnError)

	var valueDest ethereum.Wei

	WeiVarFlag(flags, &valueDest, flagName, defaultValue, "")

	if valueDest.Cmp(defaultValue.Int) != 0 {
		t.Errorf(
			"\nexpected: %s\nactual:   %s",
			defaultValue,
			valueDest,
		)
	}

}
