PACKAGE_NAME=github.com/hatotaka/nasne-exporter
CONTAINER_NAME=quay.io/hatotaka/nasne-exporter

.PHONY: build-local build-container
build-local:
	go build -o nasne-exporter $(PACKAGE_NAME)/cmd/nasne-exporter

build-container:
	docker build \
		-t $(CONTAINER_NAME):local \
		-f build/package/Dockerfile.amd64 \
		.
