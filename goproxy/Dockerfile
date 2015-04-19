FROM progrium/busybox

WORKDIR /opt/goproxy

RUN opkg-install curl
RUN echo insecure >> ~/.curlrc

RUN \
  curl -Lks https://bintray.com$(curl -Lk http://bintray.com/phuslu/goproxy/goproxy/_latestVersion | grep linux_amd64.tar.gz | sed -n "/href/ s/.*href=['\"]\([^'\"]*\)['\"].*/\1/gp") | gunzip | tar -xf - -C /opt/goproxy/ && \
    mv goproxy*/* /bin && \
  chmod +x ./bin/goproxy

EXPOSE 8000
EXPOSE 8001

ENV GOPROXY_HOME /opt/goproxy
ENV PATH ${PATH}:${GOPROXY_HOME}/bin

ENV WEED_MASTER 127.0.0.1:9333
ENTRYPOINT ["goproxy"]
