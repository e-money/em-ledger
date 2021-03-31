package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/authority/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) GasPrices(c context.Context, req *types.QueryGasPricesRequest) (*types.QueryGasPricesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	gasPrices := k.GetGasPrices(ctx)
	return &types.QueryGasPricesResponse{MinGasPrices: gasPrices}, nil
}
