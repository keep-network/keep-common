package ethlikeutil

import (
	"context"
	"fmt"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
	"math/big"
	"time"
)

type rateLimiter struct {
	ethlike.Client

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
	// can be executed against the underlying contract backend at the same time.
	ConcurrencyLimit int

	// AcquirePermitTimeout determines how long a request can wait trying
	// to acquire a permit from the rate limiter.
	AcquirePermitTimeout time.Duration
}

// WrapRateLimiting wraps the given contract backend with rate limiting
// capabilities with respect to the provided configuration.
// All types of requests to the contract are rate-limited,
// including view function calls.
func WrapRateLimiting(
	client ethlike.Client,
	config *RateLimiterConfig,
) ethlike.Client {
	rateLimiter := &rateLimiter{Client: client}

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
	contract ethlike.Address,
	blockNumber *big.Int,
) ([]byte, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.CodeAt(ctx, contract, blockNumber)
}

func (rl *rateLimiter) CallContract(
	ctx context.Context,
	call ethlike.CallMsg,
	blockNumber *big.Int,
) ([]byte, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.CallContract(ctx, call, blockNumber)
}

func (rl *rateLimiter) PendingCodeAt(
	ctx context.Context,
	account ethlike.Address,
) ([]byte, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.PendingCodeAt(ctx, account)
}

func (rl *rateLimiter) PendingNonceAt(
	ctx context.Context,
	account ethlike.Address,
) (uint64, error) {
	err := rl.acquirePermit()
	if err != nil {
		return 0, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.PendingNonceAt(ctx, account)
}

func (rl *rateLimiter) SuggestGasPrice(
	ctx context.Context,
) (*big.Int, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.SuggestGasPrice(ctx)
}

func (rl *rateLimiter) EstimateGas(
	ctx context.Context,
	call ethlike.CallMsg,
) (uint64, error) {
	err := rl.acquirePermit()
	if err != nil {
		return 0, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.EstimateGas(ctx, call)
}

func (rl *rateLimiter) SendTransaction(
	ctx context.Context,
	tx ethlike.Transaction,
) error {
	err := rl.acquirePermit()
	if err != nil {
		return fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.SendTransaction(ctx, tx)
}

func (rl *rateLimiter) FilterLogs(
	ctx context.Context,
	query ethlike.FilterQuery,
) ([]ethlike.Log, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.FilterLogs(ctx, query)
}

func (rl *rateLimiter) SubscribeFilterLogs(
	ctx context.Context,
	query ethlike.FilterQuery,
	ch chan<- ethlike.Log,
) (ethlike.Subscription, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.SubscribeFilterLogs(ctx, query, ch)
}

func (rl *rateLimiter) BlockByHash(
	ctx context.Context,
	hash ethlike.Hash,
) (ethlike.Block, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.BlockByHash(ctx, hash)
}

func (rl *rateLimiter) BlockByNumber(
	ctx context.Context,
	number *big.Int,
) (ethlike.Block, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.BlockByNumber(ctx, number)
}

func (rl *rateLimiter) HeaderByHash(
	ctx context.Context,
	hash ethlike.Hash,
) (ethlike.Header, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.HeaderByHash(ctx, hash)
}

func (rl *rateLimiter) HeaderByNumber(
	ctx context.Context,
	number *big.Int,
) (ethlike.Header, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.HeaderByNumber(ctx, number)
}

func (rl *rateLimiter) TransactionCount(
	ctx context.Context,
	blockHash ethlike.Hash,
) (uint, error) {
	err := rl.acquirePermit()
	if err != nil {
		return 0, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.TransactionCount(ctx, blockHash)
}

func (rl *rateLimiter) TransactionInBlock(
	ctx context.Context,
	blockHash ethlike.Hash,
	index uint,
) (ethlike.Transaction, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.TransactionInBlock(ctx, blockHash, index)
}

func (rl *rateLimiter) SubscribeNewHead(
	ctx context.Context,
	ch chan<- ethlike.Header,
) (ethlike.Subscription, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.SubscribeNewHead(ctx, ch)
}

func (rl *rateLimiter) TransactionByHash(
	ctx context.Context,
	txHash ethlike.Hash,
) (ethlike.Transaction, bool, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, false, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.TransactionByHash(ctx, txHash)
}

func (rl *rateLimiter) TransactionReceipt(
	ctx context.Context,
	txHash ethlike.Hash,
) (ethlike.Receipt, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.TransactionReceipt(ctx, txHash)
}

func (rl *rateLimiter) BalanceAt(
	ctx context.Context,
	account ethlike.Address,
	blockNumber *big.Int,
) (*big.Int, error) {
	err := rl.acquirePermit()
	if err != nil {
		return nil, fmt.Errorf("cannot acquire rate limiter permit: [%v]", err)
	}
	defer rl.releasePermit()

	return rl.Client.BalanceAt(ctx, account, blockNumber)
}
