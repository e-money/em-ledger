package market

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/e-money/em-ledger/x/market/keeper"
)

func NewHandler(k *keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {

		default:
			errMsg := fmt.Sprintf("unrecognized market message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
