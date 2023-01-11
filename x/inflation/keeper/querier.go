package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/inflation/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier returns an inflation Querier handler.
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, _ abci.RequestQuery) ([]byte, error) {
		switch path[0] {

		case types.QueryInflation:
			return queryInflation(ctx, k, legacyQuerierCdc)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized inflation query endpoint: %s", path[0])
		}
	}
}

func queryInflation(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	inflationState := k.GetState(ctx)

	// TODO Introduce a more presentation-friendly response type
	return legacyQuerierCdc.MarshalJSON(inflationState)
}
