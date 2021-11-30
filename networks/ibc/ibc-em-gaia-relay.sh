#!/bin/bash

set -xe

GAIA_ID="gaia"
EMONEY_CHAIN_ID="localnet_reuse"

#----------------- token relays

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