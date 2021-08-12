#!/bin/sh

set -o errexit -o nounset

CHAINID=$1

if [ -z "$1" ]; then
  echo "Need to input chain id..."
  exit 1
fi

rm -rf ./config
rm -rf ./data
rm -rf ./keyring-test

# Build genesis file incl account for passed address
coins="10000000000stake,100000000000samoleans"
gaiad init $CHAINID --chain-id $CHAINID --home="/gaia" --overwrite
gaiad keys add validator --keyring-backend="test" --home="/gaia"
gaiad add-genesis-account $(gaiad keys show validator -a --keyring-backend="test" --home="/gaia") $coins --home="/gaia"

# create genesis User 
USER_SEED="then nuclear favorite advance plate glare shallow enhance replace embody list dose quick scale service sentence hover announce advance nephew phrase order useful this"
(echo "$USER_SEED"; echo "$USER_SEED") | gaiad keys add user1key --recover --keyring-backend=test --home="/gaia"
USER_ADDR=$(gaiad keys show user1key -a --keyring-backend=test --home="/gaia")
gaiad add-genesis-account "$USER_ADDR" $coins --home="/gaia"
gaiad gentx validator 5000000000stake --keyring-backend="test" --chain-id $CHAINID --home="/gaia"
gaiad collect-gentxs --home="/gaia"

# Set proper defaults and change ports
sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' /gaia/config/config.toml
sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/g' /gaia/config/config.toml
sed -i 's/index_all_keys = false/index_all_keys = true/g' /gaia/config/config.toml

# Start the gaia
gaiad start --home="/gaia" --rpc.laddr=tcp://0.0.0.0:26657 --grpc.address=0.0.0.0:9091 --pruning=nothing 2>&1