package authority

import sdk "github.com/cosmos/cosmos-sdk/types"

type GenesisState struct {
	AuthorityKey sdk.AccAddress `json:"key" yaml:"key"`
}

func NewGenesisState(authorityKey sdk.AccAddress) GenesisState {
	return GenesisState{
		AuthorityKey: authorityKey,
	}
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, state GenesisState) {
	keeper.SetAuthority(ctx, state.AuthorityKey)
}
