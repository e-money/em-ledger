// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package authority

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/authority/keeper"

	"github.com/e-money/em-ledger/x/authority/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func newHandler(k Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (result *sdk.Result, err error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgCreateIssuer:
			res, err := msgServer.CreateIssuer(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgDestroyIssuer:
			res, err := msgServer.DestroyIssuer(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgSetGasPrices:
			res, err := msgServer.SetGasPrices(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}
