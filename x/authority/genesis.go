package authority

import sdk "github.com/cosmos/cosmos-sdk/types"

type GenesisState struct {
	Authority sdk.AccAddress
}

func DefaultGenesisState() GenesisState {
	// TODO Remove
	authority, _ := sdk.AccAddressFromBech32("emoney127teu2esvmqhhcn5hnh29eq7ndh7f3etnsww7v")
	return GenesisState{
		Authority: authority,
	}
}
