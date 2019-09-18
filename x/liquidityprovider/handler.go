package liquidityprovider

import (
	"fmt"

	"emoney/x/liquidityprovider/keeper"
	types "emoney/x/liquidityprovider/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// TODO Remove
	defaultCredit = sdk.NewCoins(
		sdk.NewCoin("x2eur", sdk.NewIntWithDecimal(50000, 2)),
	)
)

// TODO Accept Keeper argument
func newHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgDevTracerBullet:
			return handleMsgDevTracerBullet(ctx, msg, k)
		case types.MsgMintTokens:
			return handleMsgMintTokens(ctx, msg, k)
		default:
			errMsg := fmt.Sprintf("Unrecognized issuance Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgMintTokens(ctx sdk.Context, msg types.MsgMintTokens, k keeper.Keeper) sdk.Result {
	fmt.Println(" *** Minting tokens handler")
	k.MintTokensFromCredit(ctx, msg.LiquidityProvider, msg.Amount)
	return sdk.Result{}
}

func handleMsgDevTracerBullet(ctx sdk.Context, msg types.MsgDevTracerBullet, k keeper.Keeper) sdk.Result {
	k.CreateLiquidityProvider(ctx, msg.Sender, defaultCredit)
	return sdk.Result{}
}
