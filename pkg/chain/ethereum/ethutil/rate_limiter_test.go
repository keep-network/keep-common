package ethutil

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"
)

const requestDuration = 250 * time.Millisecond

func TestRateLimiter(t *testing.T) {
	requestsPerSecondLimit := 200
	concurrencyLimit := 50
	acquirePermitTimeout := time.Minute
	requests := 1000

	backend := &mockBackend{make([]string, 0), sync.Mutex{}}

	rateLimitingBackend := WrapRateLimiting(
		backend,
		&RateLimiterConfig{
			RequestsPerSecondLimit: requestsPerSecondLimit,
			ConcurrencyLimit:       concurrencyLimit,
			AcquirePermitTimeout:   acquirePermitTimeout,
		},
	)

	wg := sync.WaitGroup{}
	wg.Add(requests)

	startSignal := make(chan struct{})

	for i := 0; i < requests; i++ {
		go func() {
			<-startSignal

			err := rateLimitingBackend.SendTransaction(context.Background(), nil)
			if err != nil {
				t.Errorf("unexpected error: [%v]", err)
			}

			wg.Done()
		}()
	}

	startTime := time.Now()
	close(startSignal)

	wg.Wait()

	duration := time.Now().Sub(startTime)
	averageRequestsPerSecond := float64(requests) / duration.Seconds()

	if averageRequestsPerSecond > float64(requestsPerSecondLimit) {
		t.Errorf(
			"average requests per second exceeded the limit\n"+
				"limit:  [%v]\n"+
				"actual: [%v]",
			requestsPerSecondLimit,
			averageRequestsPerSecond,
		)
	}

	maxConcurrency := 0
	temporaryConcurrency := 0
	for _, event := range backend.events {
		if event == "start" {
			temporaryConcurrency++
		}

		if event == "end" {
			temporaryConcurrency--
		}

		if temporaryConcurrency > maxConcurrency {
			maxConcurrency = temporaryConcurrency
		}
	}

	if maxConcurrency > concurrencyLimit {
		t.Errorf(
			"max concurrency exceeded the limit\n"+
				"limit:  [%v]\n"+
				"actual: [%v]",
			concurrencyLimit,
			maxConcurrency,
		)
	}

	fmt.Printf(
		"actual average requests per second: [%v]\n"+
			"actual maximum concurrency: [%v]\n",
		averageRequestsPerSecond,
		maxConcurrency,
	)
}

func TestRateLimiter_AcquirePermitTimout(t *testing.T) {
	requestsPerSecondLimit := 1
	concurrencyLimit := 1
	acquirePermitTimeout := 10 * time.Millisecond
	requests := 3

	backend := &mockBackend{make([]string, 0), sync.Mutex{}}

	rateLimitingBackend := WrapRateLimiting(
		backend,
		&RateLimiterConfig{
			RequestsPerSecondLimit: requestsPerSecondLimit,
			ConcurrencyLimit:       concurrencyLimit,
			AcquirePermitTimeout:   acquirePermitTimeout,
		},
	)

	wg := sync.WaitGroup{}
	wg.Add(requests)

	startSignal := make(chan struct{})
	errorChannel := make(chan error, requests)
	errors := make([]error, 0)

	go func() {
		for e := range errorChannel {
			errors = append(errors, e)
		}
	}()

	for i := 0; i < requests; i++ {
		go func() {
			<-startSignal

			err := rateLimitingBackend.SendTransaction(context.Background(), nil)
			if err != nil {
				errorChannel <- err
			}

			wg.Done()
		}()
	}

	close(startSignal)

	wg.Wait()

	if len(errors) == 0 {
		t.Fatalf("at least one timeout error should be present")
	}

	for _, e := range errors {
		if !strings.Contains(e.Error(), "context deadline") {
			t.Errorf(
				"error should be related with the context deadline\n"+
					"actual error: [%v]",
				e,
			)
		}
	}
}

type mockBackend struct {
	events []string
	mutex  sync.Mutex
}

func (mb *mockBackend) CodeAt(
	ctx context.Context,
	contract common.Address,
	blockNumber *big.Int,
) ([]byte, error) {
	return nil, nil
}

func (mb *mockBackend) CallContract(
	ctx context.Context,
	call ethereum.CallMsg,
	blockNumber *big.Int,
) ([]byte, error) {
	return nil, nil
}

func (mb *mockBackend) PendingCodeAt(
	ctx context.Context,
	account common.Address,
) ([]byte, error) {
	return nil, nil
}

func (mb *mockBackend) PendingNonceAt(
	ctx context.Context,
	account common.Address,
) (uint64, error) {
	return 0, nil
}

func (mb *mockBackend) SuggestGasPrice(
	ctx context.Context,
) (*big.Int, error) {
	return nil, nil
}

func (mb *mockBackend) EstimateGas(
	ctx context.Context,
	call ethereum.CallMsg,
) (uint64, error) {
	return 0, nil
}

func (mb *mockBackend) SendTransaction(
	ctx context.Context,
	tx *types.Transaction,
) error {
	mb.mutex.Lock()
	mb.events = append(mb.events, "start")
	mb.mutex.Unlock()

	time.Sleep(requestDuration)

	mb.mutex.Lock()
	mb.events = append(mb.events, "end")
	mb.mutex.Unlock()

	return nil
}

func (mb *mockBackend) FilterLogs(
	ctx context.Context,
	query ethereum.FilterQuery,
) ([]types.Log, error) {
	return nil, nil
}

func (mb *mockBackend) SubscribeFilterLogs(
	ctx context.Context,
	query ethereum.FilterQuery,
	ch chan<- types.Log,
) (ethereum.Subscription, error) {
	return nil, nil
}
