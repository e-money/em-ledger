// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package authority

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/e-money/em-ledger/x/authority/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func newHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (result *sdk.Result, err error) {
		//defer func() {
		//	if r := recover(); r != nil {
		//		switch o := r.(type) {
		//		case sdk.Result:
		//			result = o
		//		case sdk.Error:
		//			result = o.Result()
		//		default:
		//			panic(r)
		//		}
		//	}
		//}()
		//
		//if err := msg.ValidateBasic(); err != nil {
		//	return sdk.ErrUnknownRequest(err.Error()).Result()
		//}

		switch msg := msg.(type) {
		case types.MsgCreateIssuer:
			return keeper.CreateIssuer(ctx, msg.Authority, msg.Issuer, msg.Denominations)
		case types.MsgDestroyIssuer:
			return keeper.DestroyIssuer(ctx, msg.Authority, msg.Issuer)
		case types.MsgSetGasPrices:
			return keeper.SetGasPrices(ctx, msg.Authority, msg.GasPrices)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}
