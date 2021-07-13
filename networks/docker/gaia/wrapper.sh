#!/usr/bin/env sh

##
## Input parameters
##
ID=${ID:-0}
LOG=${LOG:-gaiad.log}
export LOGLEVEL=debug

export GAIA_ID="gaia"
export GAIADHOME="/gaiad/gnode${ID}"
export GAIA_DATA="${GAIADHOME}/data"
export GAIA_RPC_PORT=26557
export GAIA_GRPC_PORT=9091
export GAIA_SAMOLEANS=100000000000
export GAIA_P2P_PORT=26556
export GAIA_PROF_PORT=6061

mkdir -p "${GAIA_DATA}"

/bin/one-chain /bin/gaiad $GAIA_ID "$GAIA_DATA" $GAIA_RPC_PORT $GAIA_P2P_PORT $GAIA_PROF_PORT $GAIA_GRPC_PORT $GAIA_SAMOLEANS