package celoutil

//go:generate go run github.com/keep-network/keep-common/pkg/chain/ethlike/generator -package-name celoutil -host-chain-module github.com/celo-org/celo-blockchain ../../ethlike/generator/template/resubscribe.go.tmpl resubscribe.go
//go:generate go run github.com/keep-network/keep-common/pkg/chain/ethlike/generator -package-name celoutil -host-chain-module github.com/celo-org/celo-blockchain ../../ethlike/generator/template/error_resolver.go.tmpl error_resolver.go
//go:generate go run github.com/keep-network/keep-common/pkg/chain/ethlike/generator -package-name celoutil -host-chain-module github.com/celo-org/celo-blockchain ../../ethlike/generator/template/logging_wrapper.go.tmpl logging_wrapper.go
