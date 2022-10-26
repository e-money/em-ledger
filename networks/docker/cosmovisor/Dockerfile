# This software is Copyright (c) 2019-2021 e-Money A/S. It is not offered under an open source license.
#
# Please contact partners@e-money.com for licensing related questions.

FROM golang:1.19-bullseye AS build-env

RUN git clone --branch tools/cosmovisor/v1.4.0 --depth 1 https://github.com/cosmos/cosmos-sdk.git
WORKDIR cosmos-sdk/tools/cosmovisor
RUN make cosmovisor
RUN cp ./cosmovisor /go/bin/cosmovisor
