package liquidityprovider

import (
	"emoney/x/liquidityprovider/types"
)

type genesisState struct {
	Accounts []types.LiquidityProviderAccount
}

func defaultGenesisState() genesisState {
	return genesisState{}
}

//
//func InitGenesis(_ *sdk.Context,  am.keeper, gs genesisState) {
//
//}
