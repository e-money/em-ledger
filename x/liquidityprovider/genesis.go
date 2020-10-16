// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package liquidityprovider

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GenesisAcc struct {
	Account  sdk.AccAddress `json:"address" yaml:"address"`
	Mintable sdk.Coins      `json:"mintable" yaml:"mintable"`
}

type genesisState struct {
	Accounts []GenesisAcc `json:"accounts" yaml:"accounts"`
}

func defaultGenesisState() genesisState {
	return genesisState{}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, gs genesisState) {
	for _, lp := range gs.Accounts {
		keeper.CreateLiquidityProvider(ctx, lp.Account, lp.Mintable)
	}
}
