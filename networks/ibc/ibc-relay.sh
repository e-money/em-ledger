#!/bin/bash

set -xe

GAIA_ID="gaia"
EMONEY_CHAIN_ID="localnet_reuse"

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

  hermes -c ./emoney-config.toml tx raw conn-try "$dst" "$src" $CLIENT_ID $CLIENT_ID -s $CONNECTION_ID
}

function chan_init() {
  local dst="$1"
  local src="$2"

  if ! hermes -c ./emoney-config.toml query channel end "$dst" transfer channel-0 ; then
    hermes -c ./emoney-config.toml tx raw chan-open-init "$dst" "$src" $CONNECTION_ID transfer transfer -o UNORDERED
  fi
}

#----------------- channel creation
client_create $EMONEY_CHAIN_ID $GAIA_ID
client_create $GAIA_ID $EMONEY_CHAIN_ID

#----------------- channel update
client_update $EMONEY_CHAIN_ID
client_update $GAIA_ID

#----------------- connection init
conn_init $EMONEY_CHAIN_ID $GAIA_ID

#----------------- connection try
hermes -c ./emoney-config.toml tx raw conn-try $GAIA_ID $EMONEY_CHAIN_ID $CLIENT_ID $CLIENT_ID -s $CONNECTION_ID

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

#----------------- token relays

# Gaia sends samoleans to e-Money
hermes -c ./emoney-config.toml tx raw ft-transfer $EMONEY_CHAIN_ID $GAIA_ID transfer channel-0 5000 -o 1000 -n 1 -d samoleans -r emoney1gjudpa2cmwd27cjzespu2khrvy2ukje6zfevk5

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

# e-Money sends half of received samoleans back to Gaia
hermes -c ./emoney-config.toml tx raw ft-transfer $GAIA_ID $EMONEY_CHAIN_ID transfer channel-0 2500 -o 1000 -n 1 -d ibc/27A6394C3F9FF9C9DCF5DFFADF9BB5FE9A37C7E92B006199894CF1824DF9AC7C -r cosmos1n5ggspeff4fxc87dvmg0ematr3qzw5l4rf4063
hermes -c ./emoney-config.toml tx raw packet-recv $GAIA_ID $EMONEY_CHAIN_ID transfer channel-0
hermes -c ./emoney-config.toml tx raw packet-ack  $GAIA_ID $EMONEY_CHAIN_ID transfer channel-0

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
hermes -c ./emoney-config.toml tx raw ft-transfer $EMONEY_CHAIN_ID $GAIA_ID transfer channel-0 2000 -o 1000 -n 1 -d ibc/93FF02E702BE88DE6309464BA7ABCC9932964FA348726B04B06EEE79ECB35768 -r emoney1gjudpa2cmwd27cjzespu2khrvy2ukje6zfevk5
hermes -c ./emoney-config.toml tx raw packet-recv $EMONEY_CHAIN_ID $GAIA_ID transfer channel-0
hermes -c ./emoney-config.toml tx raw packet-ack  $GAIA_ID $EMONEY_CHAIN_ID transfer channel-0