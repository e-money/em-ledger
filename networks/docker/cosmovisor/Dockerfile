# syntax=docker/dockerfile:1
FROM golang:1.19-bullseye AS build-env
RUN git clone --branch tools/cosmovisor/v1.4.0 --depth 1 https://github.com/cosmos/cosmos-sdk.git
WORKDIR cosmos-sdk/tools/cosmovisor
RUN make cosmovisor

FROM ubuntu:18.04
COPY --from=build-env /go/cosmos-sdk/tools/cosmovisor/cosmovisor /go/bin/cosmovisor
