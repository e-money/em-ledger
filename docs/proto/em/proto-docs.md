<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [em/authority/v1/authority.proto](#em/authority/v1/authority.proto)
    - [Authority](#em.authority.v1.Authority)
    - [GasPrices](#em.authority.v1.GasPrices)
  
- [em/authority/v1/genesis.proto](#em/authority/v1/genesis.proto)
    - [GenesisState](#em.authority.v1.GenesisState)
  
- [em/authority/v1/query.proto](#em/authority/v1/query.proto)
    - [QueryGasPricesRequest](#em.authority.v1.QueryGasPricesRequest)
    - [QueryGasPricesResponse](#em.authority.v1.QueryGasPricesResponse)
    - [QueryUpgradePlanRequest](#em.authority.v1.QueryUpgradePlanRequest)
    - [QueryUpgradePlanResponse](#em.authority.v1.QueryUpgradePlanResponse)
  
    - [Query](#em.authority.v1.Query)
  
- [em/authority/v1/tx.proto](#em/authority/v1/tx.proto)
    - [Denomination](#em.authority.v1.Denomination)
    - [MsgCreateIssuer](#em.authority.v1.MsgCreateIssuer)
    - [MsgCreateIssuerResponse](#em.authority.v1.MsgCreateIssuerResponse)
    - [MsgDestroyIssuer](#em.authority.v1.MsgDestroyIssuer)
    - [MsgDestroyIssuerResponse](#em.authority.v1.MsgDestroyIssuerResponse)
    - [MsgReplaceAuthority](#em.authority.v1.MsgReplaceAuthority)
    - [MsgReplaceAuthorityResponse](#em.authority.v1.MsgReplaceAuthorityResponse)
    - [MsgScheduleUpgrade](#em.authority.v1.MsgScheduleUpgrade)
    - [MsgScheduleUpgradeResponse](#em.authority.v1.MsgScheduleUpgradeResponse)
    - [MsgSetGasPrices](#em.authority.v1.MsgSetGasPrices)
    - [MsgSetGasPricesResponse](#em.authority.v1.MsgSetGasPricesResponse)
    - [MsgSetParameters](#em.authority.v1.MsgSetParameters)
    - [MsgSetParametersResponse](#em.authority.v1.MsgSetParametersResponse)
  
    - [Msg](#em.authority.v1.Msg)
  
- [em/buyback/v1/genesis.proto](#em/buyback/v1/genesis.proto)
    - [GenesisState](#em.buyback.v1.GenesisState)
  
- [em/buyback/v1/query.proto](#em/buyback/v1/query.proto)
    - [QueryBalanceRequest](#em.buyback.v1.QueryBalanceRequest)
    - [QueryBalanceResponse](#em.buyback.v1.QueryBalanceResponse)
    - [QueryBuybackTimeRequest](#em.buyback.v1.QueryBuybackTimeRequest)
    - [QueryBuybackTimeResponse](#em.buyback.v1.QueryBuybackTimeResponse)
  
    - [Query](#em.buyback.v1.Query)
  
- [em/inflation/v1/inflation.proto](#em/inflation/v1/inflation.proto)
    - [InflationAsset](#em.inflation.v1.InflationAsset)
    - [InflationState](#em.inflation.v1.InflationState)
  
- [em/inflation/v1/genesis.proto](#em/inflation/v1/genesis.proto)
    - [GenesisState](#em.inflation.v1.GenesisState)
  
- [em/inflation/v1/query.proto](#em/inflation/v1/query.proto)
    - [QueryInflationRequest](#em.inflation.v1.QueryInflationRequest)
    - [QueryInflationResponse](#em.inflation.v1.QueryInflationResponse)
  
    - [Query](#em.inflation.v1.Query)
  
- [em/issuer/v1/issuer.proto](#em/issuer/v1/issuer.proto)
    - [Issuer](#em.issuer.v1.Issuer)
    - [Issuers](#em.issuer.v1.Issuers)
  
- [em/issuer/v1/genesis.proto](#em/issuer/v1/genesis.proto)
    - [GenesisState](#em.issuer.v1.GenesisState)
  
- [em/issuer/v1/query.proto](#em/issuer/v1/query.proto)
    - [QueryIssuersRequest](#em.issuer.v1.QueryIssuersRequest)
    - [QueryIssuersResponse](#em.issuer.v1.QueryIssuersResponse)
  
    - [Query](#em.issuer.v1.Query)
  
- [em/issuer/v1/tx.proto](#em/issuer/v1/tx.proto)
    - [MsgDecreaseMintable](#em.issuer.v1.MsgDecreaseMintable)
    - [MsgDecreaseMintableResponse](#em.issuer.v1.MsgDecreaseMintableResponse)
    - [MsgIncreaseMintable](#em.issuer.v1.MsgIncreaseMintable)
    - [MsgIncreaseMintableResponse](#em.issuer.v1.MsgIncreaseMintableResponse)
    - [MsgRevokeLiquidityProvider](#em.issuer.v1.MsgRevokeLiquidityProvider)
    - [MsgRevokeLiquidityProviderResponse](#em.issuer.v1.MsgRevokeLiquidityProviderResponse)
    - [MsgSetInflation](#em.issuer.v1.MsgSetInflation)
    - [MsgSetInflationResponse](#em.issuer.v1.MsgSetInflationResponse)
  
    - [Msg](#em.issuer.v1.Msg)
  
- [em/liquidityprovider/v1/genesis.proto](#em/liquidityprovider/v1/genesis.proto)
    - [GenesisAcc](#em.liquidityprovider.v1.GenesisAcc)
    - [GenesisState](#em.liquidityprovider.v1.GenesisState)
  
- [em/liquidityprovider/v1/liquidityprovider.proto](#em/liquidityprovider/v1/liquidityprovider.proto)
    - [LiquidityProviderAccount](#em.liquidityprovider.v1.LiquidityProviderAccount)
  
- [em/liquidityprovider/v1/query.proto](#em/liquidityprovider/v1/query.proto)
    - [QueryListRequest](#em.liquidityprovider.v1.QueryListRequest)
    - [QueryListResponse](#em.liquidityprovider.v1.QueryListResponse)
    - [QueryMintableRequest](#em.liquidityprovider.v1.QueryMintableRequest)
    - [QueryMintableResponse](#em.liquidityprovider.v1.QueryMintableResponse)
  
    - [Query](#em.liquidityprovider.v1.Query)
  
- [em/liquidityprovider/v1/tx.proto](#em/liquidityprovider/v1/tx.proto)
    - [MsgBurnTokens](#em.liquidityprovider.v1.MsgBurnTokens)
    - [MsgBurnTokensResponse](#em.liquidityprovider.v1.MsgBurnTokensResponse)
    - [MsgMintTokens](#em.liquidityprovider.v1.MsgMintTokens)
    - [MsgMintTokensResponse](#em.liquidityprovider.v1.MsgMintTokensResponse)
  
    - [Msg](#em.liquidityprovider.v1.Msg)
  
- [em/market/v1/market.proto](#em/market/v1/market.proto)
    - [ExecutionPlan](#em.market.v1.ExecutionPlan)
    - [Instrument](#em.market.v1.Instrument)
    - [MarketData](#em.market.v1.MarketData)
    - [Order](#em.market.v1.Order)
  
    - [TimeInForce](#em.market.v1.TimeInForce)
  
- [em/market/v1/query.proto](#em/market/v1/query.proto)
    - [QueryByAccountRequest](#em.market.v1.QueryByAccountRequest)
    - [QueryByAccountResponse](#em.market.v1.QueryByAccountResponse)
    - [QueryInstrumentRequest](#em.market.v1.QueryInstrumentRequest)
    - [QueryInstrumentResponse](#em.market.v1.QueryInstrumentResponse)
    - [QueryInstrumentsRequest](#em.market.v1.QueryInstrumentsRequest)
    - [QueryInstrumentsResponse](#em.market.v1.QueryInstrumentsResponse)
    - [QueryInstrumentsResponse.Element](#em.market.v1.QueryInstrumentsResponse.Element)
    - [QueryOrderResponse](#em.market.v1.QueryOrderResponse)
  
    - [Query](#em.market.v1.Query)
  
- [em/market/v1/tx.proto](#em/market/v1/tx.proto)
    - [MsgAddLimitOrder](#em.market.v1.MsgAddLimitOrder)
    - [MsgAddLimitOrderResponse](#em.market.v1.MsgAddLimitOrderResponse)
    - [MsgAddMarketOrder](#em.market.v1.MsgAddMarketOrder)
    - [MsgAddMarketOrderResponse](#em.market.v1.MsgAddMarketOrderResponse)
    - [MsgCancelOrder](#em.market.v1.MsgCancelOrder)
    - [MsgCancelOrderResponse](#em.market.v1.MsgCancelOrderResponse)
    - [MsgCancelReplaceLimitOrder](#em.market.v1.MsgCancelReplaceLimitOrder)
    - [MsgCancelReplaceLimitOrderResponse](#em.market.v1.MsgCancelReplaceLimitOrderResponse)
    - [MsgCancelReplaceMarketOrder](#em.market.v1.MsgCancelReplaceMarketOrder)
    - [MsgCancelReplaceMarketOrderResponse](#em.market.v1.MsgCancelReplaceMarketOrderResponse)
  
    - [Msg](#em.market.v1.Msg)
  
- [em/queries/v1/query.proto](#em/queries/v1/query.proto)
    - [MissedBlocksInfo](#em.queries.v1.MissedBlocksInfo)
    - [QueryCirculatingRequest](#em.queries.v1.QueryCirculatingRequest)
    - [QueryCirculatingResponse](#em.queries.v1.QueryCirculatingResponse)
    - [QueryMissedBlocksRequest](#em.queries.v1.QueryMissedBlocksRequest)
    - [QueryMissedBlocksResponse](#em.queries.v1.QueryMissedBlocksResponse)
    - [QuerySpendableRequest](#em.queries.v1.QuerySpendableRequest)
    - [QuerySpendableResponse](#em.queries.v1.QuerySpendableResponse)
  
    - [Query](#em.queries.v1.Query)
  
- [Scalar Value Types](#scalar-value-types)



<a name="em/authority/v1/authority.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/authority/v1/authority.proto



<a name="em.authority.v1.Authority"></a>

### Authority



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `former_address` | [string](#string) |  |  |
| `last_modified` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="em.authority.v1.GasPrices"></a>

### GasPrices



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minimum` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/authority/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/authority/v1/genesis.proto



<a name="em.authority.v1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `min_gas_prices` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/authority/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/authority/v1/query.proto



<a name="em.authority.v1.QueryGasPricesRequest"></a>

### QueryGasPricesRequest







<a name="em.authority.v1.QueryGasPricesResponse"></a>

### QueryGasPricesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_gas_prices` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |






<a name="em.authority.v1.QueryUpgradePlanRequest"></a>

### QueryUpgradePlanRequest







<a name="em.authority.v1.QueryUpgradePlanResponse"></a>

### QueryUpgradePlanResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `plan` | [cosmos.upgrade.v1beta1.Plan](#cosmos.upgrade.v1beta1.Plan) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.authority.v1.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `GasPrices` | [QueryGasPricesRequest](#em.authority.v1.QueryGasPricesRequest) | [QueryGasPricesResponse](#em.authority.v1.QueryGasPricesResponse) |  | GET|/e-money/authority/v1/gasprices|
| `UpgradePlan` | [QueryUpgradePlanRequest](#em.authority.v1.QueryUpgradePlanRequest) | [QueryUpgradePlanResponse](#em.authority.v1.QueryUpgradePlanResponse) |  | GET|/e-money/authority/v1/upgrade_plan|

 <!-- end services -->



<a name="em/authority/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/authority/v1/tx.proto



<a name="em.authority.v1.Denomination"></a>

### Denomination



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base` | [string](#string) |  | base represents the base denom (should be the DenomUnit with exponent = 0). |
| `display` | [string](#string) |  | display indicates the suggested denom that should be displayed in clients. |
| `description` | [string](#string) |  |  |






<a name="em.authority.v1.MsgCreateIssuer"></a>

### MsgCreateIssuer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `issuer` | [string](#string) |  |  |
| `denominations` | [Denomination](#em.authority.v1.Denomination) | repeated |  |






<a name="em.authority.v1.MsgCreateIssuerResponse"></a>

### MsgCreateIssuerResponse







<a name="em.authority.v1.MsgDestroyIssuer"></a>

### MsgDestroyIssuer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `issuer` | [string](#string) |  |  |






<a name="em.authority.v1.MsgDestroyIssuerResponse"></a>

### MsgDestroyIssuerResponse







<a name="em.authority.v1.MsgReplaceAuthority"></a>

### MsgReplaceAuthority



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `new_authority` | [string](#string) |  |  |






<a name="em.authority.v1.MsgReplaceAuthorityResponse"></a>

### MsgReplaceAuthorityResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `new_authority_address` | [string](#string) |  |  |






<a name="em.authority.v1.MsgScheduleUpgrade"></a>

### MsgScheduleUpgrade



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `plan` | [cosmos.upgrade.v1beta1.Plan](#cosmos.upgrade.v1beta1.Plan) |  |  |






<a name="em.authority.v1.MsgScheduleUpgradeResponse"></a>

### MsgScheduleUpgradeResponse







<a name="em.authority.v1.MsgSetGasPrices"></a>

### MsgSetGasPrices



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `gas_prices` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |






<a name="em.authority.v1.MsgSetGasPricesResponse"></a>

### MsgSetGasPricesResponse







<a name="em.authority.v1.MsgSetParameters"></a>

### MsgSetParameters



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `changes` | [cosmos.params.v1beta1.ParamChange](#cosmos.params.v1beta1.ParamChange) | repeated |  |






<a name="em.authority.v1.MsgSetParametersResponse"></a>

### MsgSetParametersResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.authority.v1.Msg"></a>

### Msg


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateIssuer` | [MsgCreateIssuer](#em.authority.v1.MsgCreateIssuer) | [MsgCreateIssuerResponse](#em.authority.v1.MsgCreateIssuerResponse) |  | |
| `DestroyIssuer` | [MsgDestroyIssuer](#em.authority.v1.MsgDestroyIssuer) | [MsgDestroyIssuerResponse](#em.authority.v1.MsgDestroyIssuerResponse) |  | |
| `SetGasPrices` | [MsgSetGasPrices](#em.authority.v1.MsgSetGasPrices) | [MsgSetGasPricesResponse](#em.authority.v1.MsgSetGasPricesResponse) |  | |
| `ReplaceAuthority` | [MsgReplaceAuthority](#em.authority.v1.MsgReplaceAuthority) | [MsgReplaceAuthorityResponse](#em.authority.v1.MsgReplaceAuthorityResponse) |  | |
| `ScheduleUpgrade` | [MsgScheduleUpgrade](#em.authority.v1.MsgScheduleUpgrade) | [MsgScheduleUpgradeResponse](#em.authority.v1.MsgScheduleUpgradeResponse) |  | |
| `SetParameters` | [MsgSetParameters](#em.authority.v1.MsgSetParameters) | [MsgSetParametersResponse](#em.authority.v1.MsgSetParametersResponse) |  | |

 <!-- end services -->



<a name="em/buyback/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/buyback/v1/genesis.proto



<a name="em.buyback.v1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `interval` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/buyback/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/buyback/v1/query.proto



<a name="em.buyback.v1.QueryBalanceRequest"></a>

### QueryBalanceRequest







<a name="em.buyback.v1.QueryBalanceResponse"></a>

### QueryBalanceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `balance` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="em.buyback.v1.QueryBuybackTimeRequest"></a>

### QueryBuybackTimeRequest







<a name="em.buyback.v1.QueryBuybackTimeResponse"></a>

### QueryBuybackTimeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `last_run` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `next_run` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.buyback.v1.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Balance` | [QueryBalanceRequest](#em.buyback.v1.QueryBalanceRequest) | [QueryBalanceResponse](#em.buyback.v1.QueryBalanceResponse) | Query for the current buyback balance | GET|/e-money/buyback/v1/balance|
| `BuybackTime` | [QueryBuybackTimeRequest](#em.buyback.v1.QueryBuybackTimeRequest) | [QueryBuybackTimeResponse](#em.buyback.v1.QueryBuybackTimeResponse) | Query for buyback time periods | GET|/e-money/buyback/v1/time|

 <!-- end services -->



<a name="em/inflation/v1/inflation.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/inflation/v1/inflation.proto



<a name="em.inflation.v1.InflationAsset"></a>

### InflationAsset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `inflation` | [string](#string) |  |  |
| `accum` | [string](#string) |  |  |






<a name="em.inflation.v1.InflationState"></a>

### InflationState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `last_applied` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `last_applied_height` | [string](#string) |  |  |
| `assets` | [InflationAsset](#em.inflation.v1.InflationAsset) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/inflation/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/inflation/v1/genesis.proto



<a name="em.inflation.v1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assets` | [InflationState](#em.inflation.v1.InflationState) |  | todo (reviewer): yaml naming is a bit inconsistent. state contains assets |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/inflation/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/inflation/v1/query.proto



<a name="em.inflation.v1.QueryInflationRequest"></a>

### QueryInflationRequest







<a name="em.inflation.v1.QueryInflationResponse"></a>

### QueryInflationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `state` | [InflationState](#em.inflation.v1.InflationState) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.inflation.v1.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Inflation` | [QueryInflationRequest](#em.inflation.v1.QueryInflationRequest) | [QueryInflationResponse](#em.inflation.v1.QueryInflationResponse) |  | GET|/e-money/inflation/v1/state|

 <!-- end services -->



<a name="em/issuer/v1/issuer.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/issuer/v1/issuer.proto



<a name="em.issuer.v1.Issuer"></a>

### Issuer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `denoms` | [string](#string) | repeated |  |






<a name="em.issuer.v1.Issuers"></a>

### Issuers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuers` | [Issuer](#em.issuer.v1.Issuer) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/issuer/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/issuer/v1/genesis.proto



<a name="em.issuer.v1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuers` | [Issuer](#em.issuer.v1.Issuer) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/issuer/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/issuer/v1/query.proto



<a name="em.issuer.v1.QueryIssuersRequest"></a>

### QueryIssuersRequest







<a name="em.issuer.v1.QueryIssuersResponse"></a>

### QueryIssuersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuers` | [Issuer](#em.issuer.v1.Issuer) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.issuer.v1.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Issuers` | [QueryIssuersRequest](#em.issuer.v1.QueryIssuersRequest) | [QueryIssuersResponse](#em.issuer.v1.QueryIssuersResponse) |  | GET|/e-money/issuer/v1/issuers|

 <!-- end services -->



<a name="em/issuer/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/issuer/v1/tx.proto



<a name="em.issuer.v1.MsgDecreaseMintable"></a>

### MsgDecreaseMintable



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuer` | [string](#string) |  |  |
| `liquidity_provider` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="em.issuer.v1.MsgDecreaseMintableResponse"></a>

### MsgDecreaseMintableResponse







<a name="em.issuer.v1.MsgIncreaseMintable"></a>

### MsgIncreaseMintable



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuer` | [string](#string) |  |  |
| `liquidity_provider` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="em.issuer.v1.MsgIncreaseMintableResponse"></a>

### MsgIncreaseMintableResponse







<a name="em.issuer.v1.MsgRevokeLiquidityProvider"></a>

### MsgRevokeLiquidityProvider



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuer` | [string](#string) |  |  |
| `liquidity_provider` | [string](#string) |  |  |






<a name="em.issuer.v1.MsgRevokeLiquidityProviderResponse"></a>

### MsgRevokeLiquidityProviderResponse







<a name="em.issuer.v1.MsgSetInflation"></a>

### MsgSetInflation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuer` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `inflation_rate` | [string](#string) |  |  |






<a name="em.issuer.v1.MsgSetInflationResponse"></a>

### MsgSetInflationResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.issuer.v1.Msg"></a>

### Msg


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `IncreaseMintable` | [MsgIncreaseMintable](#em.issuer.v1.MsgIncreaseMintable) | [MsgIncreaseMintableResponse](#em.issuer.v1.MsgIncreaseMintableResponse) |  | |
| `DecreaseMintable` | [MsgDecreaseMintable](#em.issuer.v1.MsgDecreaseMintable) | [MsgDecreaseMintableResponse](#em.issuer.v1.MsgDecreaseMintableResponse) |  | |
| `RevokeLiquidityProvider` | [MsgRevokeLiquidityProvider](#em.issuer.v1.MsgRevokeLiquidityProvider) | [MsgRevokeLiquidityProviderResponse](#em.issuer.v1.MsgRevokeLiquidityProviderResponse) |  | |
| `SetInflation` | [MsgSetInflation](#em.issuer.v1.MsgSetInflation) | [MsgSetInflationResponse](#em.issuer.v1.MsgSetInflationResponse) |  | |

 <!-- end services -->



<a name="em/liquidityprovider/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/liquidityprovider/v1/genesis.proto



<a name="em.liquidityprovider.v1.GenesisAcc"></a>

### GenesisAcc



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `mintable` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="em.liquidityprovider.v1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `accounts` | [GenesisAcc](#em.liquidityprovider.v1.GenesisAcc) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/liquidityprovider/v1/liquidityprovider.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/liquidityprovider/v1/liquidityprovider.proto



<a name="em.liquidityprovider.v1.LiquidityProviderAccount"></a>

### LiquidityProviderAccount



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | Any string address representation with the accompanying supporting encoding and validation functions starting with bech32. However, in the interest of cultivating wider acceptance for this module other arbitrary address encodings outside the supported cosmos sdk formats perhaps would fit nicely with this loosely defined provider identity specifier. |
| `mintable` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/liquidityprovider/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/liquidityprovider/v1/query.proto



<a name="em.liquidityprovider.v1.QueryListRequest"></a>

### QueryListRequest







<a name="em.liquidityprovider.v1.QueryListResponse"></a>

### QueryListResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `liquidity_providers` | [LiquidityProviderAccount](#em.liquidityprovider.v1.LiquidityProviderAccount) | repeated |  |






<a name="em.liquidityprovider.v1.QueryMintableRequest"></a>

### QueryMintableRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address defines the liquidity provider address to query mintable. |






<a name="em.liquidityprovider.v1.QueryMintableResponse"></a>

### QueryMintableResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `mintable` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.liquidityprovider.v1.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `List` | [QueryListRequest](#em.liquidityprovider.v1.QueryListRequest) | [QueryListResponse](#em.liquidityprovider.v1.QueryListResponse) |  | GET|/e-money/liquidityprovider/v1/list|
| `Mintable` | [QueryMintableRequest](#em.liquidityprovider.v1.QueryMintableRequest) | [QueryMintableResponse](#em.liquidityprovider.v1.QueryMintableResponse) |  | GET|/e-money/liquidityprovider/v1/mintable/{address}|

 <!-- end services -->



<a name="em/liquidityprovider/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/liquidityprovider/v1/tx.proto



<a name="em.liquidityprovider.v1.MsgBurnTokens"></a>

### MsgBurnTokens



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `liquidity_provider` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="em.liquidityprovider.v1.MsgBurnTokensResponse"></a>

### MsgBurnTokensResponse







<a name="em.liquidityprovider.v1.MsgMintTokens"></a>

### MsgMintTokens



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `liquidity_provider` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="em.liquidityprovider.v1.MsgMintTokensResponse"></a>

### MsgMintTokensResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.liquidityprovider.v1.Msg"></a>

### Msg


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `MintTokens` | [MsgMintTokens](#em.liquidityprovider.v1.MsgMintTokens) | [MsgMintTokensResponse](#em.liquidityprovider.v1.MsgMintTokensResponse) |  | |
| `BurnTokens` | [MsgBurnTokens](#em.liquidityprovider.v1.MsgBurnTokens) | [MsgBurnTokensResponse](#em.liquidityprovider.v1.MsgBurnTokensResponse) |  | |

 <!-- end services -->



<a name="em/market/v1/market.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/market/v1/market.proto



<a name="em.market.v1.ExecutionPlan"></a>

### ExecutionPlan



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `price` | [string](#string) |  |  |
| `first_order` | [Order](#em.market.v1.Order) |  |  |
| `second_order` | [Order](#em.market.v1.Order) |  |  |






<a name="em.market.v1.Instrument"></a>

### Instrument



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  |  |
| `destination` | [string](#string) |  |  |






<a name="em.market.v1.MarketData"></a>

### MarketData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  |  |
| `destination` | [string](#string) |  |  |
| `last_price` | [string](#string) |  |  |
| `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="em.market.v1.Order"></a>

### Order



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  |  |
| `time_in_force` | [TimeInForce](#em.market.v1.TimeInForce) |  |  |
| `owner` | [string](#string) |  |  |
| `client_order_id` | [string](#string) |  |  |
| `source` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `source_remaining` | [string](#string) |  |  |
| `source_filled` | [string](#string) |  |  |
| `destination` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `destination_filled` | [string](#string) |  |  |
| `created` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |





 <!-- end messages -->


<a name="em.market.v1.TimeInForce"></a>

### TimeInForce


| Name | Number | Description |
| ---- | ------ | ----------- |
| TIME_IN_FORCE_UNSPECIFIED | 0 |  |
| TIME_IN_FORCE_GOOD_TILL_CANCEL | 1 |  |
| TIME_IN_FORCE_IMMEDIATE_OR_CANCEL | 2 |  |
| TIME_IN_FORCE_FILL_OR_KILL | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="em/market/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/market/v1/query.proto



<a name="em.market.v1.QueryByAccountRequest"></a>

### QueryByAccountRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="em.market.v1.QueryByAccountResponse"></a>

### QueryByAccountResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `orders` | [Order](#em.market.v1.Order) | repeated |  |






<a name="em.market.v1.QueryInstrumentRequest"></a>

### QueryInstrumentRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  |  |
| `destination` | [string](#string) |  |  |






<a name="em.market.v1.QueryInstrumentResponse"></a>

### QueryInstrumentResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  |  |
| `destination` | [string](#string) |  |  |
| `orders` | [QueryOrderResponse](#em.market.v1.QueryOrderResponse) | repeated |  |






<a name="em.market.v1.QueryInstrumentsRequest"></a>

### QueryInstrumentsRequest







<a name="em.market.v1.QueryInstrumentsResponse"></a>

### QueryInstrumentsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `instruments` | [QueryInstrumentsResponse.Element](#em.market.v1.QueryInstrumentsResponse.Element) | repeated |  |






<a name="em.market.v1.QueryInstrumentsResponse.Element"></a>

### QueryInstrumentsResponse.Element



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  |  |
| `destination` | [string](#string) |  |  |
| `last_price` | [string](#string) |  |  |
| `best_price` | [string](#string) |  |  |
| `last_traded` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="em.market.v1.QueryOrderResponse"></a>

### QueryOrderResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  |  |
| `owner` | [string](#string) |  |  |
| `source_remaining` | [string](#string) |  |  |
| `client_order_id` | [string](#string) |  |  |
| `price` | [string](#string) |  |  |
| `created` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.market.v1.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `ByAccount` | [QueryByAccountRequest](#em.market.v1.QueryByAccountRequest) | [QueryByAccountResponse](#em.market.v1.QueryByAccountResponse) |  | GET|/e-money/market/v1/account/{address}|
| `Instruments` | [QueryInstrumentsRequest](#em.market.v1.QueryInstrumentsRequest) | [QueryInstrumentsResponse](#em.market.v1.QueryInstrumentsResponse) |  | GET|/e-money/market/v1/instruments|
| `Instrument` | [QueryInstrumentRequest](#em.market.v1.QueryInstrumentRequest) | [QueryInstrumentResponse](#em.market.v1.QueryInstrumentResponse) |  | GET|/e-money/market/v1/instrument/{source}/{destination}|

 <!-- end services -->



<a name="em/market/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/market/v1/tx.proto



<a name="em.market.v1.MsgAddLimitOrder"></a>

### MsgAddLimitOrder



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `client_order_id` | [string](#string) |  |  |
| `time_in_force` | [TimeInForce](#em.market.v1.TimeInForce) |  |  |
| `source` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `destination` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="em.market.v1.MsgAddLimitOrderResponse"></a>

### MsgAddLimitOrderResponse







<a name="em.market.v1.MsgAddMarketOrder"></a>

### MsgAddMarketOrder



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `client_order_id` | [string](#string) |  |  |
| `time_in_force` | [TimeInForce](#em.market.v1.TimeInForce) |  |  |
| `source` | [string](#string) |  |  |
| `destination` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `maximum_slippage` | [string](#string) |  |  |






<a name="em.market.v1.MsgAddMarketOrderResponse"></a>

### MsgAddMarketOrderResponse







<a name="em.market.v1.MsgCancelOrder"></a>

### MsgCancelOrder



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `client_order_id` | [string](#string) |  |  |






<a name="em.market.v1.MsgCancelOrderResponse"></a>

### MsgCancelOrderResponse







<a name="em.market.v1.MsgCancelReplaceLimitOrder"></a>

### MsgCancelReplaceLimitOrder



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `original_client_order_id` | [string](#string) |  |  |
| `new_client_order_id` | [string](#string) |  |  |
| `time_in_force` | [TimeInForce](#em.market.v1.TimeInForce) |  |  |
| `source` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `destination` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="em.market.v1.MsgCancelReplaceLimitOrderResponse"></a>

### MsgCancelReplaceLimitOrderResponse







<a name="em.market.v1.MsgCancelReplaceMarketOrder"></a>

### MsgCancelReplaceMarketOrder



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `original_client_order_id` | [string](#string) |  |  |
| `new_client_order_id` | [string](#string) |  |  |
| `time_in_force` | [TimeInForce](#em.market.v1.TimeInForce) |  |  |
| `source` | [string](#string) |  |  |
| `destination` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `maximum_slippage` | [string](#string) |  |  |






<a name="em.market.v1.MsgCancelReplaceMarketOrderResponse"></a>

### MsgCancelReplaceMarketOrderResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.market.v1.Msg"></a>

### Msg


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `AddLimitOrder` | [MsgAddLimitOrder](#em.market.v1.MsgAddLimitOrder) | [MsgAddLimitOrderResponse](#em.market.v1.MsgAddLimitOrderResponse) |  | |
| `AddMarketOrder` | [MsgAddMarketOrder](#em.market.v1.MsgAddMarketOrder) | [MsgAddMarketOrderResponse](#em.market.v1.MsgAddMarketOrderResponse) |  | |
| `CancelOrder` | [MsgCancelOrder](#em.market.v1.MsgCancelOrder) | [MsgCancelOrderResponse](#em.market.v1.MsgCancelOrderResponse) |  | |
| `CancelReplaceLimitOrder` | [MsgCancelReplaceLimitOrder](#em.market.v1.MsgCancelReplaceLimitOrder) | [MsgCancelReplaceLimitOrderResponse](#em.market.v1.MsgCancelReplaceLimitOrderResponse) |  | |
| `CancelReplaceMarketOrder` | [MsgCancelReplaceMarketOrder](#em.market.v1.MsgCancelReplaceMarketOrder) | [MsgCancelReplaceMarketOrderResponse](#em.market.v1.MsgCancelReplaceMarketOrderResponse) |  | |

 <!-- end services -->



<a name="em/queries/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## em/queries/v1/query.proto



<a name="em.queries.v1.MissedBlocksInfo"></a>

### MissedBlocksInfo
ValidatorSigningInfo defines a validator's missed blocks info.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cons_address` | [string](#string) |  |  |
| `missed_blocks_counter` | [int64](#int64) |  | missed blocks counter (to avoid scanning the array every time) |
| `total_blocks_counter` | [int64](#int64) |  |  |






<a name="em.queries.v1.QueryCirculatingRequest"></a>

### QueryCirculatingRequest







<a name="em.queries.v1.QueryCirculatingResponse"></a>

### QueryCirculatingResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `total` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="em.queries.v1.QueryMissedBlocksRequest"></a>

### QueryMissedBlocksRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cons_address` | [string](#string) |  | cons_address is the address to query the missed blocks signing info |






<a name="em.queries.v1.QueryMissedBlocksResponse"></a>

### QueryMissedBlocksResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `missed_blocks_info` | [MissedBlocksInfo](#em.queries.v1.MissedBlocksInfo) |  | val_signing_info is the signing info of requested val cons address |






<a name="em.queries.v1.QuerySpendableRequest"></a>

### QuerySpendableRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="em.queries.v1.QuerySpendableResponse"></a>

### QuerySpendableResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `balance` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="em.queries.v1.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Circulating` | [QueryCirculatingRequest](#em.queries.v1.QueryCirculatingRequest) | [QueryCirculatingResponse](#em.queries.v1.QueryCirculatingResponse) |  | GET|/e-money/bank/v1/circulating|
| `MissedBlocks` | [QueryMissedBlocksRequest](#em.queries.v1.QueryMissedBlocksRequest) | [QueryMissedBlocksResponse](#em.queries.v1.QueryMissedBlocksResponse) |  | GET|/e-money/slashing/v1/missedblocks/{cons_address}|
| `Spendable` | [QuerySpendableRequest](#em.queries.v1.QuerySpendableRequest) | [QuerySpendableResponse](#em.queries.v1.QuerySpendableResponse) |  | GET|/e-money/bank/v1/spendable/{address}|

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

