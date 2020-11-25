package ethereum

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/params"
)

// Weis is a custom type to handle Ether value parsing in configuration files
// using BurntSushi/toml package. It supports wei, Gwei and ether units. The
// Ether value is kept as `wei` and `wei` is the default unit.
// The value can be provided in the text file as e.g.: `1 wei`, `200 Gwei` or
// `0.5 ether`.
type Weis struct {
	*big.Int
}

// The most common units for Ether values.
const (
	Wei Unit = iota
	Gwei
	Ether
)

// Unit represents Ether value unit.
type Unit int

func (u Unit) String() string {
	return [...]string{"wei", "Gwei", "ether"}[u]
}

// UnmarshalText is a function used to parse a value of Ethers.
func (e *Weis) UnmarshalText(text []byte) error {
	re := regexp.MustCompile(`^(\d+[\.]?[\d]*)[ ]?([\w]*)$`)
	matched := re.FindSubmatch(text)

	if len(matched) < 1 {
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
	case strings.ToLower(Ether.String()):
		number.Mul(number, big.NewFloat(params.Ether))
		e.Int, _ = number.Int(nil)
	case strings.ToLower(Gwei.String()):
		number.Mul(number, big.NewFloat(params.GWei))
		e.Int, _ = number.Int(nil)
	case strings.ToLower(Wei.String()):
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
