#!/bin/bash

set -e

# Copied from https://github.com/informalsystems/ibc-rs/blob/master/scripts/ and modified to initialize a Gaia chain.

GAIA_ID="gaia"
GAIA_DATA="../build/gdata"
GAIA_RPC_PORT=26557
GAIA_GRPC_PORT=9091
GAIA_SAMOLEANS=100000000000
GAIA_P2P_PORT=26556
GAIA_PROF_PORT=6061
EMONEY_HOME=".."
EMONEY_DC_YML="$EMONEY_HOME/docker-compose.yml"

# Ensure user understands what will be deleted
if [[ -d $GAIA_DATA ]] && [[ ! "$3" == "skip" ]]; then
  echo "WARNING: $0 will DELETE the '$GAIA_DATA' folder."
  read -p "> Do you wish to continue? (y/n): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      exit 1
  fi
fi

# Ensure emoney local testnet is running
if ! [ -f "$EMONEY_DC_YML" ]; then
  echo "Error: emd is not running. Try running 'cd ../; REUSE=1 make local-testnet'" >&2
  exit 1
fi

# Ensure gaiad is installed
if ! [ -x "$(which gaiad)" ]; then
  echo "Error: gaiad is not installed. Try running 'make build-gaia'" >&2
  exit 1
fi

# Display software version
echo "GAIA VERSION INFO: $(gaiad version --log_level info)"

# Delete data from old runs
echo "Deleting $GAIA_DATA folder..."
rm -rf "$GAIA_DATA"

# Stop existing e-money gaiad processes
killall gaiad &> /dev/null || true

echo "Generating gaia configurations..."
mkdir -p "$GAIA_DATA"

# e-money node 0 ports
#      - "26656-26657:26656-26657"
#      - "1317:1317" # rest legacy
#      - "9090:9090" # grpc query

# e-money "$ONE_CHAIN" gaiad "$CHAIN_0_ID" ./data $GAIA_RPC_PORT 26656 6060 $GAIA_GRPC_PORT $GAIA_SAMOLEANS
# gaia
./one-chain gaiad $GAIA_ID $GAIA_DATA $GAIA_RPC_PORT $GAIA_P2P_PORT $GAIA_PROF_PORT $GAIA_GRPC_PORT $GAIA_SAMOLEANS

# Check platform
platform='unknown'
unamestr=`uname`
if [ "$unamestr" = 'Linux' ]; then
   platform='linux'
fi

# Set proper defaults and change ports (use a different sed for Mac or Linux)
echo "Stopping the e-money testnet to make IBC appropriate genesis edits..."
set -x
docker-compose -f $EMONEY_DC_YML stop || true
for (( i = 0; i < 4; i++ )); do
  if [ $platform = 'linux' ]; then
    sed -i 's#"172800s"#"200s"#g' $EMONEY_HOME/build/node$i/config/genesis.json
    sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $EMONEY_HOME/build/node$i/config/config.toml
    sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $EMONEY_HOME/build/node$i/config/config.toml
    sed -i 's/index_all_keys = false/index_all_keys = true/g' $EMONEY_HOME/build/node$i/config/config.toml
  else
    sed -i '' 's#"172800s"#"200s"#g' $EMONEY_HOME/build/node$i/config/genesis.json
    sed -i '' 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $EMONEY_HOME/build/node$i/config/config.toml
    sed -i '' 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $EMONEY_HOME/build/node$i/config/config.toml
    sed -i '' 's/index_all_keys = false/index_all_keys = true/g' $EMONEY_HOME/build/node$i/config/config.toml
  fi
done
docker-compose -f $EMONEY_DC_YML start
set +x

./keys-2-hermes.sh ./emoney-config.toml localnet_reuse gaia
