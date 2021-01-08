// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/slashing/types"
	db "github.com/tendermint/tm-db"
)

// Stored by *validator* address (not operator address)
func (k Keeper) getValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (info types.ValidatorSigningInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValidatorSigningInfoKey(address))
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
	ctx sdk.Context, handler func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorSigningInfoKey)

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
func (k Keeper) SetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress, info types.ValidatorSigningInfo) {
	// NOTE Information regarding missed blocks have been moved away from IAVL.
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	store.Set(types.GetValidatorSigningInfoKey(address), bz)
}

// Stored by *validator* address (not operator address)
func (k Keeper) getValidatorMissedBlockBitArray(address sdk.ConsAddress, index int64) (missed bool) {
	bz, err := k.database.Get(types.GetValidatorMissedBlockBitArrayKey(address, index))
	if err != nil {
		panic(err) // TODO Better handling
	}

	if bz == nil {
		// lazy: treat empty key as not missed
		missed = false
		return
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &missed)
	return
}

// Stored by *validator* address (not operator address)
func (k Keeper) setValidatorMissedBlockBitArray(batch db.Batch, address sdk.ConsAddress, index int64, missed bool) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(missed)
	batch.Set(types.GetValidatorMissedBlockBitArrayKey(address, index), bz)
}

// Stored by *validator* address (not operator address)
func (k Keeper) clearValidatorMissedBlockBitArray(batch db.Batch, address sdk.ConsAddress) {
	iter, err := k.database.Iterator(types.GetValidatorMissedBlockBitArrayPrefixKey(address), sdk.PrefixEndBytes(address))
	if err != nil {
		panic(err) // TODO Better handling
	}

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		batch.Delete(iter.Key())
	}
}
