// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/inflation/internal/types"
)

// NewQuerier returns an inflation Querier handler.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, _ abci.RequestQuery) ([]byte, error) {
		switch path[0] {

		case types.QueryInflation:
			return queryInflation(ctx, k)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized inflation query endpoint: %s", path[0])
		}
	}
}

func queryInflation(ctx sdk.Context, k Keeper) ([]byte, error) {
	inflationState := k.GetState(ctx)

	// TODO Introduce a more presentation-friendly response type
	return types.ModuleCdc.MarshalJSON(inflationState)
}
