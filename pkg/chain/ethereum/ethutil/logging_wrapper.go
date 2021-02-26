// Code generated - DO NOT EDIT.

package ethutil

import (
	"context"
	hostchain "github.com/ethereum/go-ethereum"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil/client"
	"math/big"
)

type loggingWrapper struct {
	client.ChainClient

	logger log.EventLogger
}

func (lw *loggingWrapper) SuggestGasPrice(
	ctx context.Context,
) (*big.Int, error) {
	price, err := lw.ChainClient.SuggestGasPrice(ctx)

	if err != nil {
		lw.logger.Debugf("error requesting gas price suggestion: [%v]", err)
		return nil, err
	}

	lw.logger.Debugf("received gas price suggestion: [%v]", price)
	return price, nil
}

func (lw *loggingWrapper) EstimateGas(
	ctx context.Context,
	msg hostchain.CallMsg,
) (uint64, error) {
	gas, err := lw.ChainClient.EstimateGas(ctx, msg)

	if err != nil {
		return 0, err
	}

	lw.logger.Debugf("received gas estimate: [%v]", gas)
	return gas, nil
}

// WrapCallLogging wraps certain call-related methods on the given `client`
// with debug logging sent to the given `logger`. Actual functionality is
// delegated to the passed client.
func WrapCallLogging(
	logger log.EventLogger,
	client client.ChainClient,
) client.ChainClient {
	return &loggingWrapper{client, logger}
}
