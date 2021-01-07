package main

// contractTemplateContent contains the template string from contract.go.tmpl
var contractTemplateContent = `// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	ethereumabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/subscription"
)

// Create a package-level logger for this contract. The logger exists at
// package level so that the logger is registered at startup and can be
// included or excluded from logging at startup by name.
var {{.ShortVar}}Logger = log.Logger("keep-contract-{{.Class}}")

const (
    // Maximum backoff time between event resubscription attempts.
    {{.ShortVar}}SubscriptionBackoffMax = 2 * time.Minute

    // Threshold below which event resubscription emits an error to the logs.
    // WS connection can be dropped at any moment and event resubscription will
    // follow. However, if WS connection for event subscription is getting
    // dropped too often, it may indicate something is wrong with Ethereum
    // client. This constant defines the minimum lifetime of an event
    // subscription required before the subscription failure happens and
    // resubscription follows so that the resubscription does not emit an error
    // to the logs alerting about potential problems with Ethereum client.
    {{.ShortVar}}SubscriptionAlertThreshold = 5 * time.Minute
)

type {{.Class}} struct {
	contract           *abi.{{.AbiClass}}
	contractAddress    common.Address
	contractABI        *ethereumabi.ABI
	caller             bind.ContractCaller
	transactor         bind.ContractTransactor
	callerOptions      *bind.CallOpts
	transactorOptions  *bind.TransactOpts
	errorResolver      *ethutil.ErrorResolver
	nonceManager       *ethutil.NonceManager
	miningWaiter       *ethutil.MiningWaiter

	transactionMutex *sync.Mutex
}

func New{{.Class}}(
    contractAddress common.Address,
    accountKey *keystore.Key,
    backend bind.ContractBackend,
    nonceManager *ethutil.NonceManager,
    miningWaiter *ethutil.MiningWaiter,
    transactionMutex *sync.Mutex,
) (*{{.Class}}, error) {
	callerOptions := &bind.CallOpts{
		From: accountKey.Address,
	}

	transactorOptions := bind.NewKeyedTransactor(
		accountKey.PrivateKey,
	)

	randomBeaconContract, err := abi.New{{.AbiClass}}(
		contractAddress,
		backend,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to instantiate contract at address: %s [%v]",
			contractAddress.String(),
			err,
		)
	}

	contractABI, err := ethereumabi.JSON(strings.NewReader(abi.{{.AbiClass}}ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ABI: [%v]", err)
	}

	return &{{.Class}}{
		contract:          randomBeaconContract,
		contractAddress:   contractAddress,
		contractABI: 	   &contractABI,
		caller:     	   backend,
		transactor:        backend,
		callerOptions:     callerOptions,
		transactorOptions: transactorOptions,
		errorResolver:     ethutil.NewErrorResolver(backend, &contractABI, &contractAddress),
		nonceManager:      nonceManager,
		miningWaiter:      miningWaiter,
		transactionMutex:  transactionMutex,
	}, nil
}

// ----- Non-const Methods ------
{{template "contract_non_const_methods.go.tmpl" .}}

// ----- Const Methods ------
{{template "contract_const_methods.go.tmpl" .}}

// ------ Events -------
{{template "contract_events.go.tmpl" . -}}
`
