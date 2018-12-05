PACKAGE_NAME=github.com/hatotaka/nasne_exporter
CONTAINER_NAME=quay.io/hatotaka/nasne_exporter


.PHONY: build-local build-container clean
build-local:
	go build -o nasne_exporter $(PACKAGE_NAME)/cmd/nasne_exporter

build-container:
	docker build \
		-t $(CONTAINER_NAME):local \
		-f build/package/Dockerfile.amd64 \
		.
clean:
	rm -rf nasne_exporter
