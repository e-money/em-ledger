package v09

import (
	"sort"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// Migrate converts the slashing module state of the v0.9.x series to the genesis state of v1.0.x series
// Partially forked from Cosmos-SDK v0.42.9 : x/slashing/legacy/v040/migrate.go
func Migrate(genState GenesisState) *slashingtypes.GenesisState {
	// Note that the two following `for` loop over a map's keys, so are not
	// deterministic.
	newSigningInfos := make([]slashingtypes.SigningInfo, 0, len(genState.SigningInfos))
	for address, signingInfo := range genState.SigningInfos {
		newSigningInfos = append(newSigningInfos, slashingtypes.SigningInfo{
			Address: address,
			ValidatorSigningInfo: slashingtypes.ValidatorSigningInfo{
				Address: signingInfo.Address.String(),
				// StartHeight:         signingInfo.StartHeight,
				// IndexOffset:         signingInfo.IndexOffset,
				JailedUntil: signingInfo.JailedUntil,
				Tombstoned:  signingInfo.Tombstoned,
				// MissedBlocksCounter: signingInfo.MissedBlocksCounter,
			},
		})
	}

	sort.Slice(newSigningInfos, func(i, j int) bool { return newSigningInfos[i].Address < newSigningInfos[j].Address })

	return &slashingtypes.GenesisState{
		Params: slashingtypes.Params{
			SignedBlocksWindow:      genState.Params.SignedBlocksWindowDuration.Nanoseconds(),
			MinSignedPerWindow:      genState.Params.MinSignedPerWindow,
			DowntimeJailDuration:    genState.Params.DowntimeJailDuration,
			SlashFractionDoubleSign: genState.Params.SlashFractionDoubleSign,
			SlashFractionDowntime:   genState.Params.SlashFractionDowntime,
		},
		SigningInfos: newSigningInfos,
		MissedBlocks: []slashingtypes.ValidatorMissedBlocks{}, // Ignore blocks missed on previous net
	}
}
