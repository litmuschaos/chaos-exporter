# Multi-stage docker build
# Build stage
FROM golang:1.14 AS builder

LABEL maintainer="LitmusChaos"

ARG TARGETPLATFORM

ADD . /chaos-exporter
WORKDIR /chaos-exporter

RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)

RUN go env

RUN CGO_ENABLED=0 go build -o /output/chaos-exporter -v ./cmd/exporter/

# Packaging stage
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.4

LABEL maintainer="LitmusChaos"

COPY --from=builder /output/chaos-exporter /

ENV USER_UID=1001

ENTRYPOINT ["./chaos-exporter"]

USER ${USER_UID}

EXPOSE 8080
