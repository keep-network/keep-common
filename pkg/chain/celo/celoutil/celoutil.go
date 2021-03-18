package celoutil

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/celo-org/celo-blockchain"
	"github.com/celo-org/celo-blockchain/accounts/abi"
	"github.com/celo-org/celo-blockchain/accounts/abi/bind"
	"github.com/celo-org/celo-blockchain/accounts/keystore"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/celo-org/celo-blockchain/crypto"
	celoclient "github.com/celo-org/celo-blockchain/ethclient"
	"github.com/celo-org/celo-blockchain/rpc"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
)

var logger = log.Logger("keep-celoutil")

// CeloClient wraps the core `bind.ContractBackend` interface with
// some other interfaces allowing to expose additional methods provided
// by client implementations.
type CeloClient interface {
	bind.ContractBackend
	celo.ChainReader
	celo.TransactionReader

	BalanceAt(
		ctx context.Context,
		account common.Address,
		blockNumber *big.Int,
	) (*big.Int, error)
}

// AddressFromHex converts the passed string to a common.Address and returns it,
// unless it is not a valid address, in which case it returns an error. Compare
// to common.HexToAddress, which assumes the address is valid and does not
// provide for an error return.
func AddressFromHex(hex string) (common.Address, error) {
	if common.IsHexAddress(hex) {
		return common.HexToAddress(hex), nil
	}

	return common.Address{}, fmt.Errorf(
		"[%v] is not a valid Celo address",
		hex,
	)
}

// DecryptKeyFile reads in a key file and uses the password to decrypt it.
func DecryptKeyFile(keyFile, password string) (*keystore.Key, error) {
	// #nosec G304 (file path provided as taint input)
	// This line is used to read a local key file. There is no user input.
	data, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read KeyFile %s [%v]", keyFile, err)
	}
	key, err := keystore.DecryptKey(data, password)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt %s [%v]", keyFile, err)
	}
	return key, nil
}

// ConnectClients takes HTTP and RPC URLs and returns initialized versions of
// standard, WebSocket, and RPC clients for the Celo node at that address.
func ConnectClients(url string, urlRPC string) (
	*celoclient.Client,
	*rpc.Client,
	*rpc.Client,
	error,
) {
	client, err := celoclient.Dial(url)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(
			"error Connecting to Celo node: %s [%v]",
			url,
			err,
		)
	}

	clientWS, err := rpc.Dial(url)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(
			"error Connecting to Celo node: %s [%v]",
			url,
			err,
		)
	}

	clientRPC, err := rpc.Dial(urlRPC)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(
			"error Connecting to Celo node: %s [%v]",
			url,
			err,
		)
	}

	return client, clientWS, clientRPC, nil
}

// CallAtBlock allows the invocation of a particular contract method at a
// particular block. It papers over the fact that abigen bindings don't directly
// support calling at a particular block, and is mostly meant for use from
// generated contract code.
func CallAtBlock(
	fromAddress common.Address,
	blockNumber *big.Int,
	value *big.Int,
	contractABI *abi.ABI,
	caller bind.ContractCaller,
	errorResolver *ErrorResolver,
	contractAddress common.Address,
	method string,
	result interface{},
	parameters ...interface{},
) error {
	input, err := contractABI.Pack(method, parameters...)
	if err != nil {
		return err
	}

	var (
		msg = celo.CallMsg{
			From:  fromAddress,
			To:    &contractAddress,
			Data:  input,
			Value: value,
		}
		code   []byte
		output []byte
	)

	output, err = caller.CallContract(context.TODO(), msg, blockNumber)
	if err == nil && len(output) == 0 {
		// Make sure we have a contract to operate on, and bail out otherwise.
		if code, err = caller.CodeAt(
			context.TODO(),
			contractAddress,
			nil,
		); err != nil {
			return err
		} else if len(code) == 0 {
			return bind.ErrNoCode
		}
	}

	err = contractABI.Unpack(result, method, output)

	if err != nil {
		return errorResolver.ResolveError(
			err,
			fromAddress,
			value,
			method,
			parameters...,
		)
	}

	return nil
}

// EstimateGas tries to estimate the gas needed to execute a specific
// transaction based on the current pending state of the backend blockchain.
// There is no guarantee that this is the true gas limit requirement as other
// transactions may be added or removed by miners, but it should provide a
// basis for setting a reasonable default.
func EstimateGas(
	from common.Address,
	to common.Address,
	method string,
	contractABI *abi.ABI,
	transactor bind.ContractTransactor,
	parameters ...interface{},
) (uint64, error) {
	input, err := contractABI.Pack(method, parameters...)
	if err != nil {
		return 0, err
	}

	msg := celo.CallMsg{
		From: from,
		To:   &to,
		Data: input,
	}

	gas, err := transactor.EstimateGas(context.TODO(), msg)
	if err != nil {
		return 0, err
	}

	return gas, nil
}

// NewBlockCounter creates a new BlockCounter instance for the provided
// Celo client.
func NewBlockCounter(client CeloClient) (*ethlike.BlockCounter, error) {
	return ethlike.CreateBlockCounter(&ethlikeAdapter{client})
}

// NewMiningWaiter creates a new MiningWaiter instance for the provided
// Celo client. It accepts two parameters setting up monitoring rules
// of the transaction mining status.
func NewMiningWaiter(
	client CeloClient,
	checkInterval time.Duration,
	maxGasPrice *big.Int,
) *ethlike.MiningWaiter {
	return ethlike.NewMiningWaiter(
		&ethlikeAdapter{client},
		checkInterval,
		maxGasPrice,
	)
}

// NewNonceManager creates NonceManager instance for the provided account
// using the provided Celo client.
func NewNonceManager(
	client CeloClient,
	account common.Address,
) *ethlike.NonceManager {
	return ethlike.NewNonceManager(
		&ethlikeAdapter{client},
		ethlike.Address(account),
	)
}

// NewKeyedTransactorWithChainID is a utility method to easily create
// a transaction signer from a single private key.
// FIXME Remove this function and rely on bind.NewKeyedTransactorWithChainID
// FIXME when celo-org/celo-blockchain merges in changes from upstream
// FIXME ethereum/go-ethereum beyond v1.9.25.
func NewKeyedTransactorWithChainID(
	key *ecdsa.PrivateKey,
	chainID *big.Int,
) (*bind.TransactOpts, error) {
	keyAddress := crypto.PubkeyToAddress(key.PublicKey)
	if chainID == nil {
		return nil, fmt.Errorf("no chain id specified")
	}
	return &bind.TransactOpts{
		From: keyAddress,
		Signer: func(
			_ types.Signer,
			address common.Address,
			tx *types.Transaction,
		) (*types.Transaction, error) {
			signer := types.NewEIP155Signer(chainID)

			if address != keyAddress {
				return nil, fmt.Errorf("not authorized to sign this account")
			}
			signature, err := crypto.Sign(signer.Hash(tx).Bytes(), key)
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
	}, nil
}
