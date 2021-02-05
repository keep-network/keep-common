package ethlikeutil

import (
	"context"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"math/big"
)

type loggingWrapper struct {
	ethlike.Client

	logger log.EventLogger
}

func (lw *loggingWrapper) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	price, err := lw.Client.SuggestGasPrice(ctx)

	if err != nil {
		lw.logger.Debugf("error requesting gas price suggestion: [%v]", err)
		return nil, err
	}

	lw.logger.Debugf("received gas price suggestion: [%v]", price)
	return price, nil
}

func (lw *loggingWrapper) EstimateGas(ctx context.Context, msg ethlike.CallMsg) (uint64, error) {
	gas, err := lw.Client.EstimateGas(ctx, msg)

	if err != nil {
		return 0, err
	}

	lw.logger.Debugf("received gas estimate: [%v]", gas)
	return gas, nil
}

// WrapCallLogging wraps certain call-related methods on the given `client`
// with debug logging sent to the given `logger`. Actual functionality is
// delegated to the passed client.
func WrapCallLogging(logger log.EventLogger, client ethlike.Client) ethlike.Client {
	return &loggingWrapper{client, logger}
}
