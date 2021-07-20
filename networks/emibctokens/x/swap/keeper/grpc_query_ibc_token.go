package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/e-money/stargate/networks/emibctokens/x/swap/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) IbcTokenAll(c context.Context, req *types.QueryAllIbcTokenRequest) (*types.QueryAllIbcTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var ibcTokens []*types.IbcToken
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	ibcTokenStore := prefix.NewStore(store, types.KeyPrefix(types.IbcTokenKeyPrefix))

	pageRes, err := query.Paginate(ibcTokenStore, req.Pagination, func(key []byte, value []byte) error {
		var ibcToken types.IbcToken
		if err := k.cdc.UnmarshalBinaryBare(value, &ibcToken); err != nil {
			return err
		}

		ibcTokens = append(ibcTokens, &ibcToken)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllIbcTokenResponse{IbcToken: ibcTokens, Pagination: pageRes}, nil
}

func (k Keeper) IbcToken(c context.Context, req *types.QueryGetIbcTokenRequest) (*types.QueryGetIbcTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetIbcToken(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "not found")
	}

	return &types.QueryGetIbcTokenResponse{IbcToken: &val}, nil
}
