FROM ubuntu:16.04

ARG TARGETARCH
ENV EXPORTER=/exporter-${TARGETARCH}

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/cache/apk/*

COPY .$EXPORTER /

EXPOSE 8080

CMD ["sh", "-c","$EXPORTER"]

