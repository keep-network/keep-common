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

// ChainID returns chain id associated with the network.
func (n Network) ChainID() int64 {
	switch n {
	case Mainnet:
		return 1
	case Goerli:
		return 5
	}
	return 0
}
