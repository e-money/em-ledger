// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package liquidityprovider

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
)

type genesisState struct {
	Accounts []types.LiquidityProviderAccount `json:"accounts" yaml:"accounts"`
}

func defaultGenesisState() genesisState {
	return genesisState{}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, gs genesisState) {
	for _, lp := range gs.Accounts {
		keeper.CreateLiquidityProvider(ctx, lp.GetAddress(), lp.Mintable)
	}
}
