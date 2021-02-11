package ethlike

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

var Units = map[string]float64{
	"wei":   1, // default unit
	"gwei":  1e9,
	"ether": 1e18,
	"celo":  1e18,
}

type Value struct {
	*big.Int
}

// UnmarshalText is a function used to parse an ETH-like value.
func (v *Value) UnmarshalText(text []byte) error {
	re := regexp.MustCompile(`^(\d+[\.]?[\d]*)[ ]?([\w]*)$`)
	matched := re.FindSubmatch(text)

	if len(matched) != 3 {
		return fmt.Errorf("failed to parse value: [%s]", text)
	}

	number, ok := new(big.Float).SetString(string(matched[1]))
	if !ok {
		return fmt.Errorf(
			"failed to set float value from string [%s]",
			string(matched[1]),
		)
	}

	unit := matched[2]
	if len(unit) == 0 {
		unit = []byte("wei")
	}

	if factor, ok := Units[strings.ToLower(string(unit))]; ok {
		number.Mul(number, big.NewFloat(factor))
		v.Int, _ = number.Int(nil)
		return nil
	}

	return fmt.Errorf(
		"invalid unit: %s; please use one of: wei, Gwei, ether, CELO",
		unit,
	)
}
