FROM golang:1.16-buster

ARG OS

RUN cd /go/src && \
    git clone https://github.com/e-money/em-ledger.git && \
    cd em-ledger && \
    git checkout test-upg-0.2.0 && \
    BIN_PREFIX=-linux LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 make build

# build native binary
WORKDIR /go/src/em-ledger
RUN LEDGER_ENABLED=false GOARCH=amd64 make build