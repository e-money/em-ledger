# Market Module

## Overview

The market module enable accounts to sell an amount of *source* tokens in exchange for an fixed amount of *destination* tokens, where the price is derived from either a specified amount of *source* and *destination* tokens or from market data on recent trades.

This is a generalisation of the classic limit order for two-sided markets:

| Side | Base | Term | Price |
|------|------|------|-------|
| Buy  | Destination Denom | Source Denom | Source Amount / Destination Amount |
| Sell | Source Denom | Destination Denom | Destination Amount / Source Amount |

As the *destination* amount is fixed, less than *source* amount of tokens will be paid if a better price exist in the market. Having the *destination* amount fixed is useful for payments where a fixed amount of foreign currency needs to be delivered.

## Features

*No instrument listing required*. Any token is immediately tradeable against other tokens.

*No execution fees*. This applies for both makers and takers, which only need to pay the standard transaction costs.

*Optimized for liquidity*. Orders do not touch the account balance until they are matched, so that makers can place multiple orders based on the same *Source*.
When the balance of the owner account changes, SourceRemaining is adjusted accordingly and any untradable orders are canceled. 

*Takers always trade at the best price*. In case there is a better price in the market, price improvement is passed to the taker who pays less than the specified amount of *Source* tokens.

*Arbitrage-free*. Sophisticated order matching ensures that no arbitrage opportunities exist in the market. Orders always trade at the best price by considering synthetic instruments, e.g. a single USD->EUR order matched against EUR->GBP and GBP->USD simultaneously.

*Price/time priority matching*. Orders at the same price will be ordered by OrderId, with the lowest matched first.  

*Immediate settlement*. Matched orders are settled immediately with finality.

## Transaction Types

The transaction types mirror those of the [FIX trading specification](https://www.fixtrading.org/online-specification/business-area-trade/) for single order handling.

The market module currently supports limit and market orders with multiple "time in force" values.

### Time in force

Each order includes a "time in force" value to control its behaviour. 
 
 | Time in force | Behaviour |
 |------|------|
 | Good til cancel  | Aggresively match the order against the book. Add the remainder passively to the book, if the order is not filled.  | 
 | Immediate or cancel | Aggresively match the order against the book. The remainder of the order is canceled. | 
 | Fill or kill | Aggresively match the *entire* order against the book. If this does not succeed, cancel the entire order. |

### Limit orders

A limit order specifies the worst (limit) price to trade at calculated as: Price = Destination Amount / Source Amount.

### Market orders

A market order is internally treated as a limit order. Upon order receipt the price is determined using the last traded price of its instrument, with a slippage value applied to determine the worst (limit) price that the order will trade at.

A market order will be rejected in case the instrument has not been traded yet. 

## Order Data

An order consists of the following data:

* Owner: a `AccAddress` which will be used for settlement and order modifications.
* OrderId: a `uint64` assigned by the market module, monotonically increasing.
* TimeInForce: an enumeration that determines order matching behaviour.
* ClientOrderId: a `string` assigned by owner, which must not be a duplicate of an existing order.
* Source: a `Coin` representing the desired amount of tokens to sell.
* SourceFilled: `Int` that tracks the sold amount so far.
* SourceRemaining: a `Int` that is adjusted with *SourceFilled* and if the owner account balance change.
* Destination: a `Coin` representing the minimum amount of tokens to buy.
* DestinationFilled: `Int` that tracks the bought amount so far.
* Price: a `Dec` calculated as *Destination* / *Source*.
