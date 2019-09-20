package issuer

import (
	"emoney/x/issuer/keeper"
	types "emoney/x/issuer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type genesisState struct {
	Issuers []types.Issuer
}

func initGenesis(ctx sdk.Context, k keeper.Keeper, state genesisState) {

}

func defaultGenesisState() genesisState {
	issuer, _ := sdk.AccAddressFromBech32("emoney127teu2esvmqhhcn5hnh29eq7ndh7f3etnsww7v")
	return genesisState{
		Issuers: []types.Issuer{
			{
				Address: issuer,
				Denoms:  []string{"x2eur", "x0jpy"},
			},
		},
	}
}
