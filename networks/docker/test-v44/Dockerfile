FROM golang:1.17-buster

ARG version
ARG branch

# checkout and build linux binary
RUN cd /go/src && \
    git clone --depth 1 --single-branch --branch $branch https://github.com/e-money/em-ledger.git && \
    cd em-ledger && \
    sed -i '16 i\    ibcconnectiontypes "github.com/cosmos/ibc-go/v2/modules/core/03-connection/types"' app.go && \
    sed -i '16 i\    store "github.com/cosmos/cosmos-sdk/store/types"' app.go && \
    sed -i -e '/upgradekeeper.NewKeeper/r networks/upg/upgv44func.txt' app.go && \
    git tag $version && \
    FAST_CONSENSUS=true BIN_PREFIX=-linux LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 make build

# build MacOS binary

WORKDIR /go/src/em-ledger
RUN FAST_CONSENSUS=true LEDGER_ENABLED=false GOOS=darwin GOARCH=amd64 make build
