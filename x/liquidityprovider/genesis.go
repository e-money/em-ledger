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
