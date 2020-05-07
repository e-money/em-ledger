# Market Module

## Overview

The market module enable accounts to sell an amount of *source* tokens in exchange for an fixed amount of *destination* tokens, where the price is derived from the amount of *source* and *destination* tokens.

This is a generalisation of the classic limit order for two-sided markets:

| Side | Base | Term | Price |
|------|------|------|-------|
| Buy  | Destination Denom | Source Denom | Source Amount / Destination Amount |
| Sell | Source Denom | Destination Denom | Destination Amount / Source Amount |

As the *destination* amount is fixed, less than *source* amount of tokens will be paid if a better price exist in the market.

Having the *destination* amount fixed is useful in the payments space, where a fixed amount of foreign currency needs to be delivered.

## Features

The market module provides some compelling features, such as:

*No instrument listing required*. Any token is immediately tradeable against other tokens.

*No execution fees*. This applies for both makers and takers, which only need to pay the standard transaction costs.

*Optimized for liquidity*. Orders do not touch the account balance until they are matched, so that makers can place multiple orders based on the same *Source*.
When the balance of the owner account changes, SourceRemaining is adjusted accordingly and any untradable orders are canceled. 

*Takers always trade at the best price*. In case there is a better price in the market, price improvement is passed to the taker who pays less than the specified amount of *Source* tokens.

*Arbitrage-free*. Sophisticated order matching ensures that no arbitrage opportunities exist in the market. Orders always trade at the best price by considering synthetic instruments, e.g. a single eusd->eeur order matched against eeur->egbp and egbp->eusd simultaneously.

*Price/time priority matching*. Orders at the same price will be ordered by OrderId, with the lowest matched first.  

*Immediate settlement*. Matched orders are settled immediately with finality.

## Order Data

Internally, an order consists of the following data:

* Owner: a `AccAddress` which will be used for settlement and order modifications.
* OrderId: a `uint64` assigned by the market module, monotonically increasing.
* ClientOrderId: a `string` assigned by owner, which must not be a duplicate of an existing order.
* Source: a `Coin` representing the desired amount of tokens to sell.
* SourceFilled: `Int` that tracks the sold amount so far.
* SourceRemaining: a `Int` that is adjusted with *SourceFilled* and if the owner account balance change.
* Destination: a `Coin` representing the minimum amount of tokens to buy.
* DestinationFilled: `Int` that tracks the bought amount so far.
* Price: a `Dec` calculated as *Destination* / *Source*.

## Transaction Types

Similarly to the [FIX trading specification](https://www.fixtrading.org/online-specification/business-area-trade/) for single order handling, orders are uniquely identified by a client generated order ID. This client order ID can be used to subsequently cancel or replace an active order.

The market transactions have fixed gas prices:
| Message | Gas Price |
|------|------|
| MsgAddOrder | 25000 |
| MsgCancelOrder | 12500 |
| MsgCancelReplaceOrder | 25000 |

### MsgAddOrder

Adds a new order to the order book. The client order id is case sensitive with a 32 character maximum and must not collide with any active order for the same account.

In the below example, account emoney1uutrx7m0ap4ekt3d0vxlnyvnhsdv247sqrt045 wishes to purchase 7462230 edkk by selling (at most) 1000000 eeur tokens. The limit price is calculated as 7.46223 (7462230 / 1000000).

```json
{
  "type": "cosmos-sdk/StdTx",
  "value": {
    "msg": [
      {
        "type": "e-money/MsgAddOrder",
        "value": {
          "owner": "emoney1uutrx7m0ap4ekt3d0vxlnyvnhsdv247sqrt045",
          "source": {
            "denom": "eeur",
            "amount": "1000000"
          },
          "destination": {
            "denom": "edkk",
            "amount": "7462230"
          },
          "client_order_id": "order1"
        }
      }
    ],
    "fee": {
      "amount": [],
      "gas": "25000"
    },
    "signatures": null,
    "memo": ""
  }
}
```

### MsgCancelOrder

Cancels the remaining part of an existing order, referenced by it's client order ID. In case the order has already been fully filled, an error will be returned. 

```json
{
  "type": "cosmos-sdk/StdTx",
  "value": {
    "msg": [
      {
        "type": "e-money/MsgCancelOrder",
        "value": {
          "owner": "emoney1uutrx7m0ap4ekt3d0vxlnyvnhsdv247sqrt045",
          "client_order_id": "order1"
        }
      }
    ],
    "fee": {
      "amount": [],
      "gas": "12500"
    },
    "signatures": null,
    "memo": ""
  }
}
```

### MsgCancelReplaceOrder

Cancels the remaining part of an existing order, referenced by it's client order ID ("original_client_order_id"). The filled part of the cancelled order is then carried over into the new order (the replacement).

Cancel/replacing orders is ideal for liquidity providers to ensure that they do not miss trading opportunities and can provide constant liquidity.

In the below example, the initial order (see MsgAddOrder above) with a limit price of 7.46223 is adjusted to 7.46523.

```json
{
  "type": "cosmos-sdk/StdTx",
  "value": {
    "msg": [
      {
        "type": "e-money/MsgCancelReplaceOrder",
        "value": {
          "owner": "emoney1uutrx7m0ap4ekt3d0vxlnyvnhsdv247sqrt045",
          "source": {
            "denom": "eeur",
            "amount": "1000000"
          },
          "destination": {
            "denom": "edkk",
            "amount": "7465230"
          },
          "original_client_order_id": "order1",
          "client_order_id": "order2"
        }
      }
    ],
    "fee": {
      "amount": [],
      "gas": "25000"
    },
    "signatures": null,
    "memo": ""
  }
}
```

## Queries

The market module can be queried using the [REST interface](https://cosmos.network/rpc/) of any em-ledger node.
A public interface is exposed at https://emoney.validator.network/light/.

### Active account orders

Active orders for a given account can be queried using `https://emoney.validator.network/light/market/account/<owner>`.

Example output from `https://emoney.validator.network/light/market/account/emoney1n6kqrlmdhenstppxuaemczplnex0cyk9ddm3yd`:

```json
{
  "height": "121504",
  "result": {
    "orders": [
      {
        "id": 0,
        "created": "2020-05-07T13:17:52.180055879Z",
        "owner": "emoney1n6kqrlmdhenstppxuaemczplnex0cyk9ddm3yd",
        "client_order_id": "eurdkk1",
        "source": {
          "denom": "eeur",
          "amount": "1000000"
        },
        "source_remaining": "1000000",
        "source_filled": "0",
        "destination": {
          "denom": "edkk",
          "amount": "7462230"
        },
        "destination_filled": "0"
      },
      {
        "id": 1,
        "created": "2020-05-07T13:19:13.66329202Z",
        "owner": "emoney1n6kqrlmdhenstppxuaemczplnex0cyk9ddm3yd",
        "client_order_id": "chfeur1",
        "source": {
          "denom": "echf",
          "amount": "200000000"
        },
        "source_remaining": "200000000",
        "source_filled": "0",
        "destination": {
          "denom": "eeur",
          "amount": "190000000"
        },
        "destination_filled": "0"
      }
    ]
  }
}
```

### Active instruments

All instruments with active orders can be queried using https://emoney.validator.network/light/market/instruments.

Note that there is no listing requirement for new instruments, so these are created on-the-fly based on new orders. Instruments without orders are not returned.

Example output from `https://emoney.validator.network/light/market/instruments`:

```json
{
  "height": "121495",
  "result": {
    "instruments": [
      {
        "source": "eeur",
        "destination": "edkk",
        "order_count": 1
      },
      {
        "source": "echf",
        "destination": "eeur",
        "order_count": 1
      }
    ]
  }
}
```

### Instrument orders

All active orders for a given instrument are can be retrieved using `https://emoney.validator.network/light/market/instrument/<source>/<destination>`.

Example output from `https://emoney.validator.network/light/market/instrument/echf/eeur`:

```json
{
  "height": "121500",
  "result": {
    "source": "echf",
    "destination": "eeur",
    "orders": [
      {
        "id": 1,
        "created": "2020-05-07T13:19:13.66329202Z",
        "owner": "emoney1n6kqrlmdhenstppxuaemczplnex0cyk9ddm3yd",
        "source_remaining": "200000000",
        "price": "0.950000000000000000"
      }
    ]
  }
}
```
