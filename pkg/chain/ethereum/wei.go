package ethereum

import (
	"github.com/ethereum/go-ethereum/params"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
)

// Units defines denominations of the Ether token.
var Units = map[string]float64{
	"wei":   params.Wei, // default unit
	"gwei":  params.GWei,
	"ether": params.Ether,
}

// Wei is a custom type to handle Ether value parsing in configuration files
// using BurntSushi/toml package. It supports wei, Gwei and ether units. The
// Ether value is kept as `wei` and `wei` is the default unit.
// The value can be provided in the text file as e.g.: `1 wei`, `200 Gwei` or
// `0.5 ether`.
type Wei struct {
	ethlike.Token
}

// UnmarshalText is a function used to parse a value of Ethers.
func (w *Wei) UnmarshalText(text []byte) error {
	return w.UnmarshalToken(text, Units)
}