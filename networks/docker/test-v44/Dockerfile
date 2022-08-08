FROM golang:1.17-buster

ARG version
ARG branch

# checkout and build linux binary
WORKDIR /go/src
RUN     git clone --depth 1 --single-branch --branch $branch https://github.com/e-money/em-ledger.git
WORKDIR em-ledger
# RUN     sed -i '16 i\    ibcconnectiontypes "github.com/cosmos/ibc-go/modules/core/03-connection/types"' app.go
# RUN     sed -i '16 i\    store "github.com/cosmos/cosmos-sdk/store/types"' app.go
# RUN     sed -i -e '/upgradekeeper.NewKeeper/r networks/upg/upgv44func.txt' app.go
RUN     git tag $version
RUN    FAST_CONSENSUS=true BIN_PREFIX=-linux LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 make build

# build MacOS binary
RUN FAST_CONSENSUS=true LEDGER_ENABLED=false GOOS=darwin GOARCH=amd64 make build
