// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/market/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(k *Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryInstruments:
			return queryInstruments(ctx, k)
		case types.QueryInstrument:
			return queryInstrument(ctx, k, path[1:], req)
		case types.QueryByAccount:
			return queryByAccount(ctx, k, path[1:], req)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unrecognized market query endpoint")
		}
	}
}

type QueryInstrumentResponse struct {
	Source      string               `json:"source" yaml:"source"`
	Destination string               `json:"destination" yaml:"destination"`
	Orders      []QueryOrderResponse `json:"orders" yaml:"orders"`
}

func (q QueryInstrumentResponse) String() string {
	sb := new(strings.Builder)

	sb.WriteString(fmt.Sprintf("%v => %v\n", q.Source, q.Destination))

	for _, order := range q.Orders {
		sb.WriteString(order.String())
	}

	return sb.String()
}

type QueryByAccountResponse struct {
	Orders OrderResponses `json:"orders" yaml:"orders"`
}

func (q QueryByAccountResponse) String() string {
	sb := new(strings.Builder)
	for _, order := range q.Orders {
		sb.WriteString(order.String())
	}

	return sb.String()
}

type QueryOrderResponse struct {
	ID uint64 `json:"id" yaml:"id"`

	Owner           sdk.AccAddress `json:"owner" yaml:"owner"`
	SourceRemaining string         `json:"source_remaining" yaml:"source_remaining"`

	ClientOrderId *string `json:"client_order_id,omitempty" yaml:"client_order_id,omitempty"`

	Price sdk.Dec `json:"price" yaml:"price"`
}

func (q QueryOrderResponse) String() string {
	return fmt.Sprintf(" - %v %v %v %v\n", q.ID, q.Price, q.SourceRemaining, q.Owner.String())
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

	// TODO Determine suitable ordering or leave undefined
	// sort.Sort(orders)

	resp := QueryByAccountResponse{orders}
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

	orders := make([]QueryOrderResponse, 0)

	idxStore := ctx.KVStore(k.keyIndices)
	key := types.GetPriorityKeyBySrcAndDst(source, destination)

	it := sdk.KVStorePrefixIterator(idxStore, key)
	defer it.Close()

	for it.Valid() {
		order := new(types.Order)
		k.cdc.MustUnmarshalBinaryBare(it.Value(), order)

		orders = append(orders, QueryOrderResponse{
			ID:              order.ID,
			Owner:           order.Owner,
			SourceRemaining: order.SourceRemaining.String(),
			Price:           order.Price(),
		})

		it.Next()
	}

	resp := QueryInstrumentResponse{
		Source:      source,
		Destination: destination,
		Orders:      orders,
	}

	return json.Marshal(resp)
}

type QueryInstrumentsWrapperResponse struct {
	Instruments []QueryInstrumentsResponse `json:"instruments" yaml:"instruments"`
}

func (q QueryInstrumentsWrapperResponse) String() string {
	sb := new(strings.Builder)
	for _, instrument := range q.Instruments {
		sb.WriteString(instrument.String())
	}

	return sb.String()
}

type QueryInstrumentsResponse struct {
	Source      string     `json:"source" yaml:"source"`
	Destination string     `json:"destination" yaml:"destination"`
	BestPrice   *sdk.Dec   `json:"best_price,omitempty" yaml:"best_price,omitempty"`
	LastPrice   *sdk.Dec   `json:"last_price,omitempty" yaml:"last_price,omitempty"`
	LastTraded  *time.Time `json:"last_traded,omitempty" yaml:"last_traded,omitempty"`
}

//
func (q QueryInstrumentsResponse) String() string {
	return fmt.Sprintf("%v => %v", q.Source, q.Destination)
}

func queryInstruments(ctx sdk.Context, k *Keeper) ([]byte, error) {
	instruments := k.GetAllInstruments(ctx)

	response := make([]QueryInstrumentsResponse, len(instruments))
	for i, v := range instruments {
		response[i] = QueryInstrumentsResponse{
			Source:      v.Source,
			Destination: v.Destination,
			LastPrice:   v.LastPrice,
			BestPrice:   v.BestPrice,
			LastTraded:  v.Timestamp,
		}
	}

	// Wrap the instruments in an object in anticipation of later expansion
	instrumentsWrapper := QueryInstrumentsWrapperResponse{response}
	return json.Marshal(instrumentsWrapper)
}
