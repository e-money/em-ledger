FROM golang:1.16-buster AS build-env

ARG tag

# Set working directory for the build
RUN mkdir -p /go/src/github.com/cosmos

# shallow checkout and build linux binary
RUN cd /go/src/github.com/cosmos && \
    git clone --depth 1 --branch $tag https://github.com/cosmos/gaia.git && \
    cd gaia && \
    LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 make install

FROM ubuntu:18.04

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y install curl jq file

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/gaiad /usr/bin/gaiad

COPY ./wrapper.sh /usr/bin/wrapper.sh

VOLUME /gaia
WORKDIR /gaia

EXPOSE 26656 26657 9091

# Run NOTE: docker run -d -p 26857:26857 wrapper.sh <chain-id>
ENTRYPOINT [ "/usr/bin/wrapper.sh" ]
CMD [ "wrapper.sh", "gaia-0" ]
