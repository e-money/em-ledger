FROM golang:1.16-buster

ARG OS

RUN cd /go/src && \
    git clone https://github.com/e-money/em-ledger.git && \
    cd em-ledger && \
    git checkout upgrade-emd-test && \
    make build-linux

# build native binary
WORKDIR /go/src/em-ledger
RUN GOOS='darwin' go build -tags "netgo ledger" -ldflags '-X github.com/cosmos/cosmos-sdk/version.Name=e-money -X github.com/cosmos/cosmos-sdk/version.AppName=emd -X github.com/cosmos/cosmos-sdk/version.Version=test-upg-0.0.1 -X "github.com/cosmos/cosmos-sdk/version.BuildTags=netgo,ledger"' -o build/emd ./cmd/emd