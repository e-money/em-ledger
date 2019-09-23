package inflation

import (
	"emoney/x/inflation/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GenesisState struct {
	InflationState InflationState `json:"assets" yaml:"assets"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(state InflationState) GenesisState {
	return GenesisState{
		InflationState: state,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return GenesisState{
		InflationState: DefaultInflationState(),
	}
}

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetState(ctx, data.InflationState)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	state := keeper.GetState(ctx)
	return NewGenesisState(state)
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	err := types.ValidateInflationState(data.InflationState)
	if err != nil {
		return err
	}

	return nil
}
