package celoutil

import (
	"context"
	"fmt"
	"github.com/celo-org/celo-blockchain"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/keep-network/keep-common/pkg/rate"
	"math/big"
)

type rateLimiter struct {
	CeloClient

	*rate.Limiter
}

// WrapRateLimiting wraps the given contract backend with rate limiting
// capabilities with respect to the provided configuration.
// All types of requests to the contract are rate-limited,
// including view function calls.
func WrapRateLimiting(
	client CeloClient,
	config *rate.LimiterConfig,
) CeloClient {
	return &rateLimiter{
		CeloClient: client,
		Limiter:    rate.NewLimiter(config),
	}
}

func (rl *rateLimiter) CodeAt(
	ctx context.Context,
	contract common.Address,
	blockNumber *big.Int,
) ([]byte, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.CodeAt(ctx, contract, blockNumber)
}

func (rl *rateLimiter) CallContract(
	ctx context.Context,
	call celo.CallMsg,
	blockNumber *big.Int,
) ([]byte, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.CallContract(ctx, call, blockNumber)
}

func (rl *rateLimiter) PendingCodeAt(
	ctx context.Context,
	account common.Address,
) ([]byte, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.PendingCodeAt(ctx, account)
}

func (rl *rateLimiter) PendingNonceAt(
	ctx context.Context,
	account common.Address,
) (uint64, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return 0, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.PendingNonceAt(ctx, account)
}

func (rl *rateLimiter) SuggestGasPrice(
	ctx context.Context,
) (*big.Int, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.SuggestGasPrice(ctx)
}

func (rl *rateLimiter) EstimateGas(
	ctx context.Context,
	call celo.CallMsg,
) (uint64, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return 0, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.EstimateGas(ctx, call)
}

func (rl *rateLimiter) SendTransaction(
	ctx context.Context,
	tx *types.Transaction,
) error {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.SendTransaction(ctx, tx)
}

func (rl *rateLimiter) FilterLogs(
	ctx context.Context,
	query celo.FilterQuery,
) ([]types.Log, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.FilterLogs(ctx, query)
}

func (rl *rateLimiter) SubscribeFilterLogs(
	ctx context.Context,
	query celo.FilterQuery,
	ch chan<- types.Log,
) (celo.Subscription, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.SubscribeFilterLogs(ctx, query, ch)
}

func (rl *rateLimiter) BlockByHash(
	ctx context.Context,
	hash common.Hash,
) (*types.Block, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.BlockByHash(ctx, hash)
}

func (rl *rateLimiter) BlockByNumber(
	ctx context.Context,
	number *big.Int,
) (*types.Block, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.BlockByNumber(ctx, number)
}

func (rl *rateLimiter) HeaderByHash(
	ctx context.Context,
	hash common.Hash,
) (*types.Header, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.HeaderByHash(ctx, hash)
}

func (rl *rateLimiter) HeaderByNumber(
	ctx context.Context,
	number *big.Int,
) (*types.Header, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.HeaderByNumber(ctx, number)
}

func (rl *rateLimiter) TransactionCount(
	ctx context.Context,
	blockHash common.Hash,
) (uint, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return 0, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.TransactionCount(ctx, blockHash)
}

func (rl *rateLimiter) TransactionInBlock(
	ctx context.Context,
	blockHash common.Hash,
	index uint,
) (*types.Transaction, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.TransactionInBlock(ctx, blockHash, index)
}

func (rl *rateLimiter) SubscribeNewHead(
	ctx context.Context,
	ch chan<- *types.Header,
) (celo.Subscription, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.SubscribeNewHead(ctx, ch)
}

func (rl *rateLimiter) TransactionByHash(
	ctx context.Context,
	txHash common.Hash,
) (*types.Transaction, bool, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, false, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.TransactionByHash(ctx, txHash)
}

func (rl *rateLimiter) TransactionReceipt(
	ctx context.Context,
	txHash common.Hash,
) (*types.Receipt, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.TransactionReceipt(ctx, txHash)
}

func (rl *rateLimiter) BalanceAt(
	ctx context.Context,
	account common.Address,
	blockNumber *big.Int,
) (*big.Int, error) {
	err := rl.Limiter.AcquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.Limiter.ReleasePermit()

	return rl.CeloClient.BalanceAt(ctx, account, blockNumber)
}
