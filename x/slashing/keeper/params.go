package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/slashing/types"
)

// SignedBlocksWindowDuration - sliding window for downtime slashing
func (k Keeper) SignedBlocksWindowDuration(ctx sdk.Context) time.Duration {
	var x int64
	k.paramspace.Get(ctx, types.KeySignedBlocksWindow, &x)
	return time.Duration(x) * time.Nanosecond // multiplication only for doc. Duration is ns
	// todo (reviewer): is int64 ok?
}

func (k Keeper) MinSignedPerWindow(ctx sdk.Context) (res sdk.Dec) {
	k.paramspace.Get(ctx, types.KeyMinSignedPerWindow, &res)
	return
}
