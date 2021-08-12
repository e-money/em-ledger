# This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
#
# Please contact partners@e-money.com for licensing related questions.

FROM ubuntu:18.04

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y install curl jq file

VOLUME  /emoney
WORKDIR /emoney
EXPOSE 26656 26657 1317
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start"]
STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh