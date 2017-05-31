FROM bashell/alpine-bash

MAINTAINER bobliu bobliu0909@gmail.com

RUN mkdir -p /opt/humpback-center/cache

RUN mkdir -p /opt/humpback-center/etc

RUN mkdir -p /opt/humpback-center/notify

COPY etc/config.yaml /opt/humpback-center/etc/config.yaml

COPY notify/template.html /opt/humpback-center/notify/template.html

COPY humpback-center /opt/humpback-center/humpback-center

RUN chmod +x /opt/humpback-center/humpback-center

WORKDIR /opt/humpback-center

VOLUME ["/opt/humpback-center/etc"]

VOLUME ["/opt/humpback-center/cache"]

CMD ["./humpback-center"]

EXPOSE 8589
