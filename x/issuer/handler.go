// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package issuer

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/issuer/keeper"
	"github.com/e-money/em-ledger/x/issuer/types"
)

// TODO Accept Keeper argument
func newHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgIncreaseMintable:
			return handleMsgIncreaseMintableAmount(ctx, msg, k)
		case types.MsgDecreaseMintable:
			return handleMsgDecreaseMintableAmount(ctx, msg, k)
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
	return k.SetInflationRate(ctx, msg.Issuer, msg.InflationRate, msg.Denom)
}

func handleMsgRevokeLiquidityProvider(ctx sdk.Context, msg types.MsgRevokeLiquidityProvider, k keeper.Keeper) sdk.Result {
	return k.RevokeLiquidityProvider(ctx, msg.LiquidityProvider, msg.Issuer)
}

func handleMsgDecreaseMintableAmount(ctx sdk.Context, msg types.MsgDecreaseMintable, k keeper.Keeper) sdk.Result {
	return k.DecreaseMintableAmountOfLiquidityProvider(ctx, msg.LiquidityProvider, msg.Issuer, msg.MintableDecrease)
}

func handleMsgIncreaseMintableAmount(ctx sdk.Context, msg types.MsgIncreaseMintable, k keeper.Keeper) sdk.Result {
	return k.IncreaseMintableAmountOfLiquidityProvider(ctx, msg.LiquidityProvider, msg.Issuer, msg.MintableIncrease)
}
