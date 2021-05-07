# State

## Order State

Order state consists of the following data:

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
* Created: the Block 'Timestamp' at which the order is processed.
