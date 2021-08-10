package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/stargate/networks/emibctokens/x/swap/types"
)

// SetDenomTrace set a specific denomTrace in the store from its index
func (k Keeper) SetDenomTrace(ctx sdk.Context, denomTrace types.DenomTrace) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DenomTraceKeyPrefix))
	b := k.cdc.MustMarshalBinaryBare(&denomTrace)
	store.Set(types.DenomTraceKey(
		denomTrace.Index,
	), b)
}

// GetDenomTrace returns a denomTrace from its index
func (k Keeper) GetDenomTrace(
	ctx sdk.Context,
	index string,

) (val types.DenomTrace, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DenomTraceKeyPrefix))

	b := store.Get(types.DenomTraceKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshalBinaryBare(b, &val)
	return val, true
}

// RemoveDenomTrace removes a denomTrace from the store
func (k Keeper) RemoveDenomTrace(
	ctx sdk.Context,
	index string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DenomTraceKeyPrefix))
	store.Delete(types.DenomTraceKey(
		index,
	))
}

// GetAllDenomTrace returns all denomTrace
func (k Keeper) GetAllDenomTrace(ctx sdk.Context) (list []types.DenomTrace) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DenomTraceKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.DenomTrace
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
