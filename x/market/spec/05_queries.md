# Queries

The market module can be queried using `emcli` or the [REST interface](https://cosmos.network/rpc/) of any em-ledger node.

A public interface is exposed at https://emoney.validator.network/light/.

## Active account orders

Active orders for a given account can be queried using `https://emoney.validator.network/light/market/account/<owner>`.

Or using `emcli query market account <owner>`.

## Active instruments

All instruments with active orders can be queried using `https://emoney.validator.network/light/market/instruments`.

Or using `emcli query market instruments`.

_Note that there is no listing requirement for new instruments, so these are created on-the-fly based on new orders._

## Active orders per instrument

All orders for a given instrument can be queried using `https://emoney.validator.network/light/market/instrument/<source>/<destination>`.

Or using `emcli query market instrument <source-denom> <destination-denom>`.
