package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func BeginBlocker(ctx sdk.Context, sk *Keeper) {
	sk.initializeFromStore(ctx)
}
