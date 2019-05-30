FROM ubuntu:16.04

COPY ./exporter /

EXPOSE 8080

CMD ["/exporter"]





