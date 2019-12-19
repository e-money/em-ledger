// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	QueryInstruments = "instruments"
)

func NewQuerier(k *Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryInstruments:
			return queryInstruments(ctx, k)

		default:
			return nil, sdk.ErrUnknownRequest("unknown bank query endpoint")
		}
	}
}

type queryInstrumentsResponse struct {
	Pair       string
	OrderCount int
}

func queryInstruments(ctx sdk.Context, k *Keeper) ([]byte, sdk.Error) {
	response := make([]queryInstrumentsResponse, len(k.instruments))
	for i, v := range k.instruments {
		response[i] = queryInstrumentsResponse{
			Pair:       fmt.Sprintf("%v/%v", v.Source, v.Destination),
			OrderCount: v.Orders.Size(),
		}
	}

	// Wrap the instruments in an object.
	var instrumentsWrapper = struct{ Instruments []queryInstrumentsResponse }{response}

	bz, err := json.Marshal(instrumentsWrapper)
	if err != nil {
		return []byte{}, sdk.ErrInternal(err.Error())
	}

	return bz, nil
}
