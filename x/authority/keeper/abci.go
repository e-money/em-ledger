package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func BeginBlocker(ctx sdk.Context, k Keeper) {
	k.initGasPrices(ctx)
}
