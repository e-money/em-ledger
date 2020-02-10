// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/market/types"
	"sync"
)

func BeginBlocker(ctx sdk.Context, sk *Keeper) {
	if ctx.BlockHeight() == 25652 {
		sk.accountOrders = types.NewOrders()
		sk.instruments = types.Instruments{}
		sk.appstateInit = new(sync.Once)
	}

	sk.initializeFromStore(ctx)
}
