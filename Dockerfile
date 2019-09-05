FROM golang:alpine
RUN mkdir -p github.com/litmuschaos/chaos-exporter
ADD . /github.com/litmuschaos/chaos-exporter
WORKDIR /github.com/litmuschaos/chaos-exporter
RUN apk update && apk add git 
RUN pwd
RUN go get github.com/litmuschaos/chaos-exporter/pkg/chaosmetrics
RUN go get github.com/Sirupsen/logrus
RUN go get github.com/prometheus/client_golang/prometheus
RUN go get k8s.io/client-go/rest

