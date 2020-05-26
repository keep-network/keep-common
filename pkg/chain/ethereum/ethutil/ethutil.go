// Package ethutil provides utilities used for dealing with Ethereum concerns in
// the context of implementing cross-chain interfaces defined in pkg/chain.
package ethutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/ipfs/go-log/v2"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

var logger = log.Logger("keep-ethutil")

// AddressFromHex converts the passed string to a common.Address and returns it,
// unless it is not a valid address, in which case it returns an error. Compare
// to common.HexToAddress, which assumes the address is valid and does not
// provide for an error return.
func AddressFromHex(hex string) (common.Address, error) {
	if common.IsHexAddress(hex) {
		return common.HexToAddress(hex), nil
	}

	return common.Address{}, fmt.Errorf("[%v] is not a valid Ethereum address", hex)
}

// DecryptKeyFile reads in a key file and uses the password to decrypt it.
func DecryptKeyFile(keyFile, password string) (*keystore.Key, error) {
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
// standard, WebSocket, and RPC clients for the Ethereum node at that address.
func ConnectClients(url string, urlRPC string) (*ethclient.Client, *rpc.Client, *rpc.Client, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(
			"error Connecting to Geth Server: %s [%v]",
			url,
			err,
		)
	}

	clientWS, err := rpc.Dial(url)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(
			"error Connecting to Geth Server: %s [%v]",
			url,
			err,
		)
	}

	clientRPC, err := rpc.Dial(urlRPC)
	if err != nil {
		return nil, nil, nil, fmt.Errorf(
			"error Connecting to Geth Server: %s [%v]",
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
		msg = ethereum.CallMsg{
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
		if code, err = caller.CodeAt(context.TODO(), contractAddress, nil); err != nil {
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

// EstimateGas tries to estimate the gas needed to execute a specific transaction based on
// the current pending state of the backend blockchain. There is no guarantee that this is
// the true gas limit requirement as other transactions may be added or removed by miners,
// but it should provide a basis for setting a reasonable default.
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

	msg := ethereum.CallMsg{
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

type loggingWrapper struct {
	bind.ContractBackend

	logger log.EventLogger
}

func (lw *loggingWrapper) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	price, err := lw.ContractBackend.SuggestGasPrice(ctx)

	if err != nil {
		lw.logger.Debugf("error requesting gas price suggestion: [%v]", err)
		return nil, err
	}

	lw.logger.Debugf("received gas price suggestion: [%v]", price)
	return price, nil
}

func (lw *loggingWrapper) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	gas, err := lw.ContractBackend.EstimateGas(ctx, msg)

	if err != nil {
		return 0, err
	}

	lw.logger.Debugf("received gas estimate: [%v]", gas)
	return gas, nil
}

// WrapCallLogging wraps certain call-related methods on the given `backend`
// with debug logging sent to the given `logger`. Actual functionality is
// delegated to the passed backend.
func WrapCallLogging(logger log.EventLogger, backend bind.ContractBackend) bind.ContractBackend {
	return &loggingWrapper{backend, logger}
}
