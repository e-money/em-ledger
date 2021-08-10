package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/stargate/networks/emibctokens/x/swap/types"
)

// SetIbcToken set a specific ibcToken in the store from its index
func (k Keeper) SetIbcToken(ctx sdk.Context, ibcToken types.IbcToken) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.IbcTokenKeyPrefix))
	b := k.cdc.MustMarshalBinaryBare(&ibcToken)
	store.Set(types.IbcTokenKey(
		ibcToken.Index,
	), b)
}

// GetIbcToken returns a ibcToken from its index
func (k Keeper) GetIbcToken(
	ctx sdk.Context,
	index string,

) (val types.IbcToken, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.IbcTokenKeyPrefix))

	b := store.Get(types.IbcTokenKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshalBinaryBare(b, &val)
	return val, true
}

// RemoveIbcToken removes a ibcToken from the store
func (k Keeper) RemoveIbcToken(
	ctx sdk.Context,
	index string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.IbcTokenKeyPrefix))
	store.Delete(types.IbcTokenKey(
		index,
	))
}

// GetAllIbcToken returns all ibcToken
func (k Keeper) GetAllIbcToken(ctx sdk.Context) (list []types.IbcToken) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.IbcTokenKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.IbcToken
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
