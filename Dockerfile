# Multi-stage docker build
# Build stage
FROM golang:alpine AS builder

LABEL maintainer="LitmusChaos"

ARG TARGETPLATFORM

ADD . /chaos-exporter
WORKDIR /chaos-exporter

RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)

RUN go env

RUN CGO_ENABLED=0 go build -buildvcs=false -o /output/chaos-exporter -v ./cmd/exporter/

# Packaging stage
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.4

LABEL maintainer="LitmusChaos"

ENV APP_DIR="/litmus"

COPY --from=builder /output/chaos-exporter $APP_DIR/
RUN chown 65534:0 $APP_DIR/chaos-exporter && chmod 755 $APP_DIR/chaos-exporter

WORKDIR $APP_DIR
USER 65534

CMD ["./chaos-exporter"]

EXPOSE 8080
