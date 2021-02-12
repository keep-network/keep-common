package main

// contractTemplateContent contains the template string from contract.go.tmpl
var contractTemplateContent = `// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"strings"
	"sync"

	backendabi "{{.BackendModule}}/accounts/abi"
	"{{.BackendModule}}/accounts/abi/bind"
	"{{.BackendModule}}/accounts/keystore"
	"{{.BackendModule}}/common"
	"{{.BackendModule}}/core/types"
	"{{.BackendModule}}/event"

	"github.com/ipfs/go-log"

	chainutil "{{.ChainUtilPackage}}"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"github.com/keep-network/keep-common/pkg/subscription"
)

// Create a package-level logger for this contract. The logger exists at
// package level so that the logger is registered at startup and can be
// included or excluded from logging at startup by name.
var {{.ShortVar}}Logger = log.Logger("keep-contract-{{.Class}}")

type {{.Class}} struct {
	contract           *abi.{{.AbiClass}}
	contractAddress    common.Address
	contractABI        *backendabi.ABI
	caller             bind.ContractCaller
	transactor         bind.ContractTransactor
	callerOptions      *bind.CallOpts
	transactorOptions  *bind.TransactOpts
	errorResolver      *chainutil.ErrorResolver
	nonceManager       *ethlike.NonceManager
	miningWaiter       *ethlike.MiningWaiter
	blockCounter	   *ethlike.BlockCounter

	transactionMutex *sync.Mutex
}

func New{{.Class}}(
    contractAddress common.Address,
    accountKey *keystore.Key,
    backend bind.ContractBackend,
    nonceManager *ethlike.NonceManager,
    miningWaiter *ethlike.MiningWaiter,
    blockCounter *ethlike.BlockCounter,
    transactionMutex *sync.Mutex,
) (*{{.Class}}, error) {
	callerOptions := &bind.CallOpts{
		From: accountKey.Address,
	}

	transactorOptions := bind.NewKeyedTransactor(
		accountKey.PrivateKey,
	)

	contract, err := abi.New{{.AbiClass}}(
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

	contractABI, err := backendabi.JSON(strings.NewReader(abi.{{.AbiClass}}ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ABI: [%v]", err)
	}

	return &{{.Class}}{
		contract:          contract,
		contractAddress:   contractAddress,
		contractABI: 	   &contractABI,
		caller:     	   backend,
		transactor:        backend,
		callerOptions:     callerOptions,
		transactorOptions: transactorOptions,
		errorResolver:     chainutil.NewErrorResolver(backend, &contractABI, &contractAddress),
		nonceManager:      nonceManager,
		miningWaiter:      miningWaiter,
		blockCounter: 	   blockCounter,
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