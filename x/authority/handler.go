package authority

import (
	"fmt"

	"emoney/x/authority/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func newHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (result sdk.Result) {
		defer func() {
			if r := recover(); r != nil {
				switch o := r.(type) {
				case sdk.Result:
					result = o
				default:
					panic(r)
				}
			}
		}()

		switch msg := msg.(type) {
		case types.MsgCreateIssuer:
			return keeper.CreateIssuer(ctx, msg.Authority, msg.Issuer, msg.Denominations)
		case types.MsgDestroyIssuer:
			return keeper.DestroyIssuer(ctx, msg.Authority, msg.Issuer)
		default:
			errMsg := fmt.Sprintf("Unrecognized inflation Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
