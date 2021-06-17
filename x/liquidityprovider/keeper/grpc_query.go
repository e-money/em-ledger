package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Mintable(c context.Context, req *types.QueryMintableRequest) (*types.QueryMintableResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	lp := k.GetLiquidityProviderAccount(sdk.UnwrapSDKContext(c), req.Address)
	if lp == nil {
		return nil, status.Error(codes.NotFound, "liquidity provider address:" + req.Address)
	}


	response := types.QueryMintableResponse{
		Mintable: lp.Mintable,
	}
	return &response, nil
}
