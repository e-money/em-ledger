## Market Module

The market module enable accounts to sell an amount of *source* tokens in exchange for an amount of *destination* tokens.

This is a generalisation of the classic limit order for two-sided markets:

| Side | Base | Term | Price |
|------|------|------|-------|
| Buy  | Destination Denom | Source Denom | Source Amount / Destination Amount |
| Sell | Source Denom | Destination Denom | Destination Amount / Source Amount |

### Order Data

An order consists of the following data:

* Owner: a `AccAddress` which will be used for settlement and order modifications.
* OrderId: a `uint64` assigned by the market module, monotonically increasing.
* ClientOrderId: a `string` assigned by owner, which must not be a duplicate of an existing order.
* Source: a `Coin` representing the desired amount of tokens to sell.
* Destination: a `Coin` representing the minimum amount of tokens to buy.
* SourceFilled: `Coin` that tracks the sold amount so far.
* SourceRemaining: a `Coin` that is the minimum of *SourceFilled* and owner account balance.
* Price: a `Dec` calculated as *Destination* / *Source*.

### Features

*No instrument listing required*. Any token is immediately tradeable against other tokens.

*No execution fees*. This applies for both makers and takers, which only need to pay the standard transaction costs.

*Optimized for liquidity*. Orders do not touch the account balance until they are matched, so that makers can place multiple orders based on the same *Source*.
When the balance of the owner account changes, SourceRemaining is adjusted accordingly and any untradable orders are canceled. 

*Takers always trade at the best price*. Price improvement is passed to the taker in case there is a better price in the market.

*Arbitrage-free*. Sophisticated matching logic ensures that no arbitrage opportunities exist in the market. Orders always trade at the best price, with the possibility of matching against multiple orders such as EUR->GBP and GBP->USD for a single EUR->USD order.

*Price/time priority matching*. Orders at the same price will be ordered by OrderId, with the lowest matched first.  

*Immediate settlement*. Matched orders are settled immediately with finality.

### Transaction Types

The transaction types mirror those of the [FIX trading specification](https://www.fixtrading.org/online-specification/business-area-trade/) for single order handling.

#### NewOrderSingle
Adds a new order to the order book. The ClientOrderId must be unique among existing orders for the same owner account.

#### CancelOrder
Cancels the remaining part of an existing order, referenced by it's ClientOrderId. In case the order has already been fully filled, an error will be returned. 

#### CancelReplaceOrder
Cancels the remaining part of an existing order, referenced by it's ClientOrderId. The filled part of the cancelled order is carried over into the new order.
