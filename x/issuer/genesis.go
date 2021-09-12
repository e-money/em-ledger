// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package issuer

import (
	authtypes "github.com/e-money/em-ledger/x/authority/types"
	"github.com/e-money/em-ledger/x/issuer/keeper"
	types "github.com/e-money/em-ledger/x/issuer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func initGenesis(ctx sdk.Context, k keeper.Keeper, state types.GenesisState) {
	for _, issuer := range state.Issuers {
		denomMetadata := make([]authtypes.Denomination, len(issuer.Denoms))
		for i, denom := range issuer.Denoms {
			denomMetadata[i].Base = denom
		}
		k.AddIssuer(ctx, issuer, denomMetadata)
	}
}

func defaultGenesisState() *types.GenesisState {
	return &types.GenesisState{}
}
