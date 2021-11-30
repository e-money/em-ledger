#!/bin/bash

set -xe

#----------------------- ibc primitives creation functions
CLIENT_ID="07-tendermint-0"
CONNECTION_ID="connection-0"

# $1 : destination chain id paying for the trx
# $2 : source or sender chain id
function client_create() {
  local dst="$1"
  local src="$2"

  if ! hermes -c ./emoney-config.toml query client state "$dst" $CLIENT_ID ; then
    hermes -c ./emoney-config.toml create client "$dst" "$src"
  fi
}

function client_update() {
  local dst="$1"

  if hermes -c ./emoney-config.toml query client state "$dst" $CLIENT_ID ; then
    hermes -c ./emoney-config.toml tx raw update-client "$dst" $CLIENT_ID
  fi
}

function conn_init() {
  local dst="$1"
  local src="$2"

  if ! hermes -c ./emoney-config.toml query connection end "$dst" $CONNECTION_ID ; then
    hermes -c ./emoney-config.toml tx raw conn-init "$dst" "$src" $CLIENT_ID $CLIENT_ID
  fi
}

function try_conn() {
  local dst="$1"
  local src="$2"

  # try 4 times to establish a connection
  tries=4
  sleep 7

  for (( i = 0; i < tries; i++ )); do
    echo attempt "$i" to conn-try
    if ! hermes -c ./emoney-config.toml tx raw conn-try "$dst" "$src" $CLIENT_ID $CLIENT_ID -s $CONNECTION_ID ; then
      sleep 7
      continue
    fi

    break
  done
}

function chan_init() {
  local dst="$1"
  local src="$2"

  if ! hermes -c ./emoney-config.toml query channel end "$dst" transfer channel-0 ; then
    hermes -c ./emoney-config.toml tx raw chan-open-init "$dst" "$src" $CONNECTION_ID transfer transfer -o UNORDERED
  fi
}

GAIA_ID="gaia"
EMONEY_CHAIN_ID="localnet_reuse"

#----------------------- ibc primitives creation functions
CLIENT_ID="07-tendermint-0"
CONNECTION_ID="connection-0"

#----------------- channel creation
client_create $EMONEY_CHAIN_ID $GAIA_ID
client_create $GAIA_ID $EMONEY_CHAIN_ID

#----------------- channel update
client_update $EMONEY_CHAIN_ID
client_update $GAIA_ID

#----------------- connection init
conn_init $EMONEY_CHAIN_ID $GAIA_ID

#----------------- connection try
try_conn $GAIA_ID $EMONEY_CHAIN_ID

#----------------- connection ack
hermes -c ./emoney-config.toml tx raw conn-ack $EMONEY_CHAIN_ID $GAIA_ID $CLIENT_ID $CLIENT_ID -d $CONNECTION_ID -s $CONNECTION_ID

#----------------- connection confirm
hermes -c ./emoney-config.toml tx raw conn-confirm $GAIA_ID $EMONEY_CHAIN_ID $CLIENT_ID $CLIENT_ID -d $CONNECTION_ID -s $CONNECTION_ID

#----------------- connections query
hermes -c ./emoney-config.toml query connection end $EMONEY_CHAIN_ID $CONNECTION_ID
hermes -c ./emoney-config.toml query connection end $GAIA_ID $CONNECTION_ID

#----------------- channel port handshake
chan_init $EMONEY_CHAIN_ID $GAIA_ID
hermes -c ./emoney-config.toml tx raw chan-open-try $GAIA_ID $EMONEY_CHAIN_ID $CONNECTION_ID transfer transfer -s channel-0
hermes -c ./emoney-config.toml tx raw chan-open-ack $EMONEY_CHAIN_ID $GAIA_ID $CONNECTION_ID transfer transfer -d channel-0 -s channel-0
hermes -c ./emoney-config.toml tx raw chan-open-confirm $GAIA_ID $EMONEY_CHAIN_ID $CONNECTION_ID transfer transfer -d channel-0 -s channel-0
hermes -c ./emoney-config.toml query channel end $EMONEY_CHAIN_ID transfer channel-0
hermes -c ./emoney-config.toml query channel end $GAIA_ID transfer channel-0

echo "the IBC chains have initialized awaiting IBC transactions!"
echo "Hermes relayer can be started with './hermes -c ./emoney-config.toml start'"
