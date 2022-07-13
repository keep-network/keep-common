package ethlike

// Account is a struct that contains the configuration for accessing an
// ETH-like network and a contract on the network.
type Account struct {
	// Keyfile is a full path to a key file.  Normally this file is one of the
	// imported keys in your local chain node.  It can normally be found in
	// a directory <some-path>/data/keystore/ and starts with its creation date
	// "UTC--.*".
	KeyFile string

	// KeyFilePassword is the password used to unlock the account specified in
	// KeyFile.
	KeyFilePassword string
}

// Config is a struct that contains the configuration needed to connect to an
// ETH-like node. This information will give access to an ETH-like network.
type Config struct {
	// Example: "ws://192.168.0.157:8546".
	URL string

	// Example: "http://192.168.0.157:8545".
	URLRPC string

	// A map from contract names to contract addresses. The keys in the map are
	// expected to be lowercase contract names.
	ContractAddresses map[string]string

	Account Account

	// MiningCheckInterval is the interval in which transaction
	// mining status is checked. If the transaction is not mined within this
	// time, the gas price is increased and transaction is resubmitted.
	MiningCheckInterval int

	// RequestsPerSecondLimit sets the maximum average number of requests
	// per second which can be executed against the ETH-like node.
	// All types of chain requests are rate-limited,
	// including view function calls.
	RequestsPerSecondLimit int

	// ConcurrencyLimit sets the maximum number of concurrent requests which
	// can be executed against the ETH-like node at the same time.
	// This limit affects all types of chain requests,
	// including view function calls.
	ConcurrencyLimit int
}
