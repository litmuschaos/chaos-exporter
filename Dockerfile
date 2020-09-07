FROM ubuntu:16.04
RUN apt-get update && apt-get install ca-certificates && rm -rf /var/cache/apk/*

COPY ./exporter /

EXPOSE 8080

CMD ["/exporter"]





