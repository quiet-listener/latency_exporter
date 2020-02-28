FROM golang:alpine AS build_base
RUN apk add --no-cache git ca-certificates && update-ca-certificates
ENV USER=promuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"
WORKDIR /tmp/latency_exporter
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go mod verify
COPY . .
ARG GOOS=linux
ARG GOARCH=amd64
RUN go build -ldflags="-w -s" -o ./out/latency_exporter

FROM scratch
LABEL maintainer="quiet-listener"
COPY --from=build_base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build_base /etc/passwd /etc/passwd
COPY --from=build_base /etc/group /etc/group
COPY --from=build_base  /tmp/latency_exporter/out/latency_exporter /bin/latency_exporter
COPY --from=build_base  /tmp/latency_exporter/templates  /tmp/latency_exporter/templates
COPY  --from=build_base  /tmp/latency_exporter/statics /tmp/latency_exporter/statics
WORKDIR /tmp/latency_exporter
ENV GIN_MODE="release"
USER promuser:promuser
ENTRYPOINT ["/bin/latency_exporter"]
EXPOSE   9101