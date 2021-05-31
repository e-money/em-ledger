package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/market/types"
	"time"
)

var _ types.MsgServer = msgServer{}

type marketKeeper interface {
	NewOrderSingle(ctx sdk.Context, aggressiveOrder types.Order) (*sdk.Result, error)
	CancelOrder(ctx sdk.Context, owner sdk.AccAddress, clientOrderId string) (*sdk.Result, error)
	CancelReplaceLimitOrder(ctx sdk.Context, newOrder types.Order, origClientOrderId string) (*sdk.Result, error)
	GetSrcFromSlippage(ctx sdk.Context, srcDenom string, dst sdk.Coin, maxSlippage sdk.Dec) (sdk.Coin, error)
	OrderSpendGas(ctx sdk.Context, order *types.Order, origOrderCreated time.Time, orderGasMeter sdk.GasMeter,	callerErr *error)
}
type msgServer struct {
	k marketKeeper
}

func NewMsgServerImpl(keeper marketKeeper) types.MsgServer {
	return &msgServer{k: keeper}
}

func (m msgServer) AddLimitOrder(
	c context.Context, msg *types.MsgAddLimitOrder,
) (_ *types.MsgAddLimitOrderResponse, err error) {
	var (
		order types.Order
	)

	ctx := sdk.UnwrapSDKContext(c)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "owner")
	}

	order, err = types.NewOrder(
		ctx.BlockTime(), msg.TimeInForce, msg.Source, msg.Destination, owner,
		msg.ClientOrderId,
	)

	if err != nil {
		return nil, err
	}

	result, err := m.k.NewOrderSingle(ctx, order)
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

	slippageSource, err := m.k.GetSrcFromSlippage(
		ctx, msg.Source, msg.Destination, msg.MaxSlippage,
	)
	if err != nil {
		return nil, err
	}

	limitMsg := &types.MsgAddLimitOrder{
		Owner:         msg.Owner,
		ClientOrderId: msg.ClientOrderId,
		TimeInForce:   msg.TimeInForce,
		Source:        slippageSource,
		Destination:   msg.Destination,
	}

	_, err = m.AddLimitOrder(c, limitMsg)
	return &types.MsgAddMarketOrderResponse{}, err
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

	if msg.Destination.Amount.LTE(sdk.ZeroInt()) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidPrice, "Destination %s price is Zero or less: %s", msg.Destination.Denom, msg.Destination.Amount)
	}

	slippageSource, err := m.k.GetSrcFromSlippage(
		ctx, msg.Source, msg.Destination, msg.MaxSlippage,
	)
	if err != nil {
		return nil, err
	}

	limitMsg := &types.MsgCancelReplaceLimitOrder{
		Owner:             msg.Owner,
		OrigClientOrderId: msg.OrigClientOrderId,
		NewClientOrderId:  msg.NewClientOrderId,
		TimeInForce:       msg.TimeInForce,
		Source:            slippageSource,
		Destination:       msg.Destination,
	}

	_, err = m.CancelReplaceLimitOrder(c, limitMsg)
	return &types.MsgCancelReplaceMarketOrderResponse{}, err
}
