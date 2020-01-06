// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/market/types"
	abci "github.com/tendermint/tendermint/abci/types"
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
	Source      string        `json:"source" yaml:"source"`
	Destination string        `json:"destination" yaml:"destination"`
	Orders      []types.Order `json:"orders" yaml:"orders"`
}

func queryInstrument(ctx sdk.Context, k *Keeper, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
	if len(path) != 2 {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
	}

	source, destination := path[0], path[1]

	instrument := k.instruments.GetInstrument(source, destination)
	if instrument == nil {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("Could not find instrument %s/%s", source, destination))
	}

	orders := make([]types.Order, 0)
	it := instrument.Orders.Iterator()
	for it.Next() {
		order := it.Key().(*types.Order)
		orders = append(orders, *order)
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
