// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package issuer

import (
	"github.com/e-money/em-ledger/x/issuer/keeper"
	types "github.com/e-money/em-ledger/x/issuer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type genesisState struct {
	Issuers []types.Issuer `json:"issuers" yaml:"issuers"`
}

func initGenesis(ctx sdk.Context, k keeper.Keeper, state genesisState) {
	for _, issuer := range state.Issuers {
		k.AddIssuer(ctx, issuer)
	}
}

func defaultGenesisState() genesisState {
	return genesisState{}
}
