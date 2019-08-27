package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// Stored by *validator* address (not operator address)
func (k Keeper) getValidatorSigningInfo(_ sdk.Context, address sdk.ConsAddress) (info types.ValidatorSigningInfo, found bool) {
	bz := k.signedBlocks.Get(types.GetValidatorSigningInfoKey(address))
	if bz == nil {
		found = false
		return
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)
	found = true
	return
}

// Stored by *validator* address (not operator address)
func (k Keeper) IterateValidatorSigningInfos(_ sdk.Context,
	handler func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool)) {

	//store := ctx.KVStore(k.storeKey)
	//iter := sdk.KVStorePrefixIterator(store, types.ValidatorSigningInfoKey)
	iter := k.signedBlocks.Iterator(types.ValidatorSigningInfoKey, sdk.PrefixEndBytes(types.ValidatorSigningInfoKey))

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		address := types.GetValidatorSigningInfoAddress(iter.Key())
		var info types.ValidatorSigningInfo
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &info)
		if handler(address, info) {
			break
		}
	}
}

// Stored by *validator* address (not operator address)
func (k Keeper) SetValidatorSigningInfo(_ sdk.Context, address sdk.ConsAddress, info types.ValidatorSigningInfo) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	k.signedBlocks.Set(types.GetValidatorSigningInfoKey(address), bz)
}

// Stored by *validator* address (not operator address)
func (k Keeper) getValidatorMissedBlockBitArray(_ sdk.Context, address sdk.ConsAddress, index int64) (missed bool) {
	bz := k.signedBlocks.Get(types.GetValidatorMissedBlockBitArrayKey(address, index))
	if bz == nil {
		// lazy: treat empty key as not missed
		missed = false
		return
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &missed)
	return
}

// Stored by *validator* address (not operator address)
func (k Keeper) IterateValidatorMissedBlockBitArray(ctx sdk.Context,
	address sdk.ConsAddress, handler func(index int64, missed bool) (stop bool)) {

	index := int64(0)
	// Array may be sparse
	signedBlockWindow := k.SignedBlocksWindow(ctx)
	for ; index < signedBlockWindow; index++ {
		var missed bool
		bz := k.signedBlocks.Get(types.GetValidatorMissedBlockBitArrayKey(address, index))
		if bz == nil {
			continue
		}
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &missed)
		if handler(index, missed) {
			break
		}
	}
}

// Stored by *validator* address (not operator address)
func (k Keeper) setValidatorMissedBlockBitArray(_ sdk.Context, address sdk.ConsAddress, index int64, missed bool) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(missed)
	k.signedBlocks.Set(types.GetValidatorMissedBlockBitArrayKey(address, index), bz)
}

// Stored by *validator* address (not operator address)
func (k Keeper) clearValidatorMissedBlockBitArray(_ sdk.Context, address sdk.ConsAddress) {
	//store := ctx.KVStore(k.storeKey)
	//iter := sdk.KVStorePrefixIterator(store, types.GetValidatorMissedBlockBitArrayPrefixKey(address))
	// TODO Verify conversion
	iter := k.signedBlocks.Iterator(types.GetValidatorMissedBlockBitArrayPrefixKey(address), sdk.PrefixEndBytes(address))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		k.signedBlocks.Delete(iter.Key())
	}
}
