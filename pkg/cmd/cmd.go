// Package cmd contains useful utilities for commands that interact with the
// Ethereum blockchain.
package cmd

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/cmd/flag"
	"github.com/urfave/cli"
)

const (
	blockFlag        string = "block"
	blockShort       string = "b"
	transactionFlag  string = "transaction"
	transactionShort string = "t"
	// SubmitFlag allows for urfave/cli definition and lookup of a boolean
	// `--submit` command-line flag indicating that a given contract interaction
	// should be submitted as a paid, mutating interaction to the configured
	// Ethereum chain.
	SubmitFlag  string = "submit"
	submitShort string = "s"
	valueFlag   string = "value"
	valueShort  string = "v"
)

var (
	// TransactionFlagValue allows for reading the transaction hash flag
	// included in ConstFlags, which represents a transaction hash from which to
	// retrieve an already-executed contract interaction. The value, if that
	// flag is passed on the command line, is stored in this variable.
	TransactionFlagValue *flag.TransactionHash = &flag.TransactionHash{}
	// BlockFlagValue allows for reading the block flag included in ConstFlags,
	// which represents the block at which to execute a contract interaction.
	// The value, if that flag is passed on the command line, is stored in this
	// variable.
	BlockFlagValue *flag.Uint256 = &flag.Uint256{}
	// ValueFlagValue allows for reading the value flag included in
	// PayableFlags, which represents an amount of ETH to send with a contract
	// interaction. The value, if that flag is passed on the command line, is
	// stored in this variable.
	ValueFlagValue *flag.Uint256 = &flag.Uint256{}
)

// AvailableCommands is the exported list of generated commands that can be
// installed on a CLI app. Generated contract command files set up init
// functions that add the contract's command and subcommands to this global
// variable, and any top-level command that wishes to include these commands can
// reference this variable and expect it to contain all generated contract
// commands.
var AvailableCommands []cli.Command

var (
	// ConstFlags provides a slice of flags useful for constant contract
	// interactions, meaning contract interactions that do not require
	// transaction submission and are used for inspecting chain state. These
	// flags include the --block flag to check an interaction's result value at
	// a specific block and the --transaction flag to check an interaction's
	// already-evaluated result value from a given transaction.
	ConstFlags = []cli.Flag{
		&cli.GenericFlag{
			Name:  blockFlag + ", " + blockShort,
			Usage: "Retrieve the result of calling this method on `BLOCK`.",
			Value: BlockFlagValue,
		},
		&cli.GenericFlag{
			Name:  transactionFlag + ", " + transactionShort,
			Usage: "Retrieve the already-evaluated result of this method in `TRANSACTION`.",
			Value: TransactionFlagValue,
		},
	}
	// NonConstFlags provides a slice of flags useful for non-constant contract
	// interactions, meaning contract interactions that can be submitted as
	// transactions and are used for modifying chain state. These flags include
	// the --submit flag to submit an interaction as a transaction, as well as
	// all flags in ConstFlags.
	NonConstFlags = append(
		[]cli.Flag{
			&cli.BoolFlag{
				Name:  SubmitFlag + ", " + submitShort,
				Usage: "Submit this call as a gas-spending network transaction.",
			},
		},
		ConstFlags...,
	)
	// PayableFlags provides a slice of flags useful for payable contract
	// interactions, meaning contract interactions that can be submitted as
	// transactions and are used for modifying chain state with a payload that
	// includes ETH. These flags include the --value flag to specify the ETH
	// amount to send with the interaction, as well as all flags in
	// NonConstFlags.
	PayableFlags = append(
		[]cli.Flag{
			&cli.GenericFlag{
				Name:  valueFlag + ", " + valueShort,
				Usage: "Send `VALUE` ether with this call.",
				Value: ValueFlagValue,
			},
		},
		NonConstFlags...,
	)
)

// ComposableArgChecker is a type that allows multiple urfave/cli BeforeFuncs to
// be chained. See AndThen for more.
type ComposableArgChecker cli.BeforeFunc

// AndThen on a ComposableArgChecker allows composing a ComposableArgChecker
// with another one, such that this ComposableArgChecker runs and, if it
// succeeds, the nextChecker runs.
//
// As an example, this allows for two BeforeFuncs to be composed as:
//
//   ComposableArgChecker(checkFlagAValue).AndThen(ComposableArgChecker(checkFlagBValue))
//
// The resulting ComposableArgChecker will run checkFlagAValue and, if it
// passes, checkFlagBValue.
func (cac ComposableArgChecker) AndThen(nextChecker ComposableArgChecker) ComposableArgChecker {
	return func(c *cli.Context) error {
		cacErr := cac(c)
		if cacErr != nil {
			return cacErr
		}

		return nextChecker(c)
	}
}

var (
	valueArgChecker ComposableArgChecker = func(c *cli.Context) error {
		if !c.IsSet(valueFlag) {
			return fmt.Errorf("expected value for this payable method")
		}

		return nil
	}
	submittedArgChecker ComposableArgChecker = func(c *cli.Context) error {
		if c.Bool(SubmitFlag) {
			if c.IsSet(blockFlag) {
				return fmt.Errorf("cannot specify --block for a submitted transaction")
			}
			if c.IsSet(transactionFlag) {
				return fmt.Errorf("cannot specify --transaction for a submitted transaction")
			}
		}

		return nil
	}
)

var (
	// NonConstArgsChecker runs any validation of parameters needed for a
	// command that will submit a mutating transaction to an Ethereum node. In
	// particular, it will disallow const arguments like --block and
	// --transaction if --submit is specified.
	NonConstArgsChecker = submittedArgChecker
	// PayableArgsChecker runs any validation of parameters needed for a
	// command that will submit a payable transaction to an Ethereum node. In
	// particular, it will disallow const arguments like --block and
	// --transaction if --submit is specified, and will also ensure that the
	// call includes specifies a --value.
	PayableArgsChecker = NonConstArgsChecker.AndThen(valueArgChecker)
)

// ArgCountChecker provides a consistent error in case the number of arguments
// passed on the command-line is incorrect.
func ArgCountChecker(expectedArgCount int) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if c.NArg() != expectedArgCount {
			return fmt.Errorf(
				"Expected [%v] arguments but got [%v]",
				expectedArgCount,
				c.NArg(),
			)
		}

		return nil
	}
}

// PrintOutput provides for custom command-line-friendly ways of printing
// addresses (as hex) and transaction and related hashes (also as hex). It falls
// back on a standard Println for other types.
func PrintOutput(output interface{}) {
	switch out := output.(type) {
	case common.Address:
		fmt.Println(out.Hex())
	case common.Hash:
		fmt.Println(out.Hex())
	default:
		fmt.Println(output)
	}
}
