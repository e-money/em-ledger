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
		case types.MsgDecreaseCredit:
			return handleMsgDecreaseCredit(ctx, msg, k)
		case types.MsgRevokeLiquidityProvider:
			return handleMsgRevokeLiquidityProvider(ctx, msg, k)
		case types.MsgSetInflation:
			return handleMsgSetInflation(ctx, msg, k)
		default:
			errMsg := fmt.Sprintf("Unrecognized issuance Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSetInflation(ctx sdk.Context, msg types.MsgSetInflation, k keeper.Keeper) sdk.Result {
	error := k.SetInflationRate(ctx, msg.Issuer, msg.InflationRate, msg.Denom)
	if error != nil {
		return error.Result()
	}

	return sdk.Result{}
}

func handleMsgRevokeLiquidityProvider(ctx sdk.Context, msg types.MsgRevokeLiquidityProvider, k keeper.Keeper) sdk.Result {
	error := k.RevokeLiquidityProvider(ctx, msg.LiquidityProvider, msg.Issuer)
	if error != nil {
		return error.Result()
	}

	return sdk.Result{}
}

func handleMsgDecreaseCredit(ctx sdk.Context, msg types.MsgDecreaseCredit, k keeper.Keeper) sdk.Result {
	error := k.DecreaseCreditOfLiquidityProvider(ctx, msg.LiquidityProvider, msg.Issuer, msg.CreditDecrease)
	if error != nil {
		return error.Result()
	}

	return sdk.Result{}
}

func handleMsgIncreaseCredit(ctx sdk.Context, msg types.MsgIncreaseCredit, k keeper.Keeper) sdk.Result {
	k.IncreaseCreditOfLiquidityProvider(ctx, msg.LiquidityProvider, msg.Issuer, msg.CreditIncrease)
	return sdk.Result{}
}
