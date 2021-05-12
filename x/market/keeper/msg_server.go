package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/market/types"
)

var _ types.MsgServer = msgServer{}

type marketKeeper interface {
	NewMarketOrderWithSlippage(ctx sdk.Context, srcDenom string, dst sdk.Coin, maxSlippage sdk.Dec, owner sdk.AccAddress, timeInForce types.TimeInForce, clientOrderId string) (*sdk.Result, error)
	NewOrderSingle(ctx sdk.Context, aggressiveOrder types.Order, messageType types.TxMessageType) (*sdk.Result, error)
	CancelOrder(ctx sdk.Context, owner sdk.AccAddress, clientOrderId string) (*sdk.Result, error)
	CancelReplaceLimitOrder(ctx sdk.Context, newOrder types.Order, origClientOrderId string) (*sdk.Result, error)
	CancelReplaceMarketOrder(ctx sdk.Context, msg *types.MsgCancelReplaceMarketOrder) (*sdk.Result, error)
}
type msgServer struct {
	k marketKeeper
}

func NewMsgServerImpl(keeper marketKeeper) types.MsgServer {
	return &msgServer{k: keeper}
}

func (m msgServer) AddLimitOrder(c context.Context, msg *types.MsgAddLimitOrder) (*types.MsgAddLimitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "owner")
	}

	order, err := types.NewOrder(ctx.BlockTime(), msg.TimeInForce, msg.Source, msg.Destination, owner, msg.ClientOrderId)
	if err != nil {
		return nil, err
	}

	result, err := m.k.NewOrderSingle(ctx, order, )
	if err != nil {
		return nil, err
	}

	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgAddLimitOrderResponse{}, nil
}

func (m msgServer) AddMarketOrder(c context.Context, msg *types.MsgAddMarketOrder) (*types.MsgAddMarketOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "owner")
	}
	result, err := m.k.NewMarketOrderWithSlippage(ctx, msg.Source, msg.Destination, msg.MaxSlippage, owner, msg.TimeInForce, msg.ClientOrderId)
	if err != nil {
		return nil, err
	}
	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}

	return &types.MsgAddMarketOrderResponse{}, nil
}

func (m msgServer) CancelOrder(c context.Context, msg *types.MsgCancelOrder) (*types.MsgCancelOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "owner")
	}

	result, err := m.k.CancelOrder(ctx, owner, msg.ClientOrderId)
	if err != nil {
		return nil, err
	}
	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgCancelOrderResponse{}, nil
}

func (m msgServer) CancelReplaceLimitOrder(c context.Context, msg *types.MsgCancelReplaceLimitOrder) (*types.MsgCancelReplaceLimitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "owner")
	}
	order, err := types.NewOrder(ctx.BlockTime(), msg.TimeInForce, msg.Source, msg.Destination, owner, msg.NewClientOrderId)
	if err != nil {
		return nil, err
	}

	result, err := m.k.CancelReplaceLimitOrder(ctx, order, msg.OrigClientOrderId)
	if err != nil {
		return nil, err
	}
	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgCancelReplaceLimitOrderResponse{}, nil
}

func (m msgServer) CancelReplaceMarketOrder(c context.Context, msg *types.MsgCancelReplaceMarketOrder) (*types.MsgCancelReplaceMarketOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid owner value:%s", msg.Owner))
	}
	if msg.Destination.Amount.LTE(sdk.ZeroInt()) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidPrice, "Destination %s price is Zero or less: %s", msg.Destination.Denom, msg.Destination.Amount)
	}

	if msg.TimeInForce <= types.TimeInForce_Unspecified || msg.TimeInForce > types.TimeInForce_FillOrKill {
		return nil, sdkerrors.Wrapf(types.ErrUnknownTimeInForce, "Invalid Time In Force: %d", msg.TimeInForce)
	}

	result, err := m.k.CancelReplaceMarketOrder(ctx, msg)
	if err != nil {
		return nil, err
	}
	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}

	return &types.MsgCancelReplaceMarketOrderResponse{}, nil
}
