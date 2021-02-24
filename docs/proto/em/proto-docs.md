<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [em/authority/v1beta1/authority.proto](#em/authority/v1beta1/authority.proto)
    - [RestrictedDenom](#em.authority.v1beta1.RestrictedDenom)
  
- [em/authority/v1beta1/genesis.proto](#em/authority/v1beta1/genesis.proto)
    - [GenesisState](#em.authority.v1beta1.GenesisState)
  
- [em/authority/v1beta1/query.proto](#em/authority/v1beta1/query.proto)
    - [QueryGasPricesRequest](#em.authority.v1beta1.QueryGasPricesRequest)
    - [QueryGasPricesResponse](#em.authority.v1beta1.QueryGasPricesResponse)
  
    - [Query](#em.authority.v1beta1.Query)
  
- [em/authority/v1beta1/tx.proto](#em/authority/v1beta1/tx.proto)
    - [MsgCreateIssuer](#em.authority.v1beta1.MsgCreateIssuer)
    - [MsgCreateIssuerResponse](#em.authority.v1beta1.MsgCreateIssuerResponse)
    - [MsgDestroyIssuer](#em.authority.v1beta1.MsgDestroyIssuer)
    - [MsgDestroyIssuerResponse](#em.authority.v1beta1.MsgDestroyIssuerResponse)
    - [MsgSetGasPrices](#em.authority.v1beta1.MsgSetGasPrices)
    - [MsgSetGasPricesResponse](#em.authority.v1beta1.MsgSetGasPricesResponse)
  
    - [Msg](#em.authority.v1beta1.Msg)
  
- [em/buyback/v1beta1/genesis.proto](#em/buyback/v1beta1/genesis.proto)
    - [GenesisState](#em.buyback.v1beta1.GenesisState)
  
- [em/inflation/v1beta1/inflation.proto](#em/inflation/v1beta1/inflation.proto)
    - [InflationAsset](#em.inflation.v1beta1.InflationAsset)
    - [InflationState](#em.inflation.v1beta1.InflationState)
  
- [em/inflation/v1beta1/genesis.proto](#em/inflation/v1beta1/genesis.proto)
    - [GenesisState](#em.inflation.v1beta1.GenesisState)
  
- [em/issuer/v1beta1/issuer.proto](#em/issuer/v1beta1/issuer.proto)
    - [Issuer](#em.issuer.v1beta1.Issuer)
  
- [em/issuer/v1beta1/genesis.proto](#em/issuer/v1beta1/genesis.proto)
    - [GenesisState](#em.issuer.v1beta1.GenesisState)
  
- [em/issuer/v1beta1/tx.proto](#em/issuer/v1beta1/tx.proto)
    - [MsgDecreaseMintable](#em.issuer.v1beta1.MsgDecreaseMintable)
    - [MsgIncreaseMintable](#em.issuer.v1beta1.MsgIncreaseMintable)
    - [MsgRevokeLiquidityProvider](#em.issuer.v1beta1.MsgRevokeLiquidityProvider)
    - [MsgSetInflation](#em.issuer.v1beta1.MsgSetInflation)
  
- [em/liquidityprovider/v1beta1/genesis.proto](#em/liquidityprovider/v1beta1/genesis.proto)
    - [GenesisAcc](#em.liquidityprovider.v1beta1.GenesisAcc)
    - [GenesisState](#em.liquidityprovider.v1beta1.GenesisState)
  
- [em/liquidityprovider/v1beta1/liquidityprovider.proto](#em/liquidityprovider/v1beta1/liquidityprovider.proto)
    - [LiquidityProviderAccount](#em.liquidityprovider.v1beta1.LiquidityProviderAccount)
  
- [em/liquidityprovider/v1beta1/tx.proto](#em/liquidityprovider/v1beta1/tx.proto)
    - [MsgBurnTokens](#em.liquidityprovider.v1beta1.MsgBurnTokens)
    - [MsgMintTokens](#em.liquidityprovider.v1beta1.MsgMintTokens)
  
- [em/market/v1beta1/market.proto](#em/market/v1beta1/market.proto)
    - [ExecutionPlan](#em.market.v1beta1.ExecutionPlan)
    - [Instrument](#em.market.v1beta1.Instrument)
    - [MarketData](#em.market.v1beta1.MarketData)
    - [Order](#em.market.v1beta1.Order)
  
    - [TimeInForce](#em.market.v1beta1.TimeInForce)
  
- [em/market/v1beta1/tx.proto](#em/market/v1beta1/tx.proto)
    - [MsgAddLimitOrder](#em.market.v1beta1.MsgAddLimitOrder)
    - [MsgAddMarketOrder](#em.market.v1beta1.MsgAddMarketOrder)
    - [MsgCancelOrder](#em.market.v1beta1.MsgCancelOrder)
    - [MsgCancelReplaceLimitOrder](#em.market.v1beta1.MsgCancelReplaceLimitOrder)
  
- [em/slashing/v1beta1/slashing.proto](#em/slashing/v1beta1/slashing.proto)
    - [Penalties](#em.slashing.v1beta1.Penalties)
    - [Penalty](#em.slashing.v1beta1.Penalty)
  
- [Scalar Value Types](#scalar-value-types)



<a name="em/authority/v1beta1/authority.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/authority/v1beta1/authority.proto



<a name="em.authority.v1beta1.RestrictedDenom"></a>

### RestrictedDenom



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | todo (reviewer) : moved from /types/ todo (reviewer) : please note the lower case json/yaml attribute names now (convention) |
| `allowed` | [string](#string) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/authority/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/authority/v1beta1/genesis.proto



<a name="em.authority.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `restricted_denoms` | [RestrictedDenom](#em.authority.v1beta1.RestrictedDenom) | repeated |  |
| `min_gas_prices` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/authority/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/authority/v1beta1/query.proto



<a name="em.authority.v1beta1.QueryGasPricesRequest"></a>

### QueryGasPricesRequest







<a name="em.authority.v1beta1.QueryGasPricesResponse"></a>

### QueryGasPricesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_gas_prices` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.authority.v1beta1.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `GasPrices` | [QueryGasPricesRequest](#em.authority.v1beta1.QueryGasPricesRequest) | [QueryGasPricesResponse](#em.authority.v1beta1.QueryGasPricesResponse) |  | GET|/e-money/authority/v1beta1/gasprices|

 <!-- end services -->



<a name="em/authority/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/authority/v1beta1/tx.proto



<a name="em.authority.v1beta1.MsgCreateIssuer"></a>

### MsgCreateIssuer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `issuer` | [string](#string) |  |  |
| `denominations` | [string](#string) | repeated |  |






<a name="em.authority.v1beta1.MsgCreateIssuerResponse"></a>

### MsgCreateIssuerResponse







<a name="em.authority.v1beta1.MsgDestroyIssuer"></a>

### MsgDestroyIssuer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `issuer` | [string](#string) |  |  |






<a name="em.authority.v1beta1.MsgDestroyIssuerResponse"></a>

### MsgDestroyIssuerResponse







<a name="em.authority.v1beta1.MsgSetGasPrices"></a>

### MsgSetGasPrices



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `gas_prices` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |






<a name="em.authority.v1beta1.MsgSetGasPricesResponse"></a>

### MsgSetGasPricesResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.authority.v1beta1.Msg"></a>

### Msg


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateIssuer` | [MsgCreateIssuer](#em.authority.v1beta1.MsgCreateIssuer) | [MsgCreateIssuerResponse](#em.authority.v1beta1.MsgCreateIssuerResponse) |  | |
| `DestroyIssuer` | [MsgDestroyIssuer](#em.authority.v1beta1.MsgDestroyIssuer) | [MsgDestroyIssuerResponse](#em.authority.v1beta1.MsgDestroyIssuerResponse) |  | |
| `SetGasPrices` | [MsgSetGasPrices](#em.authority.v1beta1.MsgSetGasPrices) | [MsgSetGasPricesResponse](#em.authority.v1beta1.MsgSetGasPricesResponse) |  | |

 <!-- end services -->



<a name="em/buyback/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/buyback/v1beta1/genesis.proto



<a name="em.buyback.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `interval` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/inflation/v1beta1/inflation.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/inflation/v1beta1/inflation.proto



<a name="em.inflation.v1beta1.InflationAsset"></a>

### InflationAsset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `inflation` | [string](#string) |  |  |
| `accum` | [string](#string) |  |  |






<a name="em.inflation.v1beta1.InflationState"></a>

### InflationState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `last_applied` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `last_applied_height` | [string](#string) |  |  |
| `assets` | [InflationAsset](#em.inflation.v1beta1.InflationAsset) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/inflation/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/inflation/v1beta1/genesis.proto



<a name="em.inflation.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assets` | [InflationState](#em.inflation.v1beta1.InflationState) |  | todo (reviewer): yaml naming is a bit inconsistent. state contains assets |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/issuer/v1beta1/issuer.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/issuer/v1beta1/issuer.proto



<a name="em.issuer.v1beta1.Issuer"></a>

### Issuer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `denoms` | [string](#string) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/issuer/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/issuer/v1beta1/genesis.proto



<a name="em.issuer.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuers` | [Issuer](#em.issuer.v1beta1.Issuer) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/issuer/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/issuer/v1beta1/tx.proto



<a name="em.issuer.v1beta1.MsgDecreaseMintable"></a>

### MsgDecreaseMintable



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuer` | [string](#string) |  |  |
| `liquidity_provider` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="em.issuer.v1beta1.MsgIncreaseMintable"></a>

### MsgIncreaseMintable



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuer` | [string](#string) |  |  |
| `liquidity_provider` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="em.issuer.v1beta1.MsgRevokeLiquidityProvider"></a>

### MsgRevokeLiquidityProvider



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuer` | [string](#string) |  |  |
| `liquidity_provider` | [string](#string) |  |  |






<a name="em.issuer.v1beta1.MsgSetInflation"></a>

### MsgSetInflation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuer` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `inflation_rate` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/liquidityprovider/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/liquidityprovider/v1beta1/genesis.proto



<a name="em.liquidityprovider.v1beta1.GenesisAcc"></a>

### GenesisAcc



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `mintable` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="em.liquidityprovider.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `accounts` | [GenesisAcc](#em.liquidityprovider.v1beta1.GenesisAcc) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/liquidityprovider/v1beta1/liquidityprovider.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/liquidityprovider/v1beta1/liquidityprovider.proto



<a name="em.liquidityprovider.v1beta1.LiquidityProviderAccount"></a>

### LiquidityProviderAccount



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `mintable` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/liquidityprovider/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/liquidityprovider/v1beta1/tx.proto



<a name="em.liquidityprovider.v1beta1.MsgBurnTokens"></a>

### MsgBurnTokens



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `liquidity_provider` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="em.liquidityprovider.v1beta1.MsgMintTokens"></a>

### MsgMintTokens



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `liquidity_provider` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/market/v1beta1/market.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/market/v1beta1/market.proto



<a name="em.market.v1beta1.ExecutionPlan"></a>

### ExecutionPlan



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `price` | [string](#string) |  |  |
| `first_order` | [Order](#em.market.v1beta1.Order) |  |  |
| `second_order` | [Order](#em.market.v1beta1.Order) |  |  |






<a name="em.market.v1beta1.Instrument"></a>

### Instrument



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  |  |
| `destination` | [string](#string) |  |  |






<a name="em.market.v1beta1.MarketData"></a>

### MarketData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  |  |
| `destination` | [string](#string) |  |  |
| `last_price` | [string](#string) |  |  |
| `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="em.market.v1beta1.Order"></a>

### Order



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  |  |
| `time_in_force` | [TimeInForce](#em.market.v1beta1.TimeInForce) |  |  |
| `owner` | [string](#string) |  |  |
| `client_order_id` | [string](#string) |  |  |
| `source` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `source_remaining` | [string](#string) |  |  |
| `source_filled` | [string](#string) |  |  |
| `destination` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `destination_filled` | [string](#string) |  |  |





 <!-- end messages -->


<a name="em.market.v1beta1.TimeInForce"></a>

### TimeInForce


| Name | Number | Description |
| ---- | ------ | ----------- |
| TIME_IN_FORCE_UNSPECIFIED | 0 |  |
| TIME_IN_FORCE_GOOD_TIL_CANCEL | 1 |  |
| TIME_IN_FORCE_IMMEDIATE_OR_CANCEL | 2 |  |
| TIME_IN_FORCE_FILL_OR_KILL | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/market/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/market/v1beta1/tx.proto



<a name="em.market.v1beta1.MsgAddLimitOrder"></a>

### MsgAddLimitOrder



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `client_order_id` | [string](#string) |  |  |
| `time_in_force` | [TimeInForce](#em.market.v1beta1.TimeInForce) |  |  |
| `source` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `destination` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="em.market.v1beta1.MsgAddMarketOrder"></a>

### MsgAddMarketOrder



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `client_order_id` | [string](#string) |  |  |
| `time_in_force` | [TimeInForce](#em.market.v1beta1.TimeInForce) |  |  |
| `source` | [string](#string) |  |  |
| `destination` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `maximum_slippage` | [string](#string) |  |  |






<a name="em.market.v1beta1.MsgCancelOrder"></a>

### MsgCancelOrder



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `client_order_id` | [string](#string) |  |  |






<a name="em.market.v1beta1.MsgCancelReplaceLimitOrder"></a>

### MsgCancelReplaceLimitOrder



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `original_client_order_id` | [string](#string) |  |  |
| `new_client_order_id` | [string](#string) |  |  |
| `source` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `destination` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/slashing/v1beta1/slashing.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/slashing/v1beta1/slashing.proto



<a name="em.slashing.v1beta1.Penalties"></a>

### Penalties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `elements` | [Penalty](#em.slashing.v1beta1.Penalty) | repeated |  |






<a name="em.slashing.v1beta1.Penalty"></a>

### Penalty



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator` | [string](#string) |  |  |
| `amounts` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

