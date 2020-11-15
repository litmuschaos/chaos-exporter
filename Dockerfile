FROM ubuntu:16.04

ARG TARGETARCH

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/cache/apk/*

COPY ./exporter-${TARGETARCH} /

EXPOSE 8080

CMD ["/exporter-${TARGETARCH}"]





