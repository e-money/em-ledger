# Events

The market module emits the following events:

## New Order

| Type        | Attribute Key   | Attribute Value     |
| ----------- | --------------- | ------------------- |
| market_new  | order_id        | {uniqueOrderId}     |
| market_new  | owner           | {ownerAddress}      |
| market_new  | client_order_id | {clientOrderId}     |
| market_new  | source          | {sourceAmount}      |
| market_new  | destination     | {destinationAmount} |

## Cancel Order

| Type           | Attribute Key   | Attribute Value    |
| -------------- | --------------- | ------------------ |
| market_cancel  | order_id        | {uniqueOrderId}    |
| market_cancel  | owner           | {ownerAddress}     |
| market_cancel  | client_order_id | {clientOrderId}    |

## Order Fill

| Type         | Attribute Key      | Attribute Value           |
| ------------ | ------------------ | ------------------------- |
| market_fill  | order_id           | {uniqueOrderId}           |
| market_fill  | owner              | {ownerAddress}            |
| market_fill  | client_order_id    | {clientOrderId}           |
| market_fill  | partial_fill       | {boolean}                 |
| market_fill  | source             | {sourceAmount}            |
| market_fill  | source_remaining   | {sourceRemainingAmount}   |
| market_fill  | source_filled      | {sourceFilledAmount}      |
| market_fill  | destination        | {destinationAmount}       |
| market_fill  | destination_filled | {destinationFilledAmount} |

While `partial_fill` is true the order is not fully filled and remains active.

Both `source_filled` and `destination_filled` are cumulative.

The average fill price for the order can be calculated as `averageFillPrice = destinationFilledAmount / sourceFilledAmount`.

## Handlers

### MsgAddLimitOrder

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| message  | module        | market             |
| message  | action        | add_limit_order    |
| message  | sender        | {senderAddress}    |

### MsgAddMarketOrder

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| message  | module        | market             |
| message  | action        | add_market_order   |
| message  | sender        | {senderAddress}    |

### MsgCancelOrder

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| message  | module        | market             |
| message  | action        | cancel_order       |
| message  | sender        | {senderAddress}    |

### MsgCancelReplaceLimitOrder

| Type     | Attribute Key | Attribute Value            |
| -------- | ------------- | -------------------------- |
| message  | module        | market                     |
| message  | action        | cancel_replace_limit_order |
| message  | sender        | {senderAddress}            |
