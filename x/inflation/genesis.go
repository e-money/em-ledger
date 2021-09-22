// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package inflation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/inflation/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(state InflationState) types.GenesisState {
	return types.GenesisState{
		InflationState: state,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *types.GenesisState {
	return &types.GenesisState{
		InflationState: DefaultInflationState(),
	}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {
	keeper.SetState(ctx, data.InflationState)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	state := keeper.GetState(ctx)
	return NewGenesisState(state)
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data types.GenesisState) error {
	err := types.ValidateInflationState(data.InflationState)
	if err != nil {
		return err
	}

	return nil
}
