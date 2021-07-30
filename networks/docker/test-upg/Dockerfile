FROM golang:1.16-buster

RUN cd /go/src && \
    git clone https://github.com/e-money/em-ledger.git && \
    cd em-ledger && \
    git checkout upgrade-emd-test && \
    make



