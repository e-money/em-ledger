#!/bin/bash

set -xe

GAIA_ID="gaia"
EMONEY_CHAIN_ID="localnet_reuse"

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