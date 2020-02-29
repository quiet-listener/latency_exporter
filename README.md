# latency_exporter
Latency exporter can be used to collect url connection latency metric from the box it's running on.


## Install
There are various ways of installing

### Docker images
Docker images are available on [dockerhub] (https://hub.docker.com/r/eyeker/latency_exporter)
```
docker run -p9101:9101 eyeker/latency_exporter --web.urls-list https://google.com,https://www.facebook.com
```


### Compiling the binary
clone the repository and build manually:
```
$ mkdir -p $GOPATH/src/github.com/quiet-listener
$ cd $GOPATH/src/github.com/quiet-listener
$ git clone https://github.com/quiet-listener/latency_exporter.git
$ cd latency_exporter
$ go mod download && go mod verify
$ go build -ldflags="-w -s" -o $GOBIN/latency_exporter
$ $GOBIN/latency_exporter --web.urls-list <urlslist>
```

### Arguments Available

| Arguments  | Description | Defaults | Requirement |
| ------------- | ------------- | ------------- | ------------- |
| web.listen-port | webserver listening Port | 9101 | Optional |
| web.telemetry-path | path at which metrics will be exposed | /metrics | Optional |
| web.urls-list | list of urls separated by delimiter | | Required |
| web.url-delimiter | delimiter for web.urls-list | , | Optional |