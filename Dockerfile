FROM bashell/alpine-bash

MAINTAINER bobliu bobliu0909@gmail.com

RUN mkdir -p /opt/humpback-center/cache

RUN mkdir -p /opt/humpback-center/etc

COPY humpback-center /opt/humpback-center/humpback-center

COPY etc/config.yaml /opt/humpback-center/etc/config.yaml

RUN chmod +x /opt/humpback-center/humpback-center

WORKDIR /opt/humpback-center

VOLUME ["/opt/humpback-center/etc"]

VOLUME ["/opt/humpback-center/cache"]

CMD ["./humpback-center"]