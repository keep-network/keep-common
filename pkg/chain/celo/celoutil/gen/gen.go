package gen

//go:generate go run github.com/keep-network/keep-common/pkg/chain/ethlike/generator -host-chain-module github.com/celo-org/celo-blockchain -chain-util-package github.com/keep-network/keep-common/pkg/chain/celo/celoutil ../../../ethlike/generator/template/resubscribe.go.tmpl resubscribe.go
//go:generate go run github.com/keep-network/keep-common/pkg/chain/ethlike/generator -host-chain-module github.com/celo-org/celo-blockchain -chain-util-package github.com/keep-network/keep-common/pkg/chain/celo/celoutil ../../../ethlike/generator/template/error_resolver.go.tmpl error_resolver.go
//go:generate go run github.com/keep-network/keep-common/pkg/chain/ethlike/generator -host-chain-module github.com/celo-org/celo-blockchain -chain-util-package github.com/keep-network/keep-common/pkg/chain/celo/celoutil ../../../ethlike/generator/template/logging_wrapper.go.tmpl logging_wrapper.go
//go:generate go run github.com/keep-network/keep-common/pkg/chain/ethlike/generator -host-chain-module github.com/celo-org/celo-blockchain -chain-util-package github.com/keep-network/keep-common/pkg/chain/celo/celoutil ../../../ethlike/generator/template/rate_limiter.go.tmpl rate_limiter.go
