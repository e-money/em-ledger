package authority

type GenesisState struct{}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}
