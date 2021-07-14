#!/bin/bash

set -ex

CLIENT_ID="07-tendermint-0"
CONNECTION_ID="connection-0"
GAIA_ID="gaia"
EMONEY_CHAIN_ID="localnet_reuse"

# $1 : destination chain id paying for the trx
# $2 : source chain id
function chan_init() {
  local dst="$1"
  local src="$2"

  if ! hermes -c ./emoney-config.toml query channel end "$dst" transfer channel-0 ; then
    hermes -c ./emoney-config.toml tx raw chan-open-init "$dst" "$src" $CONNECTION_ID transfer transfer -o UNORDERED
  fi
}

#----------------- channel handshake
chan_init $EMONEY_CHAIN_ID $GAIA_ID
hermes -c ./emoney-config.toml tx raw chan-open-try $GAIA_ID $EMONEY_CHAIN_ID $CONNECTION_ID transfer transfer -s channel-0
hermes -c ./emoney-config.toml tx raw chan-open-ack $EMONEY_CHAIN_ID $GAIA_ID $CONNECTION_ID transfer transfer -d channel-0 -s channel-0
hermes -c ./emoney-config.toml tx raw chan-open-confirm $GAIA_ID $EMONEY_CHAIN_ID $CONNECTION_ID transfer transfer -d channel-0 -s channel-0
hermes -c ./emoney-config.toml query channel end $EMONEY_CHAIN_ID transfer channel-0
hermes -c ./emoney-config.toml query channel end $GAIA_ID transfer channel-0