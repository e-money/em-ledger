package issuer

import (
	"fmt"

	"emoney/x/issuer/keeper"
	"emoney/x/issuer/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO Accept Keeper argument
func newHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgIncreaseCredit:
			return handleMsgIncreaseCredit(ctx, msg, k)
		default:
			errMsg := fmt.Sprintf("Unrecognized issuance Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgIncreaseCredit(ctx sdk.Context, msg types.MsgIncreaseCredit, k keeper.Keeper) sdk.Result {
	k.IncreaseCreditOfLiquidityProvider(ctx, msg.LiquidityProvider, msg.Issuer, msg.CreditIncrease)
	return sdk.Result{}
}
