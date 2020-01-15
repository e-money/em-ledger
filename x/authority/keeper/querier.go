package keeper

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/authority/types"
	abci "github.com/tendermint/tendermint/abci/types"
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

type queryGasPricesResponse struct {
	MinGasPrices string `json:"min_gas_prices" yaml:"min_gas_prices"`
}

func queryGasPrices(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	gasPrices := k.GetGasPrices(ctx)

	response := queryGasPricesResponse{gasPrices.String()}
	bz, err := json.Marshal(response)
	if err != nil {
		return []byte{}, sdk.ErrInternal(err.Error())
	}

	return bz, nil
}
