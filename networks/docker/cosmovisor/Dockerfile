# This software is Copyright (c) 2019-2021 e-Money A/S. It is not offered under an open source license.
#
# Please contact partners@e-money.com for licensing related questions.

FROM golang:1.16-buster AS build-env

# after the SDK upgrade to 44.x+ we should change to @latest
# Note ref https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor#cosmosvisor
# If validators would like to enable the auto-download option, and they are
# currently running an application using Cosmos SDK v0.42, they will need to use
# Cosmovisor v0.1. Later versions of Cosmovisor do not support Cosmos SDK v0.42
# or earlier if the auto-download option is enabled.
RUN go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v0.1.0