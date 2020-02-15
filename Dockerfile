ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox:latest
LABEL maintainer="quiet-listener"

ARG ARCH="amd64"
ARG OS="linux"
ARG url-list
COPY ${GOBIN}/latency_exporter /bin/latency_exporter
ENV GIN_MODE="release"
RUN test -n "$url-list"

ENTRYPOINT ["/bin/latency_exporter","--web.urls-list",url-list]
EXPOSE     9101