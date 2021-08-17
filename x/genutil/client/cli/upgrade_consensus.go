package cli

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"time"
)

// upgradeModuleParams modifies parameters between emoney-2 and emoney-3.
func upgradeModuleParams(cdc codec.JSONMarshaler, appState types.AppMap) {
	increaseValidatorSet(cdc, appState)
	changeDowntimeJailing(cdc, appState)
}

func changeDowntimeJailing(cdc codec.JSONMarshaler, appState types.AppMap) {
	slashingGenesis := appState[slashing.ModuleName]
	var genesis slashing.GenesisState
	cdc.MustUnmarshalJSON(slashingGenesis, &genesis)

	genesis.Params.SignedBlocksWindow = (24 * time.Hour).Nanoseconds()

	delete(appState, slashing.ModuleName)
	appState[slashing.ModuleName] = cdc.MustMarshalJSON(&genesis)
}

func increaseValidatorSet(cdc codec.JSONMarshaler, appState types.AppMap) {
	stakingGenesis := appState[stakingtypes.ModuleName]
	var genesis stakingtypes.GenesisState
	cdc.MustUnmarshalJSON(stakingGenesis, &genesis)

	genesis.Params.MaxValidators = 100

	delete(appState, stakingtypes.ModuleName)
	appState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(&genesis)
}
