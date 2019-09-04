package slashing

import (
	"emoney/x/slashing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Stored by *validator* address (not operator address)
func (k Keeper) getValidatorSigningInfo(address sdk.ConsAddress) (info types.ValidatorSigningInfo, found bool) {
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
func (k Keeper) IterateValidatorSigningInfos(
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
func (k Keeper) SetValidatorSigningInfo(address sdk.ConsAddress, info types.ValidatorSigningInfo) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	k.signedBlocks.Set(types.GetValidatorSigningInfoKey(address), bz)
}

// Stored by *validator* address (not operator address)
func (k Keeper) getValidatorMissedBlockBitArray(address sdk.ConsAddress, index int64) (missed bool) {
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
func (k Keeper) setValidatorMissedBlockBitArray(address sdk.ConsAddress, index int64, missed bool) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(missed)
	k.signedBlocks.Set(types.GetValidatorMissedBlockBitArrayKey(address, index), bz)
}

// Stored by *validator* address (not operator address)
func (k Keeper) clearValidatorMissedBlockBitArray(address sdk.ConsAddress) {
	iter := k.signedBlocks.Iterator(types.GetValidatorMissedBlockBitArrayPrefixKey(address), sdk.PrefixEndBytes(address))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		k.signedBlocks.Delete(iter.Key())
	}
}
