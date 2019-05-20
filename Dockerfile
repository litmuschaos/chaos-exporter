FROM ubuntu:16.04

COPY cmd/exporter/exporter /

EXPOSE 8080

CMD ["/exporter"]





