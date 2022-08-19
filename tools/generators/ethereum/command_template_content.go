package main

// commandTemplateContent contains the template string from command.go.tmpl
var commandTemplateContent = `// Code generated - DO NOT EDIT.
// This file is a generated command and any manual changes will be lost.

package cmd

import (
    "sync"

    "{{.HostChainModule}}/common"
    "{{.HostChainModule}}/common/hexutil"
    "{{.HostChainModule}}/core/types"
    "{{.HostChainModule}}/ethclient"

    chainutil "{{.ChainUtilPackage}}"
    "github.com/keep-network/keep-common/pkg/cmd"

    "github.com/spf13/cobra"
)

var {{.Class}}Command *cobra.Command

var {{.FullVar}}Description = ` + "`" + `The {{.DashedName}} command allows calling the {{.Class}} contract on an
	Ethereum network. It has subcommands corresponding to each contract method,
	which respectively each take parameters based on the contract method's
	parameters.

	Subcommands will submit a non-mutating call to the network and output the
	result.

	All subcommands can be called against a specific block by passing the
	-b/--block flag.

	All subcommands can be used to investigate the result of a previous
	transaction that called that same method by passing the -t/--transaction
	flag with the transaction hash.

	Subcommands for mutating methods may be submitted as a mutating transaction
	by passing the -s/--submit flag. In this mode, this command will terminate
	successfully once the transaction has been submitted, but will not wait for
	the transaction to be included in a block. They return the transaction hash.

	Calls that require ether to be paid will get 0 ether by default, which can
	be changed by passing the -v/--value flag.` + "`" + `

func init() {
    {{.Class}}Command := &cobra.Command{
        Use:        "{{.DashedName}}",
        Short:       ` + "`" + `Provides access to the {{.Class}} contract.` + "`" + `,
        Long: {{.FullVar}}Description,
    }

    {{.Class}}Command.AddCommand(
        {{- $contract := . -}}
        {{- range $i, $method := .ConstMethods }}
        {{- if $method.CommandCallable }}
            {{$contract.ShortVar}}{{$method.CapsName}}Command(),
        {{- end -}}
        {{- end -}}
        {{- range $i, $method := .NonConstMethods }}
        {{- if $method.CommandCallable }}
            {{$contract.ShortVar}}{{$method.CapsName}}Command(),
        {{- end -}}
        {{- end }}
    )

    Command.AddCommand({{.Class}}Command)
}

/// ------------------- Const methods -------------------

{{- $contract := . -}}
{{- range $i, $method := .ConstMethods -}}
{{- if $method.CommandCallable }}

func {{$contract.ShortVar}}{{$method.CapsName}}Command() *cobra.Command {
		c := &cobra.Command{
				Use: "{{$method.DashedName}}{{ range $i, $param := $method.ParamInfos }} [{{$param.Name}}]{{ end }}",
				Short: "Calls the {{$method.Modifiers -}} method {{$method.LowerName}} on the {{$contract.Class}} contract.",
                Args: cmd.ArgCountChecker({{$method.ParamInfos | len}}),
				RunE: {{$contract.ShortVar}}{{$method.CapsName}},
                SilenceUsage: true,
				DisableFlagsInUseLine: true,
			}

		cmd.InitConstFlags(c)

		return c
}

func {{$contract.ShortVar}}{{$method.CapsName}}(c *cobra.Command, args []string) error {
    contract, err := initialize{{$contract.Class}}(c)
    if err != nil {
        return err
    }

   	{{- range $i, $param := .ParamInfos }}
   	{{$param.Name}}, err := {{$param.ParsingFn}}(args[{{$i}}])
   	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter {{$param.Name}}, a {{$param.Type}}, from passed value %v",
			args[{{$i}}],
		)
   	}
   	{{ end }}

    result, err := contract.{{$method.CapsName}}AtBlock(
        {{ $method.Params -}}
        cmd.BlockFlagValue.Int,
    )

    if err != nil {
    	return err
    }

    cmd.PrintOutput(result)

    return nil
}

{{- end -}}
{{- end }}

/// ------------------- Non-const methods -------------------

{{- range $i, $method := .NonConstMethods -}}
{{- if $method.CommandCallable }}

func {{$contract.ShortVar}}{{$method.CapsName}}Command() *cobra.Command {
		c := &cobra.Command{
				Use: "{{$method.DashedName}}{{ range $i, $param := $method.ParamInfos }} [{{$param.Name}}]{{ end }}",
				Short: "Calls the {{$method.Modifiers -}} method {{$method.LowerName}} on the {{$contract.Class}} contract.",
                Args: cmd.ArgCountChecker({{$method.ParamInfos | len}}),
				RunE: {{$contract.ShortVar}}{{$method.CapsName}},
                SilenceUsage: true,
				DisableFlagsInUseLine: true,
			}

        {{if $method.Payable -}}
        c.PreRunE = cmd.PayableArgsChecker
        cmd.InitPayableFlags(c)
        {{- else -}}
        c.PreRunE = cmd.NonConstArgsChecker
        cmd.InitNonConstFlags(c)
        {{- end }}

		return c
}

func {{$contract.ShortVar}}{{$method.CapsName}}(c *cobra.Command, args []string) error {
    contract, err := initialize{{$contract.Class}}(c)
    if err != nil {
        return err
    }

    {{ range $i, $param := .ParamInfos }}
    {{$param.Name}}, err := {{$param.ParsingFn}}(args[{{$i}}])
    if err != nil {
        return fmt.Errorf(
            "couldn't parse parameter {{$param.Name}}, a {{$param.Type}}, from passed value %v",
            args[{{$i}}],
        )
    }

    {{ end -}}

    var (
        transaction *types.Transaction
        {{ if gt (len $method.Return.Type) 0 -}}
        result {{$method.Return.Type}}
        {{ end -}}
    )

    if shouldSubmit, _ := c.Flags().GetBool(cmd.SubmitFlag); shouldSubmit {
        // Do a regular submission. Take payable into account.
        transaction, err = contract.{{$method.CapsName}}(
            {{$method.Params}}
            {{- if $method.Payable -}} cmd.ValueFlagValue.Int(), {{- end -}}
        )
        if err != nil {
            return err
        }

        cmd.PrintOutput(transaction.Hash())
    } else {
        // Do a call.
        {{ if gt (len $method.Return.Type) 0 -}} result, {{ end -}} err = contract.Call{{$method.CapsName}}(
            {{$method.Params}}
            {{- if $method.Payable -}} cmd.ValueFlagValue.Int(), {{- end -}}
            cmd.BlockFlagValue.Int,
        )
        if err != nil {
            return err
        }

        {{ if gt (len $method.Return.Type) 0 -}}
        cmd.PrintOutput(result)
        {{- else -}}
        cmd.PrintOutput("success")
        {{- end }}
    }


    return nil
}

{{- end -}}
{{- end }}

/// ------------------- Initialization -------------------

func initialize{{.Class}}(c *cobra.Command) (*contract.{{.Class}}, error) {
    cfg, err := {{.ConfigReader}}(c.Flags())
    if err != nil {
        return nil, fmt.Errorf("error reading config from file: [%v]", err)
    }

    client, err := ethclient.Dial(cfg.URL)
    if err != nil {
        return nil, fmt.Errorf("error connecting to host chain node: [%v]", err)
    }

   	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf(
			"failed to resolve host chain id: [%v]",
			err,
		)
	}

    key, err := chainutil.DecryptKeyFile(
        cfg.Account.KeyFile,
        cfg.Account.KeyFilePassword,
    )
    if err != nil {
        return nil, fmt.Errorf(
            "failed to read KeyFile: %s: [%v]",
            cfg.Account.KeyFile,
            err,
        )
    }

	miningWaiter := chainutil.NewMiningWaiter(client, cfg)

	blockCounter, err := chainutil.NewBlockCounter(client)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create block counter: [%v]",
			err,
		)
	}

    address, err := cfg.ContractAddress("{{.Class}}")
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get %s address: [%w]",
            "{{.Class}}",
			err,
		)
	}

    return contract.New{{.Class}}(
        address,
        chainID,
        key,
        client,
        chainutil.NewNonceManager(client, key.Address),
        miningWaiter,
        blockCounter,
        &sync.Mutex{},
    )
}
`
