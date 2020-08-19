package ethutil

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	requestsPerSecondLimit := 500
	concurrencyLimit := 5
	acquirePermitTimeout := time.Minute
	requests := 500
	requestDuration := 10 * time.Millisecond

	backend := &mockBackend{
		requestDuration,
		make([]string, 0),
		sync.Mutex{},
	}

	rateLimitingBackend := WrapRateLimiting(
		backend,
		&RateLimiterConfig{
			RequestsPerSecondLimit: requestsPerSecondLimit,
			ConcurrencyLimit:       concurrencyLimit,
			AcquirePermitTimeout:   acquirePermitTimeout,
		},
	)

	for testName, test := range getTests(rateLimitingBackend) {
		t.Run(testName, func(t *testing.T) {
			wg := sync.WaitGroup{}
			wg.Add(requests)

			startSignal := make(chan struct{})

			for i := 0; i < requests; i++ {
				go func() {
					<-startSignal

					err := test.function()
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
		})
	}
}

func TestRateLimiter_RequestsPerSecondLimitOnly(t *testing.T) {
	requestsPerSecondLimit := 500
	concurrencyLimit := 0 // disable the concurrency limit
	acquirePermitTimeout := time.Minute
	requests := 500
	requestDuration := 10 * time.Millisecond

	backend := &mockBackend{
		requestDuration,
		make([]string, 0),
		sync.Mutex{},
	}

	rateLimitingBackend := WrapRateLimiting(
		backend,
		&RateLimiterConfig{
			RequestsPerSecondLimit: requestsPerSecondLimit,
			ConcurrencyLimit:       concurrencyLimit,
			AcquirePermitTimeout:   acquirePermitTimeout,
		},
	)

	for testName, test := range getTests(rateLimitingBackend) {
		t.Run(testName, func(t *testing.T) {
			wg := sync.WaitGroup{}
			wg.Add(requests)

			startSignal := make(chan struct{})

			for i := 0; i < requests; i++ {
				go func() {
					<-startSignal

					err := test.function()
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

			// The actual average can exceed the limit a little bit.
			// Here we set the maximum acceptable deviation to 2%.
			maxDeviation := 0.02

			if averageRequestsPerSecond > (1+maxDeviation)*float64(requestsPerSecondLimit) {
				t.Errorf(
					"average requests per second exceeded the limit\n"+
						"limit:  [%v]\n"+
						"actual: [%v]",
					requestsPerSecondLimit,
					averageRequestsPerSecond,
				)
			}
		})
	}
}

func TestRateLimiter_ConcurrencyLimitOnly(t *testing.T) {
	requestsPerSecondLimit := 0 // disable the requests per second limit
	concurrencyLimit := 50
	acquirePermitTimeout := time.Minute
	requests := 500
	requestDuration := 10 * time.Millisecond

	backend := &mockBackend{
		requestDuration,
		make([]string, 0),
		sync.Mutex{},
	}

	rateLimitingBackend := WrapRateLimiting(
		backend,
		&RateLimiterConfig{
			RequestsPerSecondLimit: requestsPerSecondLimit,
			ConcurrencyLimit:       concurrencyLimit,
			AcquirePermitTimeout:   acquirePermitTimeout,
		},
	)

	for testName, test := range getTests(rateLimitingBackend) {
		t.Run(testName, func(t *testing.T) {
			wg := sync.WaitGroup{}
			wg.Add(requests)

			startSignal := make(chan struct{})

			for i := 0; i < requests; i++ {
				go func() {
					<-startSignal

					err := test.function()
					if err != nil {
						t.Errorf("unexpected error: [%v]", err)
					}

					wg.Done()
				}()
			}

			close(startSignal)

			wg.Wait()

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
		})
	}
}

func TestRateLimiter_AcquirePermitTimout(t *testing.T) {
	requestsPerSecondLimit := 1
	concurrencyLimit := 1
	acquirePermitTimeout := 10 * time.Millisecond
	requests := 3
	requestDuration := 250 * time.Millisecond

	backend := &mockBackend{
		requestDuration,
		make([]string, 0),
		sync.Mutex{},
	}

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
	errors := make(chan error, requests)

	for i := 0; i < requests; i++ {
		go func() {
			<-startSignal

			err := rateLimitingBackend.SendTransaction(context.Background(), nil)
			if err != nil {
				errors <- err
			}

			wg.Done()
		}()
	}

	close(startSignal)

	wg.Wait()

	close(errors)
	if len(errors) == 0 {
		t.Fatalf("at least one timeout error should be present")
	}

	for e := range errors {
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
	requestDuration time.Duration

	events []string
	mutex  sync.Mutex
}

func (mb *mockBackend) mockRequest() {
	mb.mutex.Lock()
	mb.events = append(mb.events, "start")
	mb.mutex.Unlock()

	time.Sleep(mb.requestDuration)

	mb.mutex.Lock()
	mb.events = append(mb.events, "end")
	mb.mutex.Unlock()
}

func (mb *mockBackend) CodeAt(
	ctx context.Context,
	contract common.Address,
	blockNumber *big.Int,
) ([]byte, error) {
	mb.mockRequest()
	return nil, nil
}

func (mb *mockBackend) CallContract(
	ctx context.Context,
	call ethereum.CallMsg,
	blockNumber *big.Int,
) ([]byte, error) {
	mb.mockRequest()
	return nil, nil
}

func (mb *mockBackend) PendingCodeAt(
	ctx context.Context,
	account common.Address,
) ([]byte, error) {
	mb.mockRequest()
	return nil, nil
}

func (mb *mockBackend) PendingNonceAt(
	ctx context.Context,
	account common.Address,
) (uint64, error) {
	mb.mockRequest()
	return 0, nil
}

func (mb *mockBackend) SuggestGasPrice(
	ctx context.Context,
) (*big.Int, error) {
	mb.mockRequest()
	return nil, nil
}

func (mb *mockBackend) EstimateGas(
	ctx context.Context,
	call ethereum.CallMsg,
) (uint64, error) {
	mb.mockRequest()
	return 0, nil
}

func (mb *mockBackend) SendTransaction(
	ctx context.Context,
	tx *types.Transaction,
) error {
	mb.mockRequest()
	return nil
}

func (mb *mockBackend) FilterLogs(
	ctx context.Context,
	query ethereum.FilterQuery,
) ([]types.Log, error) {
	mb.mockRequest()
	return nil, nil
}

func (mb *mockBackend) SubscribeFilterLogs(
	ctx context.Context,
	query ethereum.FilterQuery,
	ch chan<- types.Log,
) (ethereum.Subscription, error) {
	mb.mockRequest()
	return nil, nil
}

func getTests(
	backend bind.ContractBackend,
) map[string]struct{ function func() error } {
	return map[string]struct{ function func() error }{
		"test CodeAt": {
			function: func() error {
				_, err := backend.CodeAt(
					context.Background(),
					[20]byte{},
					nil,
				)
				return err
			},
		},
		"test CallContract": {
			function: func() error {
				_, err := backend.CallContract(
					context.Background(),
					ethereum.CallMsg{},
					nil,
				)
				return err
			},
		},
		"test PendingCodeAt": {
			function: func() error {
				_, err := backend.PendingCodeAt(
					context.Background(),
					[20]byte{},
				)
				return err
			},
		},
		"test PendingNonceAt": {
			function: func() error {
				_, err := backend.PendingNonceAt(
					context.Background(),
					[20]byte{},
				)
				return err
			},
		},
		"test SuggestGasPrice": {
			function: func() error {
				_, err := backend.SuggestGasPrice(
					context.Background(),
				)
				return err
			},
		},
		"test EstimateGas": {
			function: func() error {
				_, err := backend.EstimateGas(
					context.Background(),
					ethereum.CallMsg{},
				)
				return err
			},
		},
		"test SendTransaction": {
			function: func() error {
				err := backend.SendTransaction(
					context.Background(),
					nil,
				)
				return err
			},
		},
		"test FilterLogs": {
			function: func() error {
				_, err := backend.FilterLogs(
					context.Background(),
					ethereum.FilterQuery{},
				)
				return err
			},
		},
		"test SubscribeFilterLogs": {
			function: func() error {
				_, err := backend.SubscribeFilterLogs(
					context.Background(),
					ethereum.FilterQuery{},
					nil,
				)
				return err
			},
		},
	}
}
