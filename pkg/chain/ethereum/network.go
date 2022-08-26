package ethereum

// Network is a type used for ethereum networks enumeration.
type Network int

// Ethereum networks enumeration.
const (
	Unknown Network = iota
	Mainnet
	Goerli
	Developer
)

func (n Network) String() string {
	return []string{"unknown", "mainnet", "goerli", "developer"}[n]
}
