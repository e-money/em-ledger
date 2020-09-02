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

Two order types are currently supported:

### Limit orders

A limit order specifies a fixed price for trading an instrument. 

### Market orders

The price of a a market order bases is based on the last traded price of its instrument. A slippage value can be included to specify the largest deviation the order may take from the market price.
 
### Time in force

Each order can include a "time in force" value to control its behaviour. 
 
 | Time in force | Behaviour |
 |------|------|
 | Good til cancel  | Attempt to match the order against the book. Add the remainder to the book, if the order is not filled.  | 
 | Immediate or cancel | Match as much of the order as possible against the book. Do not add the remainder of the order to the book. | 
 | Fill or kill | Attempt to match the entire order. If it does not succeed, do not execute any part of the order |


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
