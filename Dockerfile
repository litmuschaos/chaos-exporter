#FROM ubuntu:16.04
#
#ARG TARGETARCH
#ENV EXPORTER=/exporter-${TARGETARCH}
#
#RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/cache/apk/*
#
#COPY .$EXPORTER /
#
#EXPOSE 8080
#
#CMD ["sh", "-c","$EXPORTER"]
#

# Multi-stage docker build
# Build stage
FROM golang:1.14 AS builder

LABEL maintainer="LitmusChaos"

# LD_FLAGS is passed as argument from Makefile. It will be empty, if no argument passed
ARG TARGETPLATFORM

ADD . /chaos-exporter
WORKDIR /chaos-exporter

RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)

RUN go env

RUN CGO_ENABLED=0 go build -o /output/chaos-exporter -v ./cmd/exporter/

RUN useradd -u 10001 chaos-exporter

# Packaging stage
FROM scratch

LABEL maintainer="LitmusChaos"

COPY --from=builder /output/chaos-exporter /
COPY --from=builder /etc/passwd /etc/passwd

USER kyverno

EXPOSE 8080

ENTRYPOINT ["./chaos-exporter"]
