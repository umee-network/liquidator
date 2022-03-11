package client

import (
	rpcClient "github.com/cosmos/cosmos-sdk/client/rpc"
)

func (lc LeverageClient) BlockHeight() (int64, error) {
	ctx, err := lc.CreateClientContext()
	if err != nil {
		return 0, err
	}
	return rpcClient.GetChainHeight(ctx)
}
