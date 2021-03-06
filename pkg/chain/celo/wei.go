package celo

import (
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
)

// Units defines denominations of the CELO token.
var Units = map[string]float64{
	"wei":  1, // default unit
	"gwei": 1e9,
	"celo": 1e18,
}

// Wei is a custom type to handle CELO value parsing in configuration files
// using BurntSushi/toml package. It supports wei, Gwei and CELO units. The
// CELO value is kept as `wei` and `wei` is the default unit.
// The value can be provided in the text file as e.g.: `1 wei`, `200 Gwei` or
// `0.5 CELO`.
type Wei struct {
	ethlike.Token
}

// WrapWei wraps the given integer value in order to represent it as Wei value.
func WrapWei(value *big.Int) *Wei {
	return &Wei{ethlike.Token{value}}
}

// UnmarshalText is a function used to parse a value of CELO.
func (w *Wei) UnmarshalText(text []byte) error {
	return w.UnmarshalToken(text, Units)
}
