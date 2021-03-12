## IBC relayer demo
Scripts to start a local ibc demo with 2 emd chains
[relayer](https://github.com/cosmos/relayer/)


The original scrips were copied from the [relayer project](https://github.com/cosmos/relayer/tree/master/scripts) and modified
to match the `emd` setup for `authority`. Many thanks to the original authors :bouquet:


### Install
Follow the install instructions for:
* `emd`
* [`rly`](https://github.com/cosmos/relayer)
* `jq` 

Tested with `rly` version [`f4e722e`](https://github.com/cosmos/relayer/tree/f4e722ebcdfa3e22f6bb928c5bac7307cbe80f20)

### Run local demo

* Start 2 chains
```shell
./two-chainz
```

* Create clients, connections and channels between both chains ICS20 `transfer` module
```shell
rly tx link demo -d
```

* Check relayer keys balances on both chains
```shell
rly q bal ibc-0
rly q bal ibc-1
```
Note that both chains have the same token denoms

* Start ICS20 transfer 
```shell
rly tx transfer ibc-0 ibc-1 100samoleans $(rly chains address ibc-1)
```
Same as `emd tx ibc-transfer`
* Relay packets/ack
```shell 
rly tx relay demo --debug
```
This command does multiple steps:
* Relay IBC packet from "ibc-0" chain to "ibc-1" 
* Relay IBC packet acknowledgemt from "ibc-1" chain to "ibc-0"


* Verify it was received as a *voucher*
```shell
rly q bal ibc-1
```
* Add a new address
```shell
PASSWORD="1234567890"
(echo "$PASSWORD"; echo "$PASSWORD") | emd keys add fred --keyring-backend=file --keyring-dir=./data/ibc-0/keyring
```

* Return the *voucher* to the original chain
```shell
DEST=$(echo "$PASSWORD" | emd keys show fred --address --keyring-backend=file --keyring-dir=./data/ibc-0/keyring)
rly tx transfer ibc-1 ibc-0 100ibc/27A6394C3F9FF9C9DCF5DFFADF9BB5FE9A37C7E92B006199894CF1824DF9AC7C "$DEST"
rly tx relay demo --debug
```

### Relayer as daemon
* Link `demo` as above
* Start relayer to listen for messages
```shell 
rly start demo-path --max-msgs 3
```

* Find channel
```shell
emd q ibc channel channels
```

* Send coins in a different window
```shell
DEST=$(rly chains address ibc-1)
emd tx ibc-transfer transfer transfer channel-0 $DEST 5samoleans --keyring-backend=file --keyring-dir=./data/ibc-0/keyring --from fred
```
