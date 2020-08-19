package ethutil

import (
	"context"
	"github.com/ethereum/go-ethereum"
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

	tests := map[string]struct {
		function func() error
	}{
		"test CodeAt rate limiting": {
			function: func() error {
				_, err := rateLimitingBackend.CodeAt(
					context.Background(),
					[20]byte{},
					nil,
				)
				return err
			},
		},
		"test CallContract rate limiting": {
			function: func() error {
				_, err := rateLimitingBackend.CallContract(
					context.Background(),
					ethereum.CallMsg{},
					nil,
				)
				return err
			},
		},
		"test PendingCodeAt rate limiting": {
			function: func() error {
				_, err := rateLimitingBackend.PendingCodeAt(
					context.Background(),
					[20]byte{},
				)
				return err
			},
		},
		"test PendingNonceAt rate limiting": {
			function: func() error {
				_, err := rateLimitingBackend.PendingNonceAt(
					context.Background(),
					[20]byte{},
				)
				return err
			},
		},
		"test SuggestGasPrice rate limiting": {
			function: func() error {
				_, err := rateLimitingBackend.SuggestGasPrice(
					context.Background(),
				)
				return err
			},
		},
		"test EstimateGas rate limiting": {
			function: func() error {
				_, err := rateLimitingBackend.EstimateGas(
					context.Background(),
					ethereum.CallMsg{},
				)
				return err
			},
		},
		"test SendTransaction rate limiting": {
			function: func() error {
				err := rateLimitingBackend.SendTransaction(
					context.Background(),
					nil,
				)
				return err
			},
		},
		"test FilterLogs rate limiting": {
			function: func() error {
				_, err := rateLimitingBackend.FilterLogs(
					context.Background(),
					ethereum.FilterQuery{},
				)
				return err
			},
		},
		"test SubscribeFilterLogs rate limiting": {
			function: func() error {
				_, err := rateLimitingBackend.SubscribeFilterLogs(
					context.Background(),
					ethereum.FilterQuery{},
					nil,
				)
				return err
			},
		},
	}

	for testName, test := range tests {
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
