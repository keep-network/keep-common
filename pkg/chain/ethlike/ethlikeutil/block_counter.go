package ethlikeutil

import (
	"context"
	"fmt"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"strconv"
	"sync"
	"time"
)

type BlockCounter struct {
	structMutex         sync.Mutex
	latestBlockHeight   uint64
	subscriptionChannel chan block
	waiters             map[uint64][]chan uint64
	watchers            []*watcher
}

type block struct {
	Number string
}

type watcher struct {
	ctx     context.Context
	channel chan uint64
}

func (bc *BlockCounter) WaitForBlockHeight(blockNumber uint64) error {
	waiter, err := bc.BlockHeightWaiter(blockNumber)
	if err != nil {
		return err
	}
	<-waiter
	return nil
}

func (bc *BlockCounter) BlockHeightWaiter(
	blockNumber uint64,
) (<-chan uint64, error) {
	newWaiter := make(chan uint64)

	bc.structMutex.Lock()
	defer bc.structMutex.Unlock()

	if blockNumber <= bc.latestBlockHeight {
		go func() { newWaiter <- blockNumber }()
	} else {
		waiterList, exists := bc.waiters[blockNumber]
		if !exists {
			waiterList = make([]chan uint64, 0)
		}

		bc.waiters[blockNumber] = append(waiterList, newWaiter)
	}

	return newWaiter, nil
}

func (bc *BlockCounter) CurrentBlock() (uint64, error) {
	return bc.latestBlockHeight, nil
}

func (bc *BlockCounter) WatchBlocks(ctx context.Context) <-chan uint64 {
	watcher := &watcher{
		ctx:     ctx,
		channel: make(chan uint64),
	}

	bc.structMutex.Lock()
	bc.watchers = append(bc.watchers, watcher)
	bc.structMutex.Unlock()

	go func() {
		<-ctx.Done()

		bc.structMutex.Lock()
		for i, w := range bc.watchers {
			if w == watcher {
				bc.watchers[i] = bc.watchers[len(bc.watchers)-1]
				bc.watchers = bc.watchers[:len(bc.watchers)-1]
				break
			}
		}
		bc.structMutex.Unlock()
	}()

	return watcher.channel
}

// receiveBlocks gets each new block back from Geth and extracts the
// block height (topBlockNumber) form it. For each block height that is being
// waited on a message will be sent.
func (bc *BlockCounter) receiveBlocks() {
	for block := range bc.subscriptionChannel {
		topBlockNumber, err := strconv.ParseInt(block.Number, 0, 32)
		if err != nil {
			logger.Errorf("error receiving a new block: [%v]", err)
			continue
		}

		// receivedBlockHeight is the current blockchain height as just
		// received in the notification. latestBlockHeightSeen is the
		// blockchain height as observed in the previous invocation of
		// receiveBlocks().
		//
		// If we have already received notification about this block,
		// we do nothing. All handlers were already called for this block
		// height.
		receivedBlockHeight := uint64(topBlockNumber)
		if receivedBlockHeight == bc.latestBlockHeight {
			continue
		}

		// We have already seen latestBlockHeightSeen during the previous
		// execution of receiveBlocks() function and all handlers for
		// latestBlockHeightSeen were called. Now we start from the next block
		// after it and that's latestBlockHeightSeen + 1.
		for unseenBlockNumber := bc.latestBlockHeight + 1; unseenBlockNumber <= receivedBlockHeight; unseenBlockNumber++ {
			bc.structMutex.Lock()
			height := unseenBlockNumber
			bc.latestBlockHeight++
			waiters := bc.waiters[height]
			delete(bc.waiters, height)
			bc.structMutex.Unlock()

			for _, waiter := range waiters {
				go func(w chan uint64) { w <- height }(waiter)
			}

			bc.structMutex.Lock()
			watchers := make([]*watcher, len(bc.watchers))
			copy(watchers, bc.watchers)
			bc.structMutex.Unlock()

			for _, watcher := range watchers {
				if watcher.ctx.Err() != nil {
					close(watcher.channel)
					continue
				}

				select {
				case watcher.channel <- height: // perfect
				default: // we don't care, let's drop it
				}
			}
		}
	}
}

// subscribeBlocks creates a subscription to Geth to get each block.
func (bc *BlockCounter) subscribeBlocks(ctx context.Context, client ethlike.ChainReader) error {
	errorChan := make(chan error)
	newBlockChan := make(chan ethlike.Header)

	subscribe := func() {
		logger.Debugf("subscribing to new blocks")

		subscribeContext, cancel := context.WithTimeout(
			ctx,
			10*time.Second, // timeout for subscription request
		)
		defer cancel()

		subscription, err := client.SubscribeNewHead(
			subscribeContext,
			newBlockChan,
		)
		if err != nil {
			logger.Warningf("could not create subscription to new blocks: [%v]", err)
			errorChan <- err
			return
		}

		for {
			select {
			case header := <-newBlockChan:
				bc.subscriptionChannel <- block{header.Number().String()}
			case err = <-subscription.Err():
				logger.Warningf("subscription to new blocks interrupted: [%v]", err)
				subscription.Unsubscribe()
				errorChan <- err
				return
			}
		}

	}

	go func() {
		for {
			go subscribe()
			<-errorChan
			time.Sleep(5 * time.Second)
		}
	}()

	lastBlock, err := client.BlockByNumber(
		ctx,
		nil, // if `nil` then latest known block is returned
	)
	if err != nil {
		return err
	}

	bc.subscriptionChannel <- block{lastBlock.Number().String()}

	return nil
}

func CreateBlockCounter(client ethlike.ChainReader) (*BlockCounter, error) {
	ctx := context.Background()

	startupBlock, err := client.BlockByNumber(
		ctx,
		nil, // if `nil` then latest known block is returned
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"failed to get initial block from the chain: [%v]",
				err,
			)
	}

	blockCounter := &BlockCounter{
		latestBlockHeight:   startupBlock.NumberU64(),
		waiters:             make(map[uint64][]chan uint64),
		subscriptionChannel: make(chan block),
	}

	go blockCounter.receiveBlocks()
	err = blockCounter.subscribeBlocks(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new blocks: [%v]", err)
	}

	return blockCounter, nil
}
