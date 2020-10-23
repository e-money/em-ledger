// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/e-money/em-ledger/x/slashing/types"
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
		keeper.SetValidatorSigningInfo(ctx, address, info)
	}

	// We omit information about blocks missed in previous net.
	//for addr, array := range data.MissedBlocks {
	//	address, err := sdk.ConsAddressFromBech32(addr)
	//	if err != nil {
	//		panic(err)
	//	}
	//	for _, missed := range array {
	//		keeper.setValidatorMissedBlockBitArray(address, missed.Index, missed.Missed)
	//	}
	//}

	keeper.paramspace.SetParamSet(ctx, &data.Params)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	signingInfos := make(map[string]types.ValidatorSigningInfo)

	keeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		bechAddr := address.String()
		signingInfos[bechAddr] = info
		return false
	})

	return types.GenesisState{
		Params:       keeper.GetParams(ctx),
		SigningInfos: signingInfos,
		MissedBlocks: nil,
	}
}
