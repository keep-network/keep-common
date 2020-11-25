package ethereum

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/params"
)

// Wei is a custom type to handle Ether value parsing in configuration files
// using BurntSushi/toml package. It supports wei, Gwei and ether units. The
// Ether value is kept as `wei` and `wei` is the default unit.
// The value can be provided in the text file as e.g.: `1 wei`, `200 Gwei` or
// `0.5 ether`.
type Wei struct {
	*big.Int
}

// The most common units for ether values.
const (
	wei unit = iota
	gwei
	ether
)

// unit represents Ether value unit.
type unit int

func (u unit) String() string {
	return [...]string{"wei", "Gwei", "ether"}[u]
}

// UnmarshalText is a function used to parse a value of Ethers.
func (e *Wei) UnmarshalText(text []byte) error {
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

	switch strings.ToLower(string(unit)) {
	case strings.ToLower(ether.String()):
		number.Mul(number, big.NewFloat(params.Ether))
		e.Int, _ = number.Int(nil)
	case strings.ToLower(gwei.String()):
		number.Mul(number, big.NewFloat(params.GWei))
		e.Int, _ = number.Int(nil)
	case strings.ToLower(wei.String()):
		number.Mul(number, big.NewFloat(params.Wei))
		e.Int, _ = number.Int(nil)
	default:
		return fmt.Errorf(
			"invalid unit: %s; please use one of: wei, Gwei, ether",
			unit,
		)
	}

	return nil
}
