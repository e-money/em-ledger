# syntax=docker/dockerfile:experimental

# 1st stage compilation
FROM golang:1.15.5-buster AS build-env

RUN mkdir -p /go/src/github.com/ava-labs

WORKDIR $GOPATH/src/github.com/ava-labs/
COPY ./avalanchego avalanchego 

WORKDIR $GOPATH/src/github.com/ava-labs/avalanchego
RUN ./scripts/build.sh

# 2nd stage binary
FROM debian:buster
EXPOSE 9650 9651
WORKDIR /
COPY --from=build-env /go/src/github.com/ava-labs/avalanchego/ ./
STOPSIGNAL SIGTERM
