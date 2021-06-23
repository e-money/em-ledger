package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) List(c context.Context, req *types.QueryListRequest) (*types.QueryListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	response := &types.QueryListResponse{
		LiquidityProviders: k.GetAllLiquidityProviderAccounts(sdk.UnwrapSDKContext(c)),
	}

	return response, nil
}

func (k Keeper) Mintable(c context.Context, req *types.QueryMintableRequest) (*types.QueryMintableResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	lqAcc, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider:" + req.Address)
	}

	lp := k.GetLiquidityProviderAccount(sdk.UnwrapSDKContext(c), lqAcc)
	if lp == nil {
		return &types.QueryMintableResponse{
			Mintable: sdk.NewCoins(),
		}, nil
	}

	response := types.QueryMintableResponse{
		Mintable: lp.Mintable,
	}

	return &response, nil
}