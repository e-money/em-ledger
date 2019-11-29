// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package liquidityprovider

import (
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
)

type genesisState struct {
	Accounts []types.LiquidityProviderAccount `json:"accounts" yaml:"accounts"`
}

func defaultGenesisState() genesisState {
	return genesisState{}
}

//
//func InitGenesis(_ *sdk.Context,  am.keeper, gs genesisState) {
//
//}
