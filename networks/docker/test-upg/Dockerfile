FROM golang:1.16-buster

ARG branch
ARG version

# checkout and build linux binary
RUN cd /go/src && \
    git clone https://github.com/e-money/em-ledger.git && \
    cd em-ledger && \
    git checkout $branch && \
   	sed -i -e '/bep3.NewKeeper/r networks/upg/upgfunc.txt' app.go && \
    git tag $version && \
    BIN_PREFIX=-linux LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 make build

# build MacOS binary
WORKDIR /go/src/em-ledger
RUN LEDGER_ENABLED=false GOOS=darwin GOARCH=amd64 make build