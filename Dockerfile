FROM frolvlad/alpine-glibc:alpine-3.7

MAINTAINER bobliu bobliu0909@gmail.com

RUN apk add --no-cache bash

RUN mkdir -p /opt/humpback-center/cache

RUN mkdir -p /opt/humpback-center/etc

RUN mkdir -p /opt/humpback-center/notify

COPY etc/config.yaml /opt/humpback-center/etc/config.yaml

COPY notify/template.html /opt/humpback-center/notify/template.html

COPY humpback-center /opt/humpback-center/humpback-center

COPY dumb-init /dumb-init
RUN chmod 777 /dumb-init

ENTRYPOINT ["/dumb-init", "--"]

WORKDIR /opt/humpback-center

VOLUME ["/opt/humpback-center/etc"]

VOLUME ["/opt/humpback-center/cache"]

VOLUME ["/opt/humpback-center/logs"]

CMD ["./humpback-center"]

EXPOSE 8589
