## Guide for testing the upgrade module
#### Overview
This document is an interactive guide for testing the upgrade module with a single docker-less e-money node. 

Please note if you'd rather test the upgrade module within the **e-money local-testnet** run:

```shell
cd em-ledger
make build-docker
make build-all
go test -v --tags="bdd" bdd_test.go upgrade_test.go
```

Two new docker images defined in em-ledger/networks/docker:

1. `emoney/cosmovisor` which builds a linux cosmovisor binary for use within the e-money local-testnet.

2. `emoney/test-upg` which builds a test upgrade mode with a trivial upgrade module handler changing the gas-fees as part of the upgrade migration process.

Note these scripts at **em-ledger/networks/upg**:
* `README.md` (this doc)
* `initchain` initializes genesis, authority for an em-legder chain.
* `startcv` (starts emoney node with cosmovisor): `cosmovisor start --home=.emd`
* `start-full-cv` runs `initchain && startcv` for testing the upgrade process to the chain with same sdk version without running migrations.
* `upg-sched` schedule an upgrade by passing the upgrade block height 
* `upgvfunc.txt` Go snippet text inserted for same chain upgrade in app.go. Enables the bdd upgrade test. No migration run.
* `cpemd` Set up the cosmovisor upgrade file folder tree. 

### Components installation
Build the revamped Docker, Linux artifacts
```shell
make build-docker
make build-all
```
As a result these artifacts generate within the build folder
```shell
.rwxrwxr-x 1 15M usr  3 Aug 13:06 cosmovisor
.rwxrwxr-x 1 60M usr  3 Aug 13:06 emd
.rwxrwxr-x 1 61M usr  3 Aug 13:06 emd-linux
.rwxrwxr-x 1 60M usr  3 Aug 13:06 emdupg
.rwxrwxr-x 1 60M usr  3 Aug 13:06 emdupg-linux
```
`emdupg` is the *upgrade* emd binary* with the migration handler setting the gas-prices upon the chain upgrade.

**The code for the upgrade handler is checked-in the *`upgrade-emd-test`* branch.*
### Install cosmovisor

#### For interactive testing and general use
Please note, the cosmovisor binary built above is a `linux` binary for use within the *local-testnet*
```shell
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@latest

# test installation success
cosmovisor version
DAEMON_NAME is not set
```

### Launch network to upgrade
```shell
# Optionally in a separate pane or terminal tab
cd networks/upg
./startcv

cd networks/upg
# schedule upgrade at future block height 6
# if you overshoot retry with lower block height
./upg-sched 6

# wait for the upgrade
# check version, upgraded gas-prices
.emd/cosmovisor/current/bin/emd version # test-upg-0.2.0
.emd/cosmovisor/current/bin/emd q authority gas-prices
min_gas_prices:
- amount: "1.000000000000000000"
  denom: ungm

#success!
```
