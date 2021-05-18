package buyback

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/e-money/em-ledger/x/buyback/internal/types"
)

func NewGenesisState(interval time.Duration) *types.GenesisState {
	return &types.GenesisState{
		Interval: interval.String(),
	}
}

func defaultGenesisState() *types.GenesisState {
	return &types.GenesisState{
		Interval: time.Hour.String(),
	}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, state types.GenesisState) error {
	updateInterval, err := time.ParseDuration(state.Interval)
	if err != nil {
		return err
	}

	keeper.SetUpdateInterval(ctx, updateInterval)
	return nil
}
