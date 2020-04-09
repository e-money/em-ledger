// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/authority/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"strings"
)

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryGasPrices:
			return queryGasPrices(ctx, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown authority query endpoint")
		}
	}

}

type QueryGasPricesResponse struct {
	MinGasPrices []sdk.DecCoin `json:"min_gas_prices" yaml:"min_gas_prices"`
}

func (q QueryGasPricesResponse) String() string {
	sb := new(strings.Builder)
	sb.WriteString("Minimum gas prices\n")
	for _, gp := range q.MinGasPrices {
		sb.WriteString(fmt.Sprintf(" - %v : %v\n", gp.Denom, gp.Amount.String()))
	}

	return sb.String()
}

func queryGasPrices(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	gasPrices := k.GetGasPrices(ctx)

	response := QueryGasPricesResponse{gasPrices}

	bz, err := json.Marshal(response)
	if err != nil {
		return []byte{}, sdk.ErrInternal(err.Error())
	}

	return bz, nil
}
