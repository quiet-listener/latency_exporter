FROM alpine:latest
LABEL maintainer="quiet-listener"
RUN apk add --update ca-certificates
ADD $GOBIN/latency_exporter  /bin/
COPY ./templates /tmp/templates
COPY ./statics /tmp/statics
WORKDIR /tmp/
ENV GIN_MODE="release"
ENTRYPOINT ["/bin/latency_exporter"]
EXPOSE     9101