## Build
FROM golang:1.10 AS build

ENV workdir /go/src/github.com/hatotaka/nasne-exporter

ADD . ${workdir}
WORKDIR ${workdir}

RUN GOARCH=arm \
  go build \
  -o /tmp/nasne-exporter-arm \
  github.com/hatotaka/nasne-exporter/cmd/nasne-exporter 

## Run
FROM scratch

EXPOSE 8080
COPY --from=build /tmp/nasne-exporter-arm /opt/bin/

ENTRYPOINT ["/opt/bin/nasne-exporter-arm"]
