package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/market/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) ByAccount(c context.Context, req *types.QueryByAccountRequest) (*types.QueryByAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	account, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}

	orders := k.GetOrdersByOwner(ctx, account)
	return &types.QueryByAccountResponse{Orders: orders}, nil
}

func (k Keeper) Instruments(c context.Context, req *types.QueryInstrumentsRequest) (*types.QueryInstrumentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	res, err := queryInstruments(ctx, &k)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrIO, fmt.Sprintf("queryInstruments failed: %v", err))
	}
	return res, nil
}

func (k Keeper) Instrument(c context.Context, req *types.QueryInstrumentRequest) (*types.QueryInstrumentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	source, destination := req.Source, req.Destination
	if sdk.ValidateDenom(source) != nil || sdk.ValidateDenom(destination) != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "Invalid denoms: %v %v", source, destination)
	}

	orders := make([]types.QueryOrderResponse, 0)

	idxStore := ctx.KVStore(k.keyIndices)
	key := types.GetPriorityKeyBySrcAndDst(source, destination)

	it := sdk.KVStorePrefixIterator(idxStore, key)
	defer it.Close()

	for it.Valid() {
		order := new(types.Order)
		k.cdc.MustUnmarshal(it.Value(), order)

		orders = append(orders, types.QueryOrderResponse{
			ID:              order.ID,
			Owner:           order.Owner,
			SourceRemaining: order.SourceRemaining.String(),
			Price:           order.Price(),
			Created:         order.Created,
		})

		it.Next()
	}

	return &types.QueryInstrumentResponse{
		Source:      source,
		Destination: destination,
		Orders:      orders,
	}, nil
}

func queryInstruments(ctx sdk.Context, k *Keeper) (*types.QueryInstrumentsResponse, error) {
	instruments, err := k.GetAllInstruments(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]types.QueryInstrumentsResponse_Element, len(instruments))
	for i, v := range instruments {
		response[i] = types.QueryInstrumentsResponse_Element{
			Source:      v.Source,
			Destination: v.Destination,
			LastPrice:   v.LastPrice,
			BestPrice:   k.GetBestPrice(ctx, v.Source, v.Destination),
			LastTraded:  v.Timestamp,
		}
	}

	return &types.QueryInstrumentsResponse{Instruments: response}, nil
}
