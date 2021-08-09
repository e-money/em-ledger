#!/bin/bash

set -xe

GM="./gm"
GAIA_DATA="./gaia"
EMONEY_HOME="../.."
EMONEY_LOG="$EMONEY_HOME/build/node0/emd.log"

if ! [ -f "./hermes" ] && ! [ -x "$(which hermes)" ]; then
  echo "Error: hermes binary is not installed. Download it from https://github.com/informalsystems/ibc-rs/releases" >&2
  exit 1
fi

# copy hermes from the path
if ! [ -f "./hermes" ]; then
  cp "$(which hermes)" ./
fi

# Ensure gaiad is installed
if ! [ -x "$(which gaiad)" ]; then
  echo "Error: gaiad is not installed. Install v5.0.5+ or clone github.com/cosmos/gaia and 'make build-gaia'" >&2
  exit 1
fi

# Display software version
echo "Requiring v5.0.5+, checking..."
echo "GAIA VERSION INFO: $(gaiad version --log_level info)"

# Ensure user understands what will be deleted
if [[ -d $GAIA_DATA ]] && [[ ! "$3" == "skip" ]]; then
  echo "WARNING: $0 will DELETE the '$GAIA_DATA' folder."
  read -p "> Do you wish to continue? (y/n): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      exit 1
  fi
fi

# Nuclear Reset of Gaia
killall -q gaiad || echo > /dev/null 2>&1
rm -rf $GAIA_DATA

$GM reset

# Ensure emoney local testnet is running
if ! [ -f "$EMONEY_LOG" ]; then
  echo "Error: emd is not running. Try running 'cd ../../; REUSE=1 make local-testnet'" >&2
  exit 1
fi

# start gaia
$GM start
$GM hermes keys

# e-money node 0 ports
#      - "26656-26657:26656-26657"
#      - "1317:1317" # rest legacy
#      - "9090:9090" # grpc query

# make emoney testnet IBC adjustments
./e-money.sh

./emkey-2-hermes.sh ./emoney-config.toml localnet_reuse

./ibc-relay.sh

$GM stop

echo "The relay test completed successfully!"