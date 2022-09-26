package cli

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/e-money/em-ledger/x/buyback"
)

// upgradeModuleParams modifies parameters between emoney-2 and emoney-3.
func upgradeModuleParams(cdc codec.JSONCodec, appState types.AppMap) {
	increaseValidatorSet(cdc, appState)
	changeDowntimeJailing(cdc, appState)
	updateBuybackInterval(cdc, appState)
}

func updateBuybackInterval(cdc codec.JSONCodec, appState types.AppMap) {
	buybackGenesis := appState[buyback.ModuleName]
	var genesis buyback.GenesisState
	cdc.MustUnmarshalJSON(buybackGenesis, &genesis)

	genesis.Interval = (24 * time.Hour).String()

	delete(appState, buyback.ModuleName)
	appState[buyback.ModuleName] = cdc.MustMarshalJSON(&genesis)
}

func changeDowntimeJailing(cdc codec.JSONCodec, appState types.AppMap) {
	slashingGenesis := appState[slashing.ModuleName]
	var genesis slashing.GenesisState
	cdc.MustUnmarshalJSON(slashingGenesis, &genesis)

	genesis.Params.SignedBlocksWindow = (24 * time.Hour).Nanoseconds()

	delete(appState, slashing.ModuleName)
	appState[slashing.ModuleName] = cdc.MustMarshalJSON(&genesis)
}

func increaseValidatorSet(cdc codec.JSONCodec, appState types.AppMap) {
	stakingGenesis := appState[stakingtypes.ModuleName]
	var genesis stakingtypes.GenesisState
	cdc.MustUnmarshalJSON(stakingGenesis, &genesis)

	genesis.Params.MaxValidators = 100

	delete(appState, stakingtypes.ModuleName)
	appState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(&genesis)
}
