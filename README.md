![e-Money wordmark](docs/e-money%20wordmark.svg)

# Introduction

The e-Money Ledger, a proof-of-stake blockchain based on the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) and [Tendermint](https://github.com/tendermint/tendermint), introduces our novel interest-bearing, currency-backed tokens into the [Cosmos Network](https://cosmos.network).

## Quickstart Instructions

This will get you a fully synced node, very quickly.

```bash
git clone https://github.com/e-money/em-ledger.git
cd em-ledger
git checkout v1.1.4
make install
emd init choose-a-cool-name
wget -O ~/.emd/config/genesis.json https://github.com/e-money/networks/raw/master/emoney-3/genesis.json
emd start --p2p.seeds 336cdb655ea16413a8337e730683ddc0a24af9de@18.198.26.2:28656,708e559271d4d75d7ea2c3842e87d2e71a465684@3.71.36.90:28656,ecec8933d80da5fccda6bdd72befe7e064279fc1@207.180.213.123:26676,0ad7bc7687112e212bac404670aa24cd6116d097@50.18.83.75:26656,1723e34f45f54584f44d193ce9fd9c65271ca0b3@13.124.62.83:26656
```



## Getting Started

To better understand em-ledger, start with a [quick tour](docs/emd.md) of the `emd` command line interface.

The [emoneyjs library](https://github.com/e-money/emoneyjs) is the recommended way for client applications to interact with em-ledger.

_Please notice that it is highly recommended to use a [Ledger Device](docs/ledger.md) to securely manage keys._

## Networks

See [https://github.com/e-money/networks](https://github.com/e-money/networks) for instructions on how to join our production and test networks.

## Integration Guide

Tokens: [docs/tokens.md](docs/tokens.md)  
Market Module: [x/market/spec/README.md](x/market/spec/README.md)  

## Stay Updated

Website: [https://e-money.com](https://e-money.com)  
Twitter: [https://twitter.com/emoney_com](https://twitter.com/emoney_com)  
Telegram: [https://t.me/emoney_com](https://t.me/emoney_com)  
