FROM alpine:latest
LABEL maintainer="quiet-listener"

ARG ARCH="amd64"
ARG OS="linux"
RUN apk update && apk add go git
COPY ./ /root/go/src/github.com/quiet-listener/latency_exporter
WORKDIR /root/go/src/github.com/quiet-listener/latency_exporter
RUN go get -d ./ && go build && go install && cp /root/go/bin/latency_exporter /bin/
RUN apk del go git 
ENV GIN_MODE="release"
ENTRYPOINT ["/bin/latency_exporter"]
EXPOSE     9101