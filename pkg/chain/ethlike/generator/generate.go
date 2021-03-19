package main

import (
	"flag"
	"fmt"
	"os"
	"text/template"
)

const (
	expectedArgs         = 2
	templateFileArgIndex = 0
	goFileArgIndex       = 1
	goExtension          = ".go"
)

type context struct {
	HostChainModule  string
	ChainUtilPackage string
}

func main() {
	hostChainModule := flag.String(
		"host-chain-module",
		"github.com/ethereum/go-ethereum",
		"ETH-like host chain Go module imported from the generated code",
	)

	chainUtilPackage := flag.String(
		"chain-util-package",
		"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil",
		"Host chain utils package imported from the generated code",
	)

	flag.Parse()

	if flag.NArg() != expectedArgs {
		errorAndExit(fmt.Sprintf(
			"expected `%v <template-file> <go-file>`, but got [%v]",
			os.Args[0],
			os.Args,
		))
	}

	templateFilePath := flag.Arg(templateFileArgIndex)
	goFilePath := flag.Arg(goFileArgIndex)

	if goFilePath[len(goFilePath)-len(goExtension):] != goExtension {
		errorAndExit("Go file should end in .go")
	}

	context := context{
		HostChainModule:  *hostChainModule,
		ChainUtilPackage: *chainUtilPackage,
	}

	tmpl, err := template.ParseFiles(templateFilePath)
	if err != nil {
		errorAndExit(fmt.Sprintf(
			"could not parse template [%v]: [%v]",
			templateFilePath,
			err,
		))
	}

	goFile, err := os.Create(goFilePath)
	if err != nil {
		errorAndExit(fmt.Sprintf(
			"could not create Go file for template [%v]: [%v]",
			goFilePath,
			err,
		))
	}

	err = tmpl.Execute(goFile, context)
	if err != nil {
		errorAndExit(fmt.Sprintf(
			"could not execute template [%v]: [%v]",
			templateFilePath,
			err,
		))
	}

	err = goFile.Close()
	if err != nil {
		errorAndExit(fmt.Sprintf(
			"could not close Go file [%v]: [%v]",
			goFile.Name(),
			err,
		))
	}
}

func errorAndExit(error string) {
	_, _ = fmt.Fprintf(os.Stderr, error+"\n")
	os.Exit(1)
}
