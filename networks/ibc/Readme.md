## Test IBC gaia <-> e-money local testnets

### Components installation

#### Gaia v5.0.2+
 
Install `gaiad` and place in path

``` shell
git clone github.com/cosmos/gaia
git checkout v5.0.2 
make build-gaia
```

#### Hermes v6.0+
Please follow the instructions to install it
https://hermes.informal.systems/installation.html

Please copy the binary `hermes` in github.com/e-money/em-ledger/networks/ibc

There is a `.gitignore` for it. 

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
1. Launches gaia with 4- full-node + 1 validator-node
2. Imports the testnets keys into the hermes configuration
3. Establishes the channel, port, connection handshakes
4. Relays and Refunds tokens bi-directionally in both chains.

### Side note gm script

The gm script is a bootstrapping script for launching IBC compatible gaia n-node testnets.
Being versatile it allows launching multiple gaia testnets.
It offers start, stop, reset functionality similarly to docker-compose.

### 3rd IBC testnet 
Within the peer ../emibctokens folder there is a Stargate IBC compatible chain to test a 3rd hop if needed. It is a Starport created chain from scratch, however parallel to its IBC module there is a POC swap module in progress. In any case its tokens relay functions are operable. Please look into its **../emibctokens/readme.md** for details to launch it. 