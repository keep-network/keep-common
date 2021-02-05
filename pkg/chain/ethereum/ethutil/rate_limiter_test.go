package ethutil

import (
	"context"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"github.com/keep-network/keep-common/pkg/chain/ethlike/ethliketest"
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

	client := ethliketest.NewMockClient(requestDuration)

	rateLimitingClient := WrapRateLimiting(
		client,
		&RateLimiterConfig{
			RequestsPerSecondLimit: requestsPerSecondLimit,
			ConcurrencyLimit:       concurrencyLimit,
			AcquirePermitTimeout:   acquirePermitTimeout,
		},
	)

	for testName, test := range getTests(rateLimitingClient) {
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
			for _, event := range client.EventsSnapshot() {
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

	client := ethliketest.NewMockClient(requestDuration)

	rateLimitingClient := WrapRateLimiting(
		client,
		&RateLimiterConfig{
			RequestsPerSecondLimit: requestsPerSecondLimit,
			ConcurrencyLimit:       concurrencyLimit,
			AcquirePermitTimeout:   acquirePermitTimeout,
		},
	)

	for testName, test := range getTests(rateLimitingClient) {
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
			// Here we set the maximum acceptable deviation to 5%.
			maxDeviation := 0.05

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

	client := ethliketest.NewMockClient(requestDuration)

	rateLimitingClient := WrapRateLimiting(
		client,
		&RateLimiterConfig{
			RequestsPerSecondLimit: requestsPerSecondLimit,
			ConcurrencyLimit:       concurrencyLimit,
			AcquirePermitTimeout:   acquirePermitTimeout,
		},
	)

	for testName, test := range getTests(rateLimitingClient) {
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
			for _, event := range client.EventsSnapshot() {
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

	client := ethliketest.NewMockClient(requestDuration)

	rateLimitingClient := WrapRateLimiting(
		client,
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

			err := rateLimitingClient.SendTransaction(context.Background(), nil)
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

func getTests(
	client ethlike.Client,
) map[string]struct{ function func() error } {
	return map[string]struct{ function func() error }{
		"test CodeAt": {
			function: func() error {
				_, err := client.CodeAt(
					context.Background(),
					&ethliketest.MockAddress{},
					nil,
				)
				return err
			},
		},
		"test CallContract": {
			function: func() error {
				_, err := client.CallContract(
					context.Background(),
					&ethliketest.MockCallMsg{},
					nil,
				)
				return err
			},
		},
		"test PendingCodeAt": {
			function: func() error {
				_, err := client.PendingCodeAt(
					context.Background(),
					&ethliketest.MockAddress{},
				)
				return err
			},
		},
		"test PendingNonceAt": {
			function: func() error {
				_, err := client.PendingNonceAt(
					context.Background(),
					&ethliketest.MockAddress{},
				)
				return err
			},
		},
		"test SuggestGasPrice": {
			function: func() error {
				_, err := client.SuggestGasPrice(
					context.Background(),
				)
				return err
			},
		},
		"test EstimateGas": {
			function: func() error {
				_, err := client.EstimateGas(
					context.Background(),
					&ethliketest.MockCallMsg{},
				)
				return err
			},
		},
		"test SendTransaction": {
			function: func() error {
				err := client.SendTransaction(
					context.Background(),
					nil,
				)
				return err
			},
		},
		"test FilterLogs": {
			function: func() error {
				_, err := client.FilterLogs(
					context.Background(),
					&ethliketest.MockFilterQuery{},
				)
				return err
			},
		},
		"test SubscribeFilterLogs": {
			function: func() error {
				_, err := client.SubscribeFilterLogs(
					context.Background(),
					&ethliketest.MockFilterQuery{},
					nil,
				)
				return err
			},
		},
		"test BlockByHash": {
			function: func() error {
				_, err := client.BlockByHash(
					context.Background(),
					&ethliketest.MockHash{},
				)
				return err
			},
		},
		"test BlockByNumber": {
			function: func() error {
				_, err := client.BlockByNumber(
					context.Background(),
					nil,
				)
				return err
			},
		},
		"test HeaderByHash": {
			function: func() error {
				_, err := client.HeaderByHash(
					context.Background(),
					&ethliketest.MockHash{},
				)
				return err
			},
		},
		"test HeaderByNumber": {
			function: func() error {
				_, err := client.HeaderByNumber(
					context.Background(),
					nil,
				)
				return err
			},
		},
		"test TransactionCount": {
			function: func() error {
				_, err := client.TransactionCount(
					context.Background(),
					&ethliketest.MockHash{},
				)
				return err
			},
		},
		"test TransactionInBlock": {
			function: func() error {
				_, err := client.TransactionInBlock(
					context.Background(),
					&ethliketest.MockHash{},
					0,
				)
				return err
			},
		},
		"test SubscribeNewHead": {
			function: func() error {
				_, err := client.SubscribeNewHead(
					context.Background(),
					nil,
				)
				return err
			},
		},
		"test TransactionByHash": {
			function: func() error {
				_, _, err := client.TransactionByHash(
					context.Background(),
					&ethliketest.MockHash{},
				)
				return err
			},
		},
		"test TransactionReceipt": {
			function: func() error {
				_, err := client.TransactionReceipt(
					context.Background(),
					&ethliketest.MockHash{},
				)
				return err
			},
		},
		"test BalanceAt": {
			function: func() error {
				_, err := client.BalanceAt(
					context.Background(),
					&ethliketest.MockAddress{},
					big.NewInt(0),
				)
				return err
			},
		},
	}
}
