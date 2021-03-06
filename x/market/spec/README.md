# Market Module ("DEX")

## Abstract

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

*Arbitrage-free*. Sophisticated order matching ensures that no arbitrage opportunities exist in the market. Orders always trade at the best price by considering synthetic instruments, e.g. a single eUSD->eEUR order matched against eEUR->eGBP and eGBP->eUSD simultaneously.

*Price/time priority matching*. Orders at the same price will be ordered by OrderId, with the lowest matched first.  

*Immediate settlement*. Matched orders are settled immediately with finality.

## Contents

1. **[State](01_state.md)**
2. **[Messages](02_messages.md)**
    - [MsgAddLimitOrder](02_messages.md#MsgAddLimitOrder)
    - [MsgAddMarketOrder](02_messages.md#MsgAddMarketOrder)
    - [MsgCancelOrder](02_messages.md#MsgCancelOrder)
    - [MsgCancelReplaceLimitOrder](02_messages.md#MsgCancelReplaceLimitOrder)
3. **[Events](03_events.md)**
    - [Order Accepted](03_events.md#order-accepted)
    - [Order Expired](03_events.md#order-expired)
    - [Order Filled](03_events.md#order-filled)
    - [Order Updated](03_events.md#order-updated)
    - [Handlers](03_events.md#Handlers)
4. **[Queries](04_queries.md)**
