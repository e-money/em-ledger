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

func (k Keeper) DenomTraceAll(c context.Context, req *types.QueryAllDenomTraceRequest) (*types.QueryAllDenomTraceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var denomTraces []*types.DenomTrace
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	denomTraceStore := prefix.NewStore(store, types.KeyPrefix(types.DenomTraceKeyPrefix))

	pageRes, err := query.Paginate(denomTraceStore, req.Pagination, func(key []byte, value []byte) error {
		var denomTrace types.DenomTrace
		if err := k.cdc.UnmarshalBinaryBare(value, &denomTrace); err != nil {
			return err
		}

		denomTraces = append(denomTraces, &denomTrace)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllDenomTraceResponse{DenomTrace: denomTraces, Pagination: pageRes}, nil
}

func (k Keeper) DenomTrace(c context.Context, req *types.QueryGetDenomTraceRequest) (*types.QueryGetDenomTraceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetDenomTrace(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "not found")
	}

	return &types.QueryGetDenomTraceResponse{DenomTrace: &val}, nil
}
