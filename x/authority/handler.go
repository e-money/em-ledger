// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package authority

import (
	"fmt"

	"github.com/e-money/em-ledger/x/authority/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func newHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (result sdk.Result) {
		defer func() {
			if r := recover(); r != nil {
				switch o := r.(type) {
				case sdk.Result:
					result = o
				case sdk.Error:
					result = o.Result()
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
