package issuance

import (
	"emoney/x/issuance/types"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO Accept Keeper argument
func newHandler() sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgMintTokens:
			return handleMsgMintTokens(ctx, msg)
		//case MsgSetName:
		//	return handleMsgSetName(ctx, keeper, msg)
		//case MsgBuyName:
		//	return handleMsgBuyName(ctx, keeper, msg)
		//case MsgDeleteName:
		//	return handleMsgDeleteName(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized issuance Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgMintTokens(ctx sdk.Context, msg types.MsgMintTokens) sdk.Result {
	fmt.Println(" *** Mint token handler invoked!")
	return sdk.Result{}
}
