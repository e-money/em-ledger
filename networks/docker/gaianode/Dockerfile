FROM ubuntu:18.04

ARG tag

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y install curl jq file

RUN curl -L --output /usr/bin/gaiad "https://github.com/cosmos/gaia/releases/download/$tag/gaiad-$tag-linux-amd64" && \
    chmod 777 /usr/bin/gaiad

COPY ./wrapper.sh /usr/bin/wrapper.sh

VOLUME /gaia
WORKDIR /gaia

EXPOSE 26656 26657 9091

# Run NOTE: docker run -d -p 26857:26857 wrapper.sh <chain-id>
ENTRYPOINT [ "/usr/bin/wrapper.sh" ]
CMD [ "wrapper.sh", "gaia-0" ]