# Messages

The market module supports limit and market orders with multiple "Time In Force" values:

 | Time In Force | Behaviour |
 |---------------|-----------|
 | GTC           | Good 'Til Cancel: Aggresively match the order against the book. Add the remainder passively to the book, if the order is not filled. |
 | IOC           | Immediate Or Cancel: Aggresively match the order against the book. The remainder of the order is canceled. |
 | FOK           | Fill Or Kill: Aggresively match the *entire* order against the book. If this does not succeed, cancel the entire order. |

The `ClientOrderId` is supplied by the order owner (sender) and must be unique among all active orders for the owner. It is used when canceling or replacing an active order.

## MsgAddLimitOrder

A limit order specifies the limit (worst) price to trade at. When the order is filled it might be filled at a better price (receive "price improvement").

The limit price is calculated as: `Price = Destination / Source`.

```go
// MsgAddLimitOrder represents a message to add a limit order.
MsgAddLimitOrder struct {
  Owner         sdk.AccAddress `json:"owner" yaml:"owner"`
  ClientOrderId string         `json:"client_order_id" yaml:"client_order_id"`
  TimeInForce   string         `json:"time_in_force" yaml:"time_in_force"`
  Source        sdk.Coin       `json:"source" yaml:"source"`
  Destination   sdk.Coin       `json:"destination" yaml:"destination"`
}
```

## MsgAddMarketOrder

Market orders are converted to limit orders on receipt: The limit price is determined using the last traded price of its instrument, with a slippage value applied to determine the limit price.

A market order will be rejected in case the instrument has not been traded yet.

```go
// MsgAddMarketOrder represents a message to add a market order.
MsgAddMarketOrder struct {
  Owner         sdk.AccAddress `json:"owner" yaml:"owner"`
  ClientOrderId string         `json:"client_order_id" yaml:"client_order_id"`
  TimeInForce   string         `json:"time_in_force" yaml:"time_in_force"`
  Source        string         `json:"source" yaml:"source"`
  Destination   sdk.Coin       `json:"destination" yaml:"destination"`
  MaxSlippage   sdk.Dec        `json:"maximum_slippage" yaml:"maximum_slippage"`
}
```

## MsgCancelOrder

The unfilled part of an active order can be canceled using MsgCancelOrder:

```go
// MsgCancelOrder represents a message to cancel an existing order.
MsgCancelOrder struct {
  Owner         sdk.AccAddress `json:"owner" yaml:"owner"`
  ClientOrderId string         `json:"client_order_id" yaml:"client_order_id"`
}
```

## MsgCancelReplaceLimitOrder

The MsgCancelReplaceLimitOrder message is useful for liquidity providers (market makers) who wish to adjust their prices while remaining in the market.

```go
// MsgCancelReplaceLimitOrder represents a message to cancel an existing order and replace it with a limit order.
MsgCancelReplaceLimitOrder struct {
  Owner             sdk.AccAddress `json:"owner" yaml:"owner"`
  OrigClientOrderId string         `json:"original_client_order_id" yaml:"original_client_order_id"`
  NewClientOrderId  string         `json:"new_client_order_id" yaml:"new_client_order_id"`
  TimeInForce       string         `json:"time_in_force" yaml:"time_in_force"`
  Source            sdk.Coin       `json:"source" yaml:"source"`
  Destination       sdk.Coin       `json:"destination" yaml:"destination"`
}
```

The unfilled part of the original order is canceled and replaced with a new limit order, taking into consideration how much of the original order was filled:

```go
// Has the previous order already achieved the goal on the source side?
if origOrder.SourceFilled.GTE(newOrder.Source.Amount) {
  return nil, sdkerrors.Wrap(types.ErrNoSourceRemaining, "")
}

[...]

// Adjust remaining according to how much of the replaced order was filled:
newOrder.SourceFilled = origOrder.SourceFilled
newOrder.SourceRemaining = newOrder.Source.Amount.Sub(newOrder.SourceFilled)
newOrder.DestinationFilled = origOrder.DestinationFilled
```

## MsgCancelReplaceMarketOrder

The MsgCancelReplaceMarketOrder message is helpful to adjust prices and slippage for previous market orders while 
remaining in the market. Please note, The new Market order is converted to limit order on receipt: The limit price is 
determined using the last traded price of its instrument, with the *adjusted* slippage value applied resulting in the 
updated limit price.

```go
// MsgCancelReplaceMarketOrder represents a message to cancel an existing order and replace it with a market order.
MsgCancelReplaceMarketOrder struct {
  Owner             sdk.AccAddress `json:"owner" yaml:"owner"`
  OrigClientOrderId string         `json:"original_client_order_id" yaml:"original_client_order_id"`
  NewClientOrderId  string         `json:"new_client_order_id" yaml:"new_client_order_id"`
  TimeInForce       string         `json:"time_in_force" yaml:"time_in_force"`
  Source            string         `protobuf:"bytes,5,opt,name=source,proto3" json:"source,omitempty" yaml:"source"`
  Destination       sdk.Coin       `json:"destination" yaml:"destination"`
  MaxSlippage       sdk.Dec        `protobuf:"bytes,7,opt,name=maximum_slippage,json=maximumSlippage,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"maximum_slippage" yaml:"maximum_slippage"`
}
```

The unfilled part of the original order is canceled and replaced with a new market order, taking into consideration how 
much of the original order was filled:

```go
// Has the previous order already achieved the goal on the source side?
if origOrder.SourceFilled.GTE(newOrder.Source.Amount) {
  return nil, sdkerrors.Wrap(types.ErrNoSourceRemaining, "")
}
if origOrder.DestinationFilled.GTE(msg.Destination.Amount) {
    return nil, sdkerrors.Wrap(
        types.ErrNoSourceRemaining, fmt.Sprintf("has already been filled filled:%s >= %s",
        	origOrder.Destination.Amount.String(), msg.Destination.Amount.String()),
    )
}
[...]

// Adjust remaining according to how much of the replaced order was filled:
destinationFilled := origOrder.DestinationFilled
dstRemaining := msg.Destination.Amount.Sub(destinationFilled)
remDstCoin := sdk.NewCoin(msg.Destination.Denom, dstRemaining)
```
