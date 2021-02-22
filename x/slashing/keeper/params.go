// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

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
}

func (k Keeper) MinSignedPerWindow(ctx sdk.Context) (res sdk.Dec) {
	k.paramspace.Get(ctx, types.KeyMinSignedPerWindow, &res)
	return
}
