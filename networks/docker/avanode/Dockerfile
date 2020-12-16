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
CMD ["build/avalanchego", "--public-ip=127.0.0.1", "--public-ip=127.0.0.1", "--snow-quorum-size=2", "--http-port=9650", "--staking-port=9651", "--db-dir=db/node1", "--staking-enabled=true", "--network-id=local", "--bootstrap-ips=", "--staking-tls-cert-file=./staking/local/staker1.crt", "--staking-tls-key-file=./staking/local/staker1.key"]
STOPSIGNAL SIGTERM
