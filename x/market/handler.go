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
		case types.MsgAddLimitOrder:
			return handleMsgAddLimitOrder(ctx, k, msg)

		case types.MsgAddMarketOrder:
			return handleMsgAddMarketOrder(ctx, k, msg)

		case types.MsgCancelOrder:
			return handleMsgCancelOrder(ctx, k, msg)

		case types.MsgCancelReplaceLimitOrder:
			return handleMsgCancelReplaceLimitOrder(ctx, k, msg)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized market message type: %T", msg)
		}
	}
}

func handleMsgAddMarketOrder(ctx sdk.Context, k *keeper.Keeper, msg types.MsgAddMarketOrder) (*sdk.Result, error) {
	timeInForce, err := types.TimeInForceFromString(msg.TimeInForce)
	if err != nil {
		return nil, err
	}

	return k.NewMarketOrderWithSlippage(ctx, msg.Source, msg.Destination, msg.MaxSlippage, msg.Owner, timeInForce, msg.ClientOrderId)
}

func handleMsgAddLimitOrder(ctx sdk.Context, k *Keeper, msg types.MsgAddLimitOrder) (*sdk.Result, error) {
	timeInForce, err := types.TimeInForceFromString(msg.TimeInForce)
	if err != nil {
		return nil, err
	}

	order, err := types.NewOrder(timeInForce, msg.Source, msg.Destination, msg.Owner, msg.ClientOrderId)
	if err != nil {
		return nil, err
	}

	return k.NewOrderSingle(ctx, order)
}

func handleMsgCancelOrder(ctx sdk.Context, k *Keeper, msg types.MsgCancelOrder) (*sdk.Result, error) {
	return k.CancelOrder(ctx, msg.Owner, msg.ClientOrderId)
}

func handleMsgCancelReplaceLimitOrder(ctx sdk.Context, k *Keeper, msg types.MsgCancelReplaceLimitOrder) (*sdk.Result, error) {
	order, err := types.NewOrder(TimeInForce_GoodTilCancel, msg.Source, msg.Destination, msg.Owner, msg.NewClientOrderId)
	if err != nil {
		return nil, err
	}

	return k.CancelReplaceOrder(ctx, order, msg.OrigClientOrderId)
}
