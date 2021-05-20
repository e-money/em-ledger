// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"encoding/json"
	"fmt"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/market/types"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(k *Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryInstruments:
			result := queryInstruments(ctx, k)
			return json.Marshal(result)
		case types.QueryInstrument:
			return queryInstrument(ctx, k, path[1:], req)
		case types.QueryByAccount:
			return queryByAccount(ctx, k, path[1:], req)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unrecognized market query endpoint")
		}
	}
}

type OrderResponses []*types.Order

func (o OrderResponses) String() string {
	panic("implement me")
}

func (o OrderResponses) Len() int {
	return len(o)
}

func (o OrderResponses) Less(i, j int) bool {
	return o[i].ID < o[j].ID
}

func (o OrderResponses) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

var _ sort.Interface = OrderResponses{}

func queryByAccount(ctx sdk.Context, k *Keeper, path []string, req abci.RequestQuery) ([]byte, error) {
	if len(path) != 1 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "%s is not a valid query request path", req.Path)
	}

	account, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprint("Address could not be parsed", path[0], err))
	}

	// o := k.accountOrders.GetAllOrders(account)
	orders := k.GetOrdersByOwner(ctx, account)
	// orders := make(OrderResponses, 0)

	sort.Slice(
		orders, func(i, j int) bool {
			return orders[i].ID < orders[j].ID
		})

	resp := types.QueryByAccountResponse{Orders: orders}
	return json.Marshal(resp)
}

func queryInstrument(ctx sdk.Context, k *Keeper, path []string, req abci.RequestQuery) ([]byte, error) {
	// NOTE Provides a list of physical (ie notably not synthetic pairs) passive orders.
	if len(path) != 2 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "%s is not a valid query request path", req.Path)
	}

	source, destination := path[0], path[1]

	if !util.ValidateDenom(source) || !util.ValidateDenom(destination) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "Invalid denoms: %v %v", source, destination)
	}

	orders := make([]types.QueryOrderResponse, 0)

	idxStore := ctx.KVStore(k.keyIndices)
	key := types.GetPriorityKeyBySrcAndDst(source, destination)

	it := sdk.KVStorePrefixIterator(idxStore, key)
	defer it.Close()

	for it.Valid() {
		order := new(types.Order)
		k.cdc.MustUnmarshalBinaryBare(it.Value(), order)

		orders = append(orders, types.QueryOrderResponse{
			ID:              order.ID,
			Owner:           order.Owner,
			SourceRemaining: order.SourceRemaining.String(),
			Price:           order.Price(),
			Created:         order.Created,
		})

		it.Next()
	}

	resp := types.QueryInstrumentResponse{
		Source:      source,
		Destination: destination,
		Orders:      orders,
	}

	return json.Marshal(resp)
}

// getBestPrice returns the best priced passive order for source and
// destination instruments. Returns nil when executePlan cannot find a best
// plan.
func getBestPrice(ctx sdk.Context, k *Keeper, source, destination string) *sdk.Dec {
	var bestPrice *sdk.Dec

	bestPlan := k.createExecutionPlan(ctx, destination, source)
	if !bestPlan.DestinationCapacity().IsZero() {
		bestPrice = &bestPlan.Price
	}

	return bestPrice
}
