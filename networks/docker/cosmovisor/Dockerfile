# This software is Copyright (c) 2019-2021 e-Money A/S. It is not offered under an open source license.
#
# Please contact partners@e-money.com for licensing related questions.

FROM golang:1.19-bullseye AS build-env

RUN git clone https://github.com/cosmos/cosmos-sdk.git
WORKDIR cosmos-sdk/cosmovisor
RUN git checkout cosmovisor/v1.2.0
RUN make cosmovisor
RUN cp ./cosmovisor /go/bin/cosmovisor
