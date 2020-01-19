// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/market/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"time"
)

const (
	QueryInstruments = "instruments"
	QueryInstrument  = "instrument"
)

func NewQuerier(k *Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryInstruments:
			return queryInstruments(ctx, k)
		case QueryInstrument:
			return queryInstrument(ctx, k, path[1:], req)
		default:
			return nil, sdk.ErrUnknownRequest("unknown market query endpoint")
		}
	}
}

type queryInstrumentResponse struct {
	Source      string               `json:"source" yaml:"source"`
	Destination string               `json:"destination" yaml:"destination"`
	Orders      []queryOrderResponse `json:"orders" yaml:"orders"`
}

type queryOrderResponse struct {
	ID      uint64    `json:"id" yaml:"id"`
	Created time.Time `json:"created" yaml:"created"`

	Owner     sdk.AccAddress `json:"owner" yaml:"owner"`
	Remaining string         `json:"remaining" yaml:"remaining"`

	Price sdk.Dec
}

func queryInstrument(ctx sdk.Context, k *Keeper, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
	if len(path) != 2 {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
	}

	source, destination := path[0], path[1]

	if !util.ValidateDenom(source) || !util.ValidateDenom(destination) {
		return nil, sdk.ErrInvalidCoins(fmt.Sprintf("Invalid denoms: %v %v", source, destination))
	}

	instrument := k.instruments.GetInstrument(source, destination)

	orders := make([]queryOrderResponse, 0)
	if instrument != nil {
		it := instrument.Orders.Iterator()
		for it.Next() {
			order := it.Key().(*types.Order)
			orders = append(orders, queryOrderResponse{
				ID:        order.ID,
				Created:   order.Created,
				Owner:     order.Owner,
				Remaining: fmt.Sprintf("%v%v", order.SourceRemaining, order.Source.Denom),
				Price:     order.Price(),
			})
		}
	}

	resp := queryInstrumentResponse{
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

type queryInstrumentsResponse struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	OrderCount  int    `json:"order_count" yaml:"order_count"`
}

func queryInstruments(ctx sdk.Context, k *Keeper) ([]byte, sdk.Error) {
	response := make([]queryInstrumentsResponse, len(k.instruments))
	for i, v := range k.instruments {
		response[i] = queryInstrumentsResponse{
			Source:      v.Source,
			Destination: v.Destination,
			OrderCount:  v.Orders.Size(),
		}
	}

	// Wrap the instruments in an object in anticipation of later expansion
	var instrumentsWrapper = struct {
		Instruments []queryInstrumentsResponse `json:"instruments" yaml:"instruments"`
	}{response}

	bz, err := json.Marshal(instrumentsWrapper)
	if err != nil {
		return []byte{}, sdk.ErrInternal(err.Error())
	}

	return bz, nil
}
