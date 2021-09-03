# Multi-stage docker build
# Build stage
FROM golang:1.16 AS builder

LABEL maintainer="LitmusChaos"

ARG TARGETPLATFORM

ADD . /chaos-exporter
WORKDIR /chaos-exporter

RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)

RUN go env

RUN CGO_ENABLED=0 go build -o /output/chaos-exporter -v ./cmd/exporter/

FROM golang:alpine as cert
RUN apk --no-cache add ca-certificates

# Packaging stage
FROM scratch

LABEL maintainer="LitmusChaos"

COPY --from=cert /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /output/chaos-exporter /

USER 1001

CMD ["./chaos-exporter"]

EXPOSE 8080
