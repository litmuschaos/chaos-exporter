FROM ubuntu:16.04
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/cache/apk/*

COPY ./exporter /

EXPOSE 8080

CMD ["/exporter"]





