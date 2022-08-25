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
	switch n {
	case Mainnet:
		return "mainnet"
	case Goerli:
		return "goerli"
	case Developer:
		return "developer"
	}
	return "unknown"
}
