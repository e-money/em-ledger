// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/e-money/em-ledger/x/issuer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryIssuers:
			return listIssuers(ctx, k)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown issuer query endpoint: %s", path[0])
		}
	}
}

func listIssuers(ctx sdk.Context, k Keeper) ([]byte, error) {
	issuers := k.GetIssuers(ctx)
	return types.ModuleCdc.MarshalJSON(issuers)
}
