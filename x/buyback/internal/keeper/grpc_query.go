package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/buyback/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Balance(c context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	response := types.QueryBalanceResponse{
		Balance: k.bankKeeper.GetAllBalances(ctx, k.GetBuybackAccountAddr()),
	}
	return &response, nil
}

func (k Keeper) BuybackTime(c context.Context, req *types.QueryBuybackTimeRequest) (*types.QueryBuybackTimeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	updateInterval := k.GetUpdateInterval(ctx)
	lastUpdated := k.GetLastUpdated(ctx)

	nextRun := lastUpdated.Add(updateInterval)

	response := types.QueryBuybackTimeResponse{
		LastRunTime: lastUpdated,
		NextRunTime: nextRun,
	}

	return &response, nil
}
