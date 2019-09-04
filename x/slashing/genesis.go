package slashing

import (
	"emoney/x/slashing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper Keeper, stakingKeeper types.StakingKeeper, data types.GenesisState) {
	stakingKeeper.IterateValidators(ctx,
		func(index int64, validator exported.ValidatorI) bool {
			keeper.addPubkey(ctx, validator.GetConsPubKey())
			return false
		},
	)

	for addr, info := range data.SigningInfos {
		address, err := sdk.ConsAddressFromBech32(addr)
		if err != nil {
			panic(err)
		}
		keeper.SetValidatorSigningInfo(address, info)
	}

	for addr, array := range data.MissedBlocks {
		address, err := sdk.ConsAddressFromBech32(addr)
		if err != nil {
			panic(err)
		}
		for _, missed := range array {
			keeper.setValidatorMissedBlockBitArray(address, missed.Index, missed.Missed)
		}
	}

	keeper.paramspace.SetParamSet(ctx, &data.Params)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	return types.GenesisState{}
}
