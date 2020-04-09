// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/market/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"strings"
)

func NewQuerier(k *Keeper) sdk.Querier {
	return func(_ sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryInstruments:
			return queryInstruments(k)
		case types.QueryInstrument:
			return queryInstrument(k, path[1:], req)
		case types.QueryByAccount:
			return queryByAccount(k, path[1:], req)
		default:
			return nil, sdk.ErrUnknownRequest("unknown market query endpoint")
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
	ID      uint64    `json:"id" yaml:"id"`
	Created time.Time `json:"created" yaml:"created"`

	Owner           sdk.AccAddress `json:"owner" yaml:"owner"`
	SourceRemaining string         `json:"source_remaining" yaml:"source_remaining"`

	ClientOrderId *string `json:"client_order_id,omitempty" yaml:"client_order_id,omitempty"`

	Price sdk.Dec `json:"price" yaml:"price"`
}

func (q QueryOrderResponse) String() string {
	return fmt.Sprintf(" - %v %v %v %v %v\n", q.ID, q.Price, q.SourceRemaining, q.Created, q.Owner.String())
}

type OrderResponses []types.Order

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

func queryByAccount(k *Keeper, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
	if len(path) != 1 {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
	}

	account, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdk.ErrInvalidAddress(fmt.Sprint("Address could not be parsed", path[0], err))
	}

	o := k.accountOrders.GetAllOrders(account)
	orders := make(OrderResponses, 0)

	it := o.Iterator()
	for it.Next() {
		order := it.Value().(*types.Order)
		orders = append(orders, *order)
	}

	sort.Sort(orders)

	resp := QueryByAccountResponse{orders}
	bz, err := json.Marshal(resp)
	if err != nil {
		return []byte{}, sdk.ErrInternal(err.Error())
	}

	return bz, nil
}

func queryInstrument(k *Keeper, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
	if len(path) != 2 {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
	}

	source, destination := path[0], path[1]

	if !util.ValidateDenom(source) || !util.ValidateDenom(destination) {
		return nil, sdk.ErrInvalidCoins(fmt.Sprintf("Invalid denoms: %v %v", source, destination))
	}

	instrument := k.instruments.GetInstrument(source, destination)

	orders := make([]QueryOrderResponse, 0)
	if instrument != nil {
		it := instrument.Orders.Iterator()
		for it.Next() {
			order := it.Key().(*types.Order)

			orders = append(orders, QueryOrderResponse{
				ID:              order.ID,
				Created:         order.Created,
				Owner:           order.Owner,
				SourceRemaining: order.SourceRemaining.String(),
				Price:           order.Price(),
			})
		}
	}

	resp := QueryInstrumentResponse{
		Source:      source,
		Destination: destination,
		Orders:      orders,
	}

	bz, err := json.Marshal(resp)
	if err != nil {
		return []byte{}, sdk.ErrInternal(err.Error())
	}

	return bz, nil
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
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	OrderCount  int    `json:"order_count" yaml:"order_count"`
}

func (q QueryInstrumentsResponse) String() string {
	return fmt.Sprintf("%v => %v (%v)", q.Source, q.Destination, q.OrderCount)
}

func queryInstruments(k *Keeper) ([]byte, sdk.Error) {
	response := make([]QueryInstrumentsResponse, len(k.instruments))
	for i, v := range k.instruments {
		response[i] = QueryInstrumentsResponse{
			Source:      v.Source,
			Destination: v.Destination,
			OrderCount:  v.Orders.Size(),
		}
	}

	// Wrap the instruments in an object in anticipation of later expansion
	instrumentsWrapper := QueryInstrumentsWrapperResponse{response}
	bz, err := json.Marshal(instrumentsWrapper)
	if err != nil {
		return []byte{}, sdk.ErrInternal(err.Error())
	}

	return bz, nil
}
