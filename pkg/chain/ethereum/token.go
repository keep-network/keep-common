package ethereum

import (
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strings"
)

// Token represents a token.
type Token struct {
	*big.Int
}

// UnmarshalToken is a function used to parse an Ethereum token.
func (t *Token) UnmarshalToken(text []byte, units map[string]float64) error {
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
		for unitName, factor := range units {
			// set the unit to the default one
			if factor == 1 {
				unit = []byte(unitName)
				break
			}
		}

		if len(unit) == 0 {
			return fmt.Errorf("could not determine default unit")
		}
	}

	if factor, ok := units[strings.ToLower(string(unit))]; ok {
		number.Mul(number, big.NewFloat(factor))
		t.Int, _ = number.Int(nil)
		return nil
	}

	unitNames := make([]string, 0)
	for unitName := range units {
		unitNames = append(unitNames, unitName)
	}

	sort.Strings(unitNames)

	return fmt.Errorf(
		"invalid unit: %s; please use one of: %v",
		unit,
		strings.Join(unitNames, ", "),
	)
}
