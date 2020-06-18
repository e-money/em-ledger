// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package market

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/market/types"

	"github.com/e-money/em-ledger/x/market/keeper"
)

func NewHandler(k *keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgAddOrder:
			return handleMsgAddOrder(ctx, k, msg)

		case types.MsgCancelOrder:
			return handleMsgCancelOrder(ctx, k, msg)

		case types.MsgCancelReplaceOrder:
			return handleMsgCancelReplaceOrder(ctx, k, msg)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized market message type: %T", msg)
		}
	}
}

func handleMsgAddOrder(ctx sdk.Context, k *Keeper, msg types.MsgAddOrder) (*sdk.Result, error) {
	order, err := types.NewOrder(msg.Source, msg.Destination, msg.Owner, ctx.BlockTime(), msg.ClientOrderId)
	if err != nil {
		return nil, err
	}

	// TODO Emit events.
	return k.NewOrderSingle(ctx, order)
}

func handleMsgCancelOrder(ctx sdk.Context, k *Keeper, msg types.MsgCancelOrder) (*sdk.Result, error) {
	// TODO Emit events.
	return k.CancelOrder(ctx, msg.Owner, msg.ClientOrderId)
}

func handleMsgCancelReplaceOrder(ctx sdk.Context, k *Keeper, msg types.MsgCancelReplaceOrder) (*sdk.Result, error) {
	// TODO Emit events.
	order, err := types.NewOrder(msg.Source, msg.Destination, msg.Owner, ctx.BlockTime(), msg.NewClientOrderId)
	if err != nil {
		return nil, err
	}
	return k.CancelReplaceOrder(ctx, order, msg.OrigClientOrderId)
}
