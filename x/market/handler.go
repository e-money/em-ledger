// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package market

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/market/types"

	"github.com/e-money/em-ledger/x/market/keeper"
)

func NewHandler(k *keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgAddOrder:
			return handleMsgAddOrder(ctx, k, msg)

		case types.MsgCancelOrder:
			return handleMsgCancelOrder(ctx, k, msg)

		case types.MsgCancelReplaceOrder:
			return handleMsgCancelReplaceOrder(ctx, k, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized market message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgAddOrder(ctx sdk.Context, k *Keeper, msg types.MsgAddOrder) sdk.Result {
	order, err := types.NewOrder(msg.Source, msg.Destination, msg.Owner, msg.ClientOrderId)
	if err != nil {
		return err.Result()
	}

	// TODO Emit events.
	return k.NewOrderSingle(ctx, order)
}

func handleMsgCancelOrder(ctx sdk.Context, k *Keeper, msg types.MsgCancelOrder) sdk.Result {
	// TODO Emit events.
	return k.CancelOrder(ctx, msg.Owner, msg.ClientOrderId)
}

func handleMsgCancelReplaceOrder(ctx sdk.Context, k *Keeper, msg types.MsgCancelReplaceOrder) sdk.Result {
	// TODO Emit events.
	order, err := types.NewOrder(msg.Source, msg.Destination, msg.Owner, msg.NewClientOrderId)
	if err != nil {
		return err.Result()
	}
	return k.CancelReplaceOrder(ctx, order, msg.OrigClientOrderId)
}
