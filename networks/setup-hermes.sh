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
EMONEY_CHAIN_ID="localnet_reuse"
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

# start gaia
./one-chain gaiad $GAIA_ID $GAIA_DATA $GAIA_RPC_PORT $GAIA_P2P_PORT $GAIA_PROF_PORT $GAIA_GRPC_PORT $GAIA_SAMOLEANS

# Check platform
platform='unknown'
unamestr=$(uname)
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

./keys-2-hermes.sh ./emoney-config.toml localnet_reuse gaia

#----------------------- ibc primitives creation functions
# $1 : destination chain id paying for the trx
# $2 : source chain id

CLIENT_ID="07-tendermint-0"
CONNECTION_ID="connection-0"

function create_client() {
  local dst="$1"
  local src="$2"

  if ! hermes -c ./emoney-config.toml query client state "$dst" $CLIENT_ID ; then
    hermes -c ./emoney-config.toml create client "$dst" "$src"
  fi
}

# $1 : destination chain id paying for the trx
# $2 : source chain id
function init_conn() {
  local dst="$1"
  local src="$2"

  if ! hermes -c ./emoney-config.toml query connection end "$dst" $CONNECTION_ID ; then
    hermes -c ./emoney-config.toml tx raw conn-init "$dst" "$src" $CLIENT_ID $CLIENT_ID
  fi
}

# $1 : destination chain id paying for the trx
# $2 : source chain id
function try_conn() {
  local dst="$1"
  local src="$2"

  hermes -c ./emoney-config.toml tx raw conn-try "$dst" "$src" $CLIENT_ID $CLIENT_ID -s $CONNECTION_ID
}

# $1 : destination chain id paying for the trx
# $2 : source chain id
function chan_init() {
  local dst="$1"
  local src="$2"

  if ! hermes -c ./emoney-config.toml query channel end "$dst" transfer channel-0 ; then
    hermes -c ./emoney-config.toml tx raw chan-open-init "$dst" "$src" $CONNECTION_ID transfer transfer -o UNORDERED
  fi
}

#----------------- channel creation
sleep 12 # give time to the e-money chain to sync
create_client $EMONEY_CHAIN_ID $GAIA_ID
create_client $GAIA_ID $EMONEY_CHAIN_ID

#----------------- connection init
init_conn $EMONEY_CHAIN_ID $GAIA_ID

#----------------- connection try
hermes -c ./emoney-config.toml tx raw conn-try $GAIA_ID $EMONEY_CHAIN_ID $CLIENT_ID $CLIENT_ID -s $CONNECTION_ID

#----------------- connection ack
hermes -c ./emoney-config.toml tx raw conn-ack $EMONEY_CHAIN_ID $GAIA_ID $CLIENT_ID $CLIENT_ID -d $CONNECTION_ID -s $CONNECTION_ID

#----------------- connection confirm
hermes -c ./emoney-config.toml tx raw conn-confirm $GAIA_ID $EMONEY_CHAIN_ID $CLIENT_ID $CLIENT_ID -d $CONNECTION_ID -s $CONNECTION_ID

#----------------- connections query
hermes -c ./emoney-config.toml query connection end $EMONEY_CHAIN_ID $CONNECTION_ID
hermes -c ./emoney-config.toml query connection end $GAIA_ID $CONNECTION_ID

#----------------- channel handshake
hermes -c ./emoney-config.toml tx raw chan-open-try $GAIA_ID $EMONEY_CHAIN_ID $CONNECTION_ID transfer transfer -s channel-0
hermes -c ./emoney-config.toml tx raw chan-open-ack $EMONEY_CHAIN_ID $GAIA_ID $CONNECTION_ID transfer transfer -d channel-0 -s channel-0
hermes -c ./emoney-config.toml tx raw chan-open-confirm $GAIA_ID $EMONEY_CHAIN_ID $CONNECTION_ID transfer transfer -d channel-0 -s channel-0
hermes -c ./emoney-config.toml query channel end $EMONEY_CHAIN_ID transfer channel-0
hermes -c ./emoney-config.toml query channel end $GAIA_ID transfer channel-0

#----------------- token relays

# Gaia sends samoleans to e-Money
hermes -c ./emoney-config.toml tx raw ft-transfer $EMONEY_CHAIN_ID $GAIA_ID transfer channel-0 5000 -o 1000 -n 1 -d samoleans

# examine Gaia's commitment -- Seqs 1
hermes -c ./emoney-config.toml query packet commitments $GAIA_ID transfer channel-0

# view unreceived packets on e-Money -- Success 1
hermes -c ./emoney-config.toml query packet unreceived-packets $EMONEY_CHAIN_ID transfer channel-0

# send recv_packet on e-Money
hermes -c ./emoney-config.toml tx raw packet-recv $EMONEY_CHAIN_ID $GAIA_ID transfer channel-0

# view unreceived packets on gaia -- []
hermes -c ./emoney-config.toml query packet unreceived-packets $GAIA_ID transfer channel-0

# send acknowledgement to Gaia of the relay
hermes -c ./emoney-config.toml tx raw packet-ack $GAIA_ID $EMONEY_CHAIN_ID transfer channel-0

# e-Money sends received samoleans back to Gaia
hermes -c ./emoney-config.toml tx raw ft-transfer $EMONEY_CHAIN_ID $GAIA_ID transfer channel-0 5000 -o 1000 -n 1 -d ibc/27A6394C3F9FF9C9DCF5DFFADF9BB5FE9A37C7E92B006199894CF1824DF9AC7C
hermes -c ./emoney-config.toml tx raw packet-recv $EMONEY_CHAIN_ID $GAIA_ID transfer channel-0
hermes -c ./emoney-config.toml tx raw packet-ack  $EMONEY_CHAIN_ID $GAIA_ID transfer channel-0

# --------------------------------------------------
# Starting with e-Money the relay now

# e-Money sends ungm to Gaia
hermes -c ./emoney-config.toml tx raw ft-transfer $GAIA_ID $EMONEY_CHAIN_ID transfer channel-0 5000 -o 1000 -n 1 -d ungm

# examine e-Money's commitment -- Seqs 1
hermes -c ./emoney-config.toml query packet commitments $EMONEY_CHAIN_ID transfer channel-0

# view unreceived packets on gaia -- Success 1
hermes -c ./emoney-config.toml query packet unreceived-packets $GAIA_ID transfer channel-0

# send recv_packet on gaia
hermes -c ./emoney-config.toml tx raw packet-recv $GAIA_ID $EMONEY_CHAIN_ID transfer channel-0

# view unreceived packets on e-Money -- []
hermes -c ./emoney-config.toml query packet unreceived-packets $EMONEY_CHAIN_ID transfer channel-0

# send acknowledgement to e-Money of the relay
hermes -c ./emoney-config.toml tx raw packet-ack $EMONEY_CHAIN_ID $GAIA_ID transfer channel-0

# gaia sends received ungm back to e-Money
hermes -c ./emoney-config.toml tx raw ft-transfer $EMONEY_CHAIN_ID $GAIA_ID transfer channel-0 2000 -o 1000 -n 1 -d ibc/93FF02E702BE88DE6309464BA7ABCC9932964FA348726B04B06EEE79ECB35768
hermes -c ./emoney-config.toml tx raw packet-recv $EMONEY_CHAIN_ID $GAIA_ID transfer channel-0
hermes -c ./emoney-config.toml tx raw packet-ack  $GAIA_ID $EMONEY_CHAIN_ID transfer channel-0