package gen

//go:generate go run github.com/keep-network/keep-common/pkg/chain/ethlike/generator ../../../ethlike/generator/template/resubscribe.go.tmpl resubscribe.go
//go:generate go run github.com/keep-network/keep-common/pkg/chain/ethlike/generator ../../../ethlike/generator/template/error_resolver.go.tmpl error_resolver.go
//go:generate go run github.com/keep-network/keep-common/pkg/chain/ethlike/generator ../../../ethlike/generator/template/logging_wrapper.go.tmpl logging_wrapper.go
//go:generate go run github.com/keep-network/keep-common/pkg/chain/ethlike/generator ../../../ethlike/generator/template/rate_limiter.go.tmpl rate_limiter.go
