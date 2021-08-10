## Test IBC gaia <-> e-money local testnets

### Components installation

#### Gaia v5.0.5+ (sdk 0.42.9 including fix for Cosmos #9800)

Install `gaiad` and place in path

``` shell
git clone https://github.com/cosmos/gaia
git checkout v5.0.5
make build
cp build/gaiad $GOBIN
```

#### Hermes v0.6.2+
Please download the latest Hermes 0.6.x binary release from the [hermes releases](https://github.com/informalsystems/ibc-rs/releases). As of this writing hermes 0.6.2 was the latest.

Copy the binary `hermes` in github.com/e-money/em-ledger/networks/ibc
There is a dependency with the `gm` script in the same folder requiring the Hermes binary to be in the same folder.

There is a `.gitignore` for the `hermes` binary.

TODO look to the possibility of reusing hermes from the path.

### Launch e-money with predictable chain-id

``` shell
cd em-ledger
docker-compose kill
make clean
REUSE=1 make local-testnet
```

### Relay bi-directionally tokens between a 4-node gaia and e-money local testnets with hermes

```shell
cd networks/ibc
./run.sh #ignore errors*
```
###*
Ignore errors that do not halt the script. `run.sh` tests the existence of IBC primitives before creating new ones and thus the errors within if test brackets. However, the script should complete in its **entirety**.

### run.sh performs these functions:
1. Launches gaia with 4 full-node + 1 validator-node local testnet
2. Imports the testnets keys into the hermes configuration
3. Establishes the channel, port, connection handshakes
4. Relays and Refunds tokens bi-directionally in both chains.

### Side note gm script

The gm script is a bootstrapping script for launching IBC compatible gaia n-node testnets.
Being versatile it allows launching multiple gaia testnets.
It offers start, stop, reset functionality similarly to docker-compose.

### 3rd IBC testnet
Within the peer ../emibctokens folder there is a Stargate IBC compatible chain to test a 3rd hop if needed. It is a Starport created chain from scratch, however parallel to its IBC module there is a POC swap module in progress. In any case its tokens relay functions are operable. Please look into its **../emibctokens/readme.md** for details to launch it.

### Test Cases
[List as a docs.google.com/spreadsheets](https://docs.google.com/spreadsheets/d/16u6TO6a-XddMoYzI1RwEyWndKAV00lbV97_A-dHrWfQ/edit?usp=sharing)

### Scripts

| name | description  |
|---|---|
| `start`  | Starts a local gaia network using `gm` and initialize an IBC connection between a local e-money net and the local gaia net. |
| `run.sh`  | Perform a test of fungible token transfers between a local e-money net and a local gaia net. |
| `gm`  |  The Gaiad Manager from Informal Systems. https://github.com/informalsystems/ibc-rs/tree/v0.6.2/scripts/gm |
| `ibc-relay.sh`  | Used by `run.sh`. Uses `hermes` to perform IBC transactions across test networks  |
| `gaia-denom-path`  | Look up human-readable denomination of an IBC asset on the gaia chain. |
| `denom-path`  |   |
| `emkey-2-hermes.sh`  |   |
| `q-gaia-client`  |   |
| `q-client`  |   |
| `e-money.sh`  | Configures the e-Money testnet with quicker block times etc. Used by `run.sh` and `start` |
| `gaia_client_upd`  |   |
| `client-upd`  |   |