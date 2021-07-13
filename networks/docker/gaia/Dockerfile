# Changed from Alpine to Ubuntu because the keyring PR is linking to libc
# Alpine uses muslc instead of libc
FROM ubuntu:18.04

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y install curl jq file

COPY gaiad /bin/gaiad
COPY one-chain /bin/one-chain
COPY wrapper.sh /bin/wrapper.sh

VOLUME /gaiad
WORKDIR /gaiad

ENTRYPOINT ["/bin/wrapper.sh"]
CMD ["start"]
STOPSIGNAL SIGTERM

EXPOSE 26656 26657 9091 6061