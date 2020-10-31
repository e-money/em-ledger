# Events

The market module emits the following events:

## Order Accepted

| Type   | Attribute Key   | Attribute Value     |
| -------| --------------- | ------------------- |
| market | action          | "accept"            |
| market | order_id        | {uniqueOrderId}     |
| market | owner           | {ownerAddress}      |
| market | client_order_id | {clientOrderId}     |
| market | source          | {sourceAmount}      |
| market | destination     | {destinationAmount} |

This event reports the *initial* state of an order when it is accepted by the market module.

The limit price an can be calculated as:
```
limit_price = destination / source
```

## Order Expired

| Type   | Attribute Key      | Attribute Value           |
| ------ | ------------------ | ------------------------- |
| market | action             | "expire"                  |
| market | order_id           | {uniqueOrderId}           |
| market | owner              | {ownerAddress}            |
| market | client_order_id    | {clientOrderId}           |
| market | source             | {sourceAmount}            |
| market | source_remaining   | {sourceRemainingAmount}   |
| market | source_filled      | {sourceFilledAmount}      |
| market | destination        | {destinationAmount}       |
| market | destination_filled | {destinationFilledAmount} |

This event reports the *final* state of an order before it is expired by the market module.

An order expires when
1. It is completely filled or
2. It is canceled by the user or
3. The owner account has an insufficient balance to execute the order.

Both `source_filled` and `destination_filled` are cumulative and can be used to calculate the average fill price:
```
average_fill_price = destination_filled / source_filled
```

## Order Filled

| Type   | Attribute Key      | Attribute Value           |
| ------ | ------------------ | ------------------------- |
| market | action             | "fill"                    |
| market | order_id           | {uniqueOrderId}           |
| market | owner              | {ownerAddress}            |
| market | client_order_id    | {clientOrderId}           |
| market | source_filled      | {sourceFilledAmount}      |
| market | destination_filled | {destinationFilledAmount} |

When the market module executes a trade, the orders on each side of the trade receive a fill event.

Both `source_filled` and `destination_filled` are specific to a single trade, i.e. in contrast to the [Order Expired](#order-expired) event they are non-cumulative.

The fill price is calculated as:
```
fill_price = destination_filled / source_filled
```

## Order Updated

| Type   | Attribute Key    | Attribute Value           |
| ------ | -----------------| ------------------------- |
| market | action           | "update"                  |
| market | order_id         | {uniqueOrderId}           |
| market | owner            | {ownerAddress}            |
| market | client_order_id  | {clientOrderId}           |
| market | source_remaining | {sourceRemainingAmount}   |

This event reports any updates to the state of an order that affects `source_remaining`. This might happen if the `owner` account balance changes for the source denomination.

## Handlers

### MsgAddLimitOrder

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| message  | module        | "market"           |
| message  | action        | "add_limit_order"  |
| message  | sender        | {senderAddress}    |

### MsgAddMarketOrder

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| message  | module        | "market"           |
| message  | action        | "add_market_order" |
| message  | sender        | {senderAddress}    |

### MsgCancelOrder

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| message  | module        | "market"           |
| message  | action        | "cancel_order"     |
| message  | sender        | {senderAddress}    |

### MsgCancelReplaceLimitOrder

| Type     | Attribute Key | Attribute Value              |
| -------- | ------------- | ---------------------------- |
| message  | module        | "market"                     |
| message  | action        | "cancel_replace_limit_order" |
| message  | sender        | {senderAddress}              |
