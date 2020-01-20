# This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
#
# Please contact partners@e-money.com for licensing related questions.

FROM alpine:3.7

RUN apk update && \
    apk upgrade && \
    apk --no-cache add curl jq file

VOLUME  /emoney
WORKDIR /emoney
EXPOSE 1317
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["rest-server"]
STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh

