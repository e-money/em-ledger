# syntax=docker/dockerfile:1

FROM golang:1.19-bullseye as build-env
ARG branch
ARG version
# checkout and build linux binary
RUN cd /go/src && \
    git clone --depth 1 --single-branch --branch $branch https://github.com/e-money/em-ledger.git && \
    cd em-ledger && \
    sed -i -e '/upgradekeeper.NewKeeper/r networks/upg/upgfunc.txt' app.go && \
    VERSION=$version FAST_CONSENSUS=true BIN_PREFIX=-linux LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 make build

# build MacOS binary
WORKDIR /go/src/em-ledger
RUN VERSION=$version FAST_CONSENSUS=true LEDGER_ENABLED=false GOOS=darwin GOARCH=amd64 make build

FROM ubuntu:18.04
COPY --from=build-env /go/src/em-ledger/build/* /go/src/em-ledger/build/
