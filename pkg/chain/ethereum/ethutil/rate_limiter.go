package ethutil

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
	"math/big"
	"time"
)

type rateLimiter struct {
	bind.ContractBackend

	limiter   *rate.Limiter
	semaphore *semaphore.Weighted

	acquirePermitTimeout time.Duration
}

// RateLimiterConfig represents the configuration of the rate limiter.
type RateLimiterConfig struct {
	// RequestsPerSecondLimit sets the maximum average number of requests
	// per second. It's important to note that in short periods of time
	// the actual average may exceed this limit slightly.
	RequestsPerSecondLimit int

	// ConcurrencyLimit sets the maximum number of concurrent requests which
	// can be executed against the underlying contract backend in the same time.
	ConcurrencyLimit int

	// AcquirePermitTimeout determines how long a request can wait trying
	// to acquire a permit from the rate limiter.
	AcquirePermitTimeout time.Duration
}

// WrapRateLimiting wraps the given contract backend with rate limiting
// capabilities with respect to the provided configuration.
func WrapRateLimiting(
	backend bind.ContractBackend,
	config *RateLimiterConfig,
) bind.ContractBackend {
	rateLimiter := &rateLimiter{ContractBackend: backend}

	if config.RequestsPerSecondLimit > 0 {
		rateLimiter.limiter = rate.NewLimiter(
			rate.Limit(config.RequestsPerSecondLimit),
			1,
		)
	}

	if config.ConcurrencyLimit > 0 {
		rateLimiter.semaphore = semaphore.NewWeighted(
			int64(config.ConcurrencyLimit),
		)
	}

	if config.AcquirePermitTimeout > 0 {
		rateLimiter.acquirePermitTimeout = config.AcquirePermitTimeout
	} else {
		rateLimiter.acquirePermitTimeout = 5 * time.Minute
	}

	return rateLimiter
}

func (rl *rateLimiter) acquirePermit() error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		rl.acquirePermitTimeout,
	)
	defer cancel()

	if rl.limiter != nil {
		err := rl.limiter.Wait(ctx)
		if err != nil {
			return err
		}
	}

	if rl.semaphore != nil {
		err := rl.semaphore.Acquire(ctx, 1)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rl *rateLimiter) releasePermit() {
	if rl.semaphore != nil {
		rl.semaphore.Release(1)
	}
}

func (rl *rateLimiter) CodeAt(
	ctx context.Context,
	contract common.Address,
	blockNumber *big.Int,
) ([]byte, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.ContractBackend.CodeAt(ctx, contract, blockNumber)
}

func (rl *rateLimiter) CallContract(
	ctx context.Context,
	call ethereum.CallMsg,
	blockNumber *big.Int,
) ([]byte, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.ContractBackend.CallContract(ctx, call, blockNumber)
}

func (rl *rateLimiter) PendingCodeAt(
	ctx context.Context,
	account common.Address,
) ([]byte, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.ContractBackend.PendingCodeAt(ctx, account)
}

func (rl *rateLimiter) PendingNonceAt(
	ctx context.Context,
	account common.Address,
) (uint64, error) {
	err := rl.acquirePermit()
	if err != nil {
		return 0, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.ContractBackend.PendingNonceAt(ctx, account)
}

func (rl *rateLimiter) SuggestGasPrice(
	ctx context.Context,
) (*big.Int, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.ContractBackend.SuggestGasPrice(ctx)
}

func (rl *rateLimiter) EstimateGas(
	ctx context.Context,
	call ethereum.CallMsg,
) (uint64, error) {
	err := rl.acquirePermit()
	if err != nil {
		return 0, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.ContractBackend.EstimateGas(ctx, call)
}

func (rl *rateLimiter) SendTransaction(
	ctx context.Context,
	tx *types.Transaction,
) error {
	err := rl.acquirePermit()
	if err != nil {
		return fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.ContractBackend.SendTransaction(ctx, tx)
}

func (rl *rateLimiter) FilterLogs(
	ctx context.Context,
	query ethereum.FilterQuery,
) ([]types.Log, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.ContractBackend.FilterLogs(ctx, query)
}

func (rl *rateLimiter) SubscribeFilterLogs(
	ctx context.Context,
	query ethereum.FilterQuery,
	ch chan<- types.Log,
) (ethereum.Subscription, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.ContractBackend.SubscribeFilterLogs(ctx, query, ch)
}
