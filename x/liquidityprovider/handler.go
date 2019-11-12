package liquidityprovider

import (
	"fmt"

	"emoney/x/liquidityprovider/keeper"
	types "emoney/x/liquidityprovider/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO Accept Keeper argument
func newHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgMintTokens:
			return k.MintTokens(ctx, msg.LiquidityProvider, msg.Amount)
		case types.MsgBurnTokens:
			return k.BurnTokensFromBalance(ctx, msg.LiquidityProvider, msg.Amount)
		default:
			errMsg := fmt.Sprintf("Unrecognized issuance Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
